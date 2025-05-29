// Copyright 2025 Democratized Data Foundation
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

type key struct{}

// MustGetTxn returns the transaction from the context or panics.
func MustGetTxn(ctx context.Context) Txn {
	return ctx.Value(key{}).(Txn) //nolint:forcetypeassert
}

// TryGetTxn returns a transaction and a bool indicating if the
// txn was retrieved from the given context.
func TryGetTxn(ctx context.Context) (Txn, bool) {
	txn, ok := ctx.Value(key{}).(Txn)
	return txn, ok
}

// SetTxn returns a new context with the txn value set.
//
// This will overwrite any previously set transaction value.
func SetTxn(ctx context.Context, txn Txn) context.Context {
	return context.WithValue(ctx, key{}, txn)
}

// EnsureContextTxn ensures that the returned context has a transaction.
//
// If a transactions exists on the context it will be made explicit,
// otherwise a new implicit transaction will be created.
//
// The returned context will contain the transaction
// along with the copied values from the input context.
func EnsureContextTxn(ctx context.Context, store corekv.TxnStore, readOnly bool) (context.Context, Txn) {
	// explicit transaction
	txn, ok := TryGetTxn(ctx)
	if ok {
		return SetTxn(ctx, txn), &explicitTxn{txn}
	}
	txn = NewTxnFrom(ctx, store, 0, readOnly)
	return SetTxn(ctx, txn), txn
}
