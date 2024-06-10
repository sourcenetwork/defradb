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

	// ID returns the unique immutable identifier for this transaction.
	ID() uint64

	// Commit finalizes a transaction, attempting to commit it to the Datastore.
	// May return an error if the transaction has gone stale. The presence of an
	// error is an indication that the data was not committed to the Datastore.
	Commit(ctx context.Context) error
	// Discard throws away changes recorded in a transaction without committing
	// them to the underlying Datastore. Any calls made to Discard after Commit
	// has been successfully called will have no effect on the transaction and
	// state of the Datastore, making it safe to defer.
	Discard(ctx context.Context)

	// OnSuccess registers a function to be called when the transaction is committed.
	OnSuccess(fn func())

	// OnError registers a function to be called when the transaction is rolled back.
	OnError(fn func())

	// OnDiscard registers a function to be called when the transaction is discarded.
	OnDiscard(fn func())

	// OnSuccessAsync registers a function to be called asynchronously when the transaction is committed.
	OnSuccessAsync(fn func())

	// OnErrorAsync registers a function to be called asynchronously when the transaction is rolled back.
	OnErrorAsync(fn func())

	// OnDiscardAsync registers a function to be called asynchronously when the transaction is discarded.
	OnDiscardAsync(fn func())
}

type txn struct {
	MultiStore
	t  ds.Txn
	id uint64

	successFns []func()
	errorFns   []func()
	discardFns []func()

	successAsyncFns []func()
	errorAsyncFns   []func()
	discardAsyncFns []func()
}

var _ Txn = (*txn)(nil)

func newTxnFrom(ctx context.Context, rootstore ds.TxnDatastore, readonly bool) (ds.Txn, error) {
	// check if our datastore natively supports iterable transaction, transactions or batching
	switch t := rootstore.(type) {
	case iterable.IterableTxnDatastore:
		return t.NewIterableTransaction(ctx, readonly)

	default:
		return rootstore.NewTransaction(ctx, readonly)
	}
}

// NewTxnFrom returns a new Txn from the rootstore.
func NewTxnFrom(ctx context.Context, rootstore ds.TxnDatastore, id uint64, readonly bool) (Txn, error) {
	rootTxn, err := newTxnFrom(ctx, rootstore, readonly)
	if err != nil {
		return nil, err
	}
	multistore := MultiStoreFrom(ShimTxnStore{rootTxn})
	return &txn{
		t:          rootTxn,
		MultiStore: multistore,
		id:         id,
	}, nil
}

func (t *txn) ID() uint64 {
	return t.id
}

func (t *txn) Commit(ctx context.Context) error {
	var fns []func()
	var asyncFns []func()

	err := t.t.Commit(ctx)
	if err != nil {
		fns = t.errorFns
		asyncFns = t.errorAsyncFns
	} else {
		fns = t.successFns
		asyncFns = t.successAsyncFns
	}

	for _, fn := range asyncFns {
		go fn()
	}
	for _, fn := range fns {
		fn()
	}
	return err
}

func (t *txn) Discard(ctx context.Context) {
	t.t.Discard(ctx)
	for _, fn := range t.discardAsyncFns {
		go fn()
	}
	for _, fn := range t.discardFns {
		fn()
	}
}

func (t *txn) OnSuccess(fn func()) {
	t.successFns = append(t.successFns, fn)
}

func (t *txn) OnError(fn func()) {
	t.errorFns = append(t.errorFns, fn)
}

func (t *txn) OnDiscard(fn func()) {
	t.discardFns = append(t.discardFns, fn)
}

func (t *txn) OnSuccessAsync(fn func()) {
	t.successAsyncFns = append(t.successAsyncFns, fn)
}

func (t *txn) OnErrorAsync(fn func()) {
	t.errorAsyncFns = append(t.errorAsyncFns, fn)
}

func (t *txn) OnDiscardAsync(fn func()) {
	t.discardAsyncFns = append(t.discardAsyncFns, fn)
}

// Shim to make ds.Txn support ds.Datastore.
type ShimTxnStore struct {
	ds.Txn
}

// Sync executes the transaction.
func (ts ShimTxnStore) Sync(ctx context.Context, prefix ds.Key) error {
	return ts.Txn.Commit(ctx)
}

// Close discards the transaction.
func (ts ShimTxnStore) Close() error {
	ts.Discard(context.TODO())
	return nil
}
