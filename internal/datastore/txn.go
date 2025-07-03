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

	"github.com/sourcenetwork/defradb/client"
)

// Txn is a common interface to the BasicTxn struct.
type Txn interface {
	// Blockstore returns the prefixed store for the blockstore
	Blockstore() Blockstore

	// Datastore returns the prefixed store for the datastore
	Datastore() corekv.ReaderWriter

	// Encstore returns the prefixed store for the encryption key store
	Encstore() Blockstore

	// Headstore returns the prefixed store for the headstore
	Headstore() corekv.ReaderWriter

	// Peerstore returns the prefixed store for the peerstore
	Peerstore() corekv.ReaderWriter

	// Rootstore returns the rootstore
	Rootstore() corekv.ReaderWriter

	// Systemstore returns the prefixed store for the systemstore
	Systemstore() corekv.ReaderWriter

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

type BasicTxn struct {
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

var _ Txn = (*BasicTxn)(nil)

// newTxnFrom returns a new Txn from the rootstore.
func NewTxnFrom(ctx context.Context, rootstore corekv.TxnStore, id uint64, readonly bool) *BasicTxn {
	rootTxn := rootstore.NewTxn(readonly)
	multistore := NewMultistore(rootTxn)
	return &BasicTxn{
		Multistore: multistore,
		txn:        rootTxn,
		id:         id,
	}
}

func (t *BasicTxn) ID() uint64 {
	return t.id
}

func (t *BasicTxn) Commit(ctx context.Context) error {
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

func (t *BasicTxn) Discard(ctx context.Context) {
	t.txn.Discard()

	for _, fn := range t.discardAsyncFns {
		go fn()
	}
	for _, fn := range t.discardFns {
		fn()
	}
}

func (t *BasicTxn) OnSuccess(fn func()) {
	t.successFns = append(t.successFns, fn)
}

func (t *BasicTxn) OnError(fn func()) {
	t.errorFns = append(t.errorFns, fn)
}

func (t *BasicTxn) OnDiscard(fn func()) {
	t.discardFns = append(t.discardFns, fn)
}

func (t *BasicTxn) OnSuccessAsync(fn func()) {
	t.successAsyncFns = append(t.successAsyncFns, fn)
}

func (t *BasicTxn) OnErrorAsync(fn func()) {
	t.errorAsyncFns = append(t.errorAsyncFns, fn)
}

func (t *BasicTxn) OnDiscardAsync(fn func()) {
	t.discardAsyncFns = append(t.discardAsyncFns, fn)
}

type txnKey struct{}

// CtxMustGetTxn returns the transaction from the context or panics.
func CtxMustGetTxn(ctx context.Context) Txn {
	return ctx.Value(txnKey{}).(Txn) //nolint:forcetypeassert
}

// CtxTryGetTxn returns a transaction and a bool indicating if the
// txn was retrieved from the given context.
func CtxTryGetTxn(ctx context.Context) (Txn, bool) {
	txn, ok := ctx.Value(txnKey{}).(Txn)
	return txn, ok
}

// CtxTryGetClientTxn returns a client transaction and a bool indicating if the
// txn was retrieved from the given context.
func CtxTryGetClientTxn(ctx context.Context) (client.Txn, bool) {
	txn, ok := ctx.Value(txnKey{}).(client.Txn)
	return txn, ok
}

// CtxSetTxn returns a new context with the txn value set.
//
// This will overwrite any previously set transaction value.
func CtxSetTxn(ctx context.Context, txn Txn) context.Context {
	return context.WithValue(ctx, txnKey{}, txn)
}

// CtxSetFromClientTxn returns a new context with the txn value set.
//
// This will overwrite any previously set transaction value.
func CtxSetFromClientTxn(ctx context.Context, txn client.Txn) context.Context {
	return context.WithValue(ctx, txnKey{}, txn)
}

// MustGetFromClientTxn returns the a Txn from a client.Txn or panics.
func MustGetFromClientTxn(txn client.Txn) Txn {
	return txn.(Txn) //nolint:forcetypeassert
}
