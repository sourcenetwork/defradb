// Copyright 2024 Democratized Data Foundation
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

	"github.com/sourcenetwork/corekv"
)

// // Txn is a common interface to the db.Txn struct.
// type Txn interface {
// 	corekv.Reader
// 	corekv.Writer

// 	// ID returns the unique immutable identifier for this transaction.
// 	ID() uint64

// 	// Commit finalizes a transaction, attempting to commit it to the Datastore.
// 	// May return an error if the transaction has gone stale. The presence of an
// 	// error is an indication that the data was not committed to the Datastore.
// 	Commit(ctx context.Context) error
// 	// Discard throws away changes recorded in a transaction without committing
// 	// them to the underlying Datastore. Any calls made to Discard after Commit
// 	// has been successfully called will have no effect on the transaction and
// 	// state of the Datastore, making it safe to defer.
// 	Discard(ctx context.Context)

// 	// OnSuccess registers a function to be called when the transaction is committed.
// 	OnSuccess(fn func())

// 	// OnError registers a function to be called when the transaction is rolled back.
// 	OnError(fn func())

// 	// OnDiscard registers a function to be called when the transaction is discarded.
// 	OnDiscard(fn func())

// 	// OnSuccessAsync registers a function to be called asynchronously when the transaction is committed.
// 	OnSuccessAsync(fn func())

// 	// OnErrorAsync registers a function to be called asynchronously when the transaction is rolled back.
// 	OnErrorAsync(fn func())

// 	// OnDiscardAsync registers a function to be called asynchronously when the transaction is discarded.
// 	OnDiscardAsync(fn func())
// }

type Txn struct {
	*Multistore

	txn corekv.Txn
	id  uint64

	successFns []func()
	errorFns   []func()
	discardFns []func()

	successAsyncFns []func()
	errorAsyncFns   []func()
	discardAsyncFns []func()
}

// var _ Txn = (*txn)(nil)

// newTxnFrom returns a new Txn from the rootstore.
func NewTxnFrom(ctx context.Context, rootstore corekv.TxnStore, id uint64, readonly bool) *Txn {
	rootTxn := rootstore.NewTxn(readonly)
	multistore := NewMultistore(rootTxn)
	return &Txn{
		Multistore: multistore,
		txn:        rootTxn,
		id:         id,
	}
}

func (t *Txn) ID() uint64 {
	return t.id
}

func (t *Txn) Commit(ctx context.Context) error {
	var fns []func()
	var asyncFns []func()

	err := t.txn.Commit()
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

func (t *Txn) Discard(ctx context.Context) {
	t.txn.Discard()

	for _, fn := range t.discardAsyncFns {
		go fn()
	}
	for _, fn := range t.discardFns {
		fn()
	}
}

func (t *Txn) OnSuccess(fn func()) {
	t.successFns = append(t.successFns, fn)
}

func (t *Txn) OnError(fn func()) {
	t.errorFns = append(t.errorFns, fn)
}

func (t *Txn) OnDiscard(fn func()) {
	t.discardFns = append(t.discardFns, fn)
}

func (t *Txn) OnSuccessAsync(fn func()) {
	t.successAsyncFns = append(t.successAsyncFns, fn)
}

func (t *Txn) OnErrorAsync(fn func()) {
	t.errorAsyncFns = append(t.errorAsyncFns, fn)
}

func (t *Txn) OnDiscardAsync(fn func()) {
	t.discardAsyncFns = append(t.discardAsyncFns, fn)
}
