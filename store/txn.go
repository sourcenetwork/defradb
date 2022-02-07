// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package store

import (
	"context"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastores/iterable"
	"github.com/sourcenetwork/defradb/db/base"
)

type txn struct {
	iterable.IterableTxn
	isBatch bool

	// wrapped DS
	rootstore   core.DSReaderWriter
	systemstore core.DSReaderWriter // wrapped txn /system namespace
	datastore   core.DSReaderWriter // wrapped txn /data namespace
	headstore   core.DSReaderWriter // wrapped txn /heads namespace
	dagstore    core.DAGStore       // wrapped txn /blocks namespace

	// @todo once we move all Txn creation from db to here
	// successFns []func()
	// errorFns   []func()
}

func NewTxnFrom(ctx context.Context, rootstore ds.Batching, readonly bool) (core.Txn, error) {
	Txn := new(txn)

	// check if our datastore natively supports iterable transaction, transactions or batching
	if iterableTxnStore, ok := rootstore.(iterable.IterableTxnDatastore); ok {
		dstxn, err := iterableTxnStore.NewIterableTransaction(ctx, readonly)
		if err != nil {
			return nil, err
		}
		Txn.IterableTxn = dstxn
	} else if txnStore, ok := rootstore.(ds.TxnDatastore); ok {
		dstxn, err := txnStore.NewTransaction(ctx, readonly)
		if err != nil {
			return nil, err
		}
		Txn.IterableTxn = iterable.NewIterableTransaction(dstxn)
		// Note: db.rootstore now has type `ds.Batching`.
	} else {
		batcher, err := rootstore.Batch(ctx)
		if err != nil {
			return nil, err
		}

		// hide a ds.Batching store as a ds.Txn
		rb := ShimBatcherTxn{
			Read:  rootstore,
			Batch: batcher,
		}
		Txn.IterableTxn = iterable.NewIterableTransaction(rb)
		Txn.isBatch = true
	}

	// add the wrapped datastores using the existing KeyTransform functions from the db
	// @todo Check if KeyTransforms are nil beforehand

	// debug stuff... ignore
	//
	// txnid := RandStringRunes(5)
	// txn.systemstore = ds.NewLogDatastore(ktds.Wrap(shimStore, db.ssKeyTransform), fmt.Sprintf("%s:systemstore", txnid))
	// txn.datastore = ds.NewLogDatastore(ktds.Wrap(shimStore, db.dsKeyTransform), fmt.Sprintf("%s:datastore", txnid))
	// txn.headstore = ds.NewLogDatastore(ktds.Wrap(shimStore, db.hsKeyTransform), fmt.Sprintf("%s:headstore", txnid))
	// batchstore := ds.NewLogDatastore(ktds.Wrap(shimStore, db.dagKeyTransform), fmt.Sprintf("%s:dagstore", txnid))

	shimStore := ShimTxnStore{Txn.IterableTxn}
	Txn.rootstore = shimStore
	Txn.systemstore = namespace.Wrap(shimStore, base.SystemStoreKey)
	Txn.datastore = namespace.Wrap(shimStore, base.DataStoreKey)
	Txn.headstore = namespace.Wrap(shimStore, base.HeadStoreKey)
	batchstore := namespace.Wrap(shimStore, base.BlockStoreKey)

	Txn.dagstore = NewDAGStore(batchstore)

	return Txn, nil
}

// Systemstore returns the txn wrapped as a systemstore under the /system namespace
func (t *txn) Systemstore() core.DSReaderWriter {
	return t.systemstore
}

// Datastore returns the txn wrapped as a datastore under the /data namespace
func (t *txn) Datastore() core.DSReaderWriter {
	return t.datastore
}

// Headstore returns the txn wrapped as a headstore under the /heads namespace
func (t *txn) Headstore() core.DSReaderWriter {
	return t.headstore
}

// DAGstore returns the txn wrapped as a blockstore for a DAGStore under the /blocks namespace
func (t *txn) DAGstore() core.DAGStore {
	return t.dagstore
}

// Rootstore returns the underlying txn as a DSReaderWriter to implement
// the MultiStore interface
func (t *txn) Rootstore() core.DSReaderWriter {
	return t.IterableTxn
}

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
