// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"

	"github.com/sourcenetwork/defradb/datastore"
)

// TxnContextKey is the key type for transaction context values.
type TxnContextKey struct{}

// explicitTxn is a transaction that is managed outside of a db operation.
type explicitTxn struct {
	datastore.Txn
}

func (t *explicitTxn) Commit(ctx context.Context) error {
	return nil // do nothing
}

func (t *explicitTxn) Discard(ctx context.Context) {
	// do nothing
}

// transactionDB is a db that can create transactions.
type transactionDB interface {
	NewTxn(context.Context, bool) (datastore.Txn, error)
}

// ensureContextTxn ensures that the returned context has a transaction.
//
// If a transactions exists on the context it will be made explicit,
// otherwise a new implicit transaction will be created.
func ensureContextTxn(ctx context.Context, db transactionDB, readOnly bool) (context.Context, error) {
	txn, ok := ctx.Value(TxnContextKey{}).(datastore.Txn)
	if ok {
		return setContextTxn(ctx, &explicitTxn{txn}), nil
	}
	txn, err := db.NewTxn(ctx, readOnly)
	if err != nil {
		return nil, err
	}
	return setContextTxn(ctx, txn), nil
}

// mustGetContextTxn returns the transaction from the context if it exists,
// otherwise it panics.
func mustGetContextTxn(ctx context.Context) datastore.Txn {
	return ctx.Value(TxnContextKey{}).(datastore.Txn)
}

// setContextTxn returns a new context with the txn value set.
func setContextTxn(ctx context.Context, txn datastore.Txn) context.Context {
	return context.WithValue(ctx, TxnContextKey{}, txn)
}
