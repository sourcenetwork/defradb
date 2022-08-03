// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package datastore

import (
	"context"

	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/datastore/iterable"
)

// Txn is a common interface to the db.Txn struct.
type Txn interface {
	MultiStore

	IsBatch() bool

	// Commit finalizes a transaction, attempting to commit it to the Datastore.
	// May return an error if the transaction has gone stale. The presence of an
	// error is an indication that the data was not committed to the Datastore.
	Commit(ctx context.Context) error
	// Discard throws away changes recorded in a transaction without committing
	// them to the underlying Datastore. Any calls made to Discard after Commit
	// has been successfully called will have no effect on the transaction and
	// state of the Datastore, making it safe to defer.
	Discard(ctx context.Context)

	OnSuccess(fn func())
	OnError(fn func())
}

type txn struct {
	t ds.Txn
	MultiStore
	isBatch bool

	successFns []func()
	errorFns   []func()
}

var _ Txn = (*txn)(nil)

func NewTxnFrom(ctx context.Context, rootstore ds.Batching, readonly bool) (Txn, error) {
	// check if our datastore natively supports iterable transaction, transactions or batching
	if iterableTxnStore, ok := rootstore.(iterable.IterableTxnDatastore); ok {
		rootTxn, err := iterableTxnStore.NewIterableTransaction(ctx, readonly)
		if err != nil {
			return nil, err
		}
		multistore := MultiStoreFrom(rootTxn)
		return &txn{
			rootTxn,
			multistore,
			false,
			[]func(){},
			[]func(){},
		}, nil
	}

	var rootTxn ds.Txn
	var err error
	var isBatch bool
	if txnStore, ok := rootstore.(ds.TxnDatastore); ok {
		rootTxn, err = txnStore.NewTransaction(ctx, readonly)
		if err != nil {
			return nil, err
		}
	} else {
		batcher, err := rootstore.Batch(ctx)
		if err != nil {
			return nil, err
		}

		// hide a ds.Batching store as a ds.Txn
		rootTxn = ShimBatcherTxn{
			Read:  rootstore,
			Batch: batcher,
		}
		isBatch = true
	}

	root := AsDSReaderWriter(ShimTxnStore{rootTxn})
	multistore := MultiStoreFrom(root)
	return &txn{
		rootTxn,
		multistore,
		isBatch,
		[]func(){},
		[]func(){},
	}, nil
}

func (t *txn) Commit(ctx context.Context) error {
	if err := t.t.Commit(ctx); err != nil {
		t.runErrorFns(ctx)
		return err
	}
	t.runSuccessFns(ctx)
	return nil
}

func (t *txn) Discard(ctx context.Context) {
	t.t.Discard(ctx)
}

func (txn *txn) OnSuccess(fn func()) {
	if fn == nil {
		return
	}
	txn.successFns = append(txn.successFns, fn)
}

func (txn *txn) OnError(fn func()) {
	if fn == nil {
		return
	}
	txn.errorFns = append(txn.errorFns, fn)
}

func (txn *txn) runErrorFns(ctx context.Context) {
	for _, fn := range txn.errorFns {
		fn()
	}
}

func (txn *txn) runSuccessFns(ctx context.Context) {
	for _, fn := range txn.successFns {
		fn()
	}
}

/*
// Systemstore returns the txn wrapped as a systemstore under the /system namespace
func (t *txn) Systemstore() DSReaderWriter {
	return t.systemstore
}

// Datastore returns the txn wrapped as a datastore under the /data namespace
func (t *txn) Datastore() DSReaderWriter {
	return t.datastore
}

// Headstore returns the txn wrapped as a headstore under the /heads namespace
func (t *txn) Headstore() DSReaderWriter {
	return t.headstore
}

// DAGstore returns the txn wrapped as a blockstore for a DAGStore under the /blocks namespace
func (t *txn) DAGstore() DAGStore {
	return t.dagstore
}

// Rootstore returns the underlying txn as a DSReaderWriter to implement
// the MultiStore interface
func (t *txn) Rootstore() DSReaderWriter {
	return t.IterableTxn
}*/

func (txn *txn) IsBatch() bool {
	return txn.isBatch
}

// func (txn *txn) Commit(ctx context.Context) error {
// 	if err := txn.IterableTxn.Commit(ctx); err != nil {
// 		txn.runErrorFns(ctx)
// 		return err
// 	}
// 	txn.runSuccessFns(ctx)
// 	return nil
// }

// Shim to make ds.Txn support ds.Datastore
type ShimTxnStore struct {
	ds.Txn
}

func (ts ShimTxnStore) Sync(ctx context.Context, prefix ds.Key) error {
	return ts.Txn.Commit(ctx)
}

func (ts ShimTxnStore) Close() error {
	ts.Discard(context.TODO())
	return nil
}

// shim to make ds.Batch implement ds.Datastore
type ShimBatcherTxn struct {
	ds.Read
	ds.Batch
}

func (ShimBatcherTxn) Discard(_ context.Context) {
	// noop
}
