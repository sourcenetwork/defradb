// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package txnctx

import (
	"context"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/datastore"
)

type Txn interface {
	Blockstore() datastore.Blockstore

	Datastore() corekv.ReaderWriter

	Encstore() datastore.Blockstore

	Headstore() corekv.ReaderWriter

	Peerstore() corekv.ReaderWriter

	Rootstore() corekv.ReaderWriter

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

type key struct{}

// MustGet returns the transaction from the context or panics.
func MustGet(ctx context.Context) Txn {
	return ctx.Value(key{}).(Txn) //nolint:forcetypeassert
}

// TryGet returns a transaction and a bool indicating if the
// txn was retrieved from the given context.
func TryGet(ctx context.Context) (Txn, bool) {
	txn, ok := ctx.Value(key{}).(Txn)
	return txn, ok
}

// Set returns a new context with the txn value set.
//
// This will overwrite any previously set transaction value.
func Set(ctx context.Context, txn Txn) context.Context {
	return context.WithValue(ctx, key{}, txn)
}
