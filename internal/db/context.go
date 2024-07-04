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

	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/encryption"
)

// txnContextKey is the key type for transaction context values.
type txnContextKey struct{}

// identityContextKey is the key type for ACP identity context values.
type identityContextKey struct{}

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

// ensureContextTxn ensures that the returned context has a transaction
// and an identity.
//
// If a transactions exists on the context it will be made explicit,
// otherwise a new implicit transaction will be created.
//
// The returned context will contain the transaction and identity
// along with the copied values from the input context.
func ensureContextTxn(ctx context.Context, db transactionDB, readOnly bool) (context.Context, datastore.Txn, error) {
	// explicit transaction
	txn, ok := TryGetContextTxn(ctx)
	if ok {
		return SetContextTxn(ctx, &explicitTxn{txn}), &explicitTxn{txn}, nil
	}
	// implicit transaction
	txn, err := db.NewTxn(ctx, readOnly)
	if err != nil {
		return nil, txn, err
	}
	ctx = encryption.ContextWithStore(ctx, txn)
	return SetContextTxn(ctx, txn), txn, nil
}

// mustGetContextTxn returns the transaction from the context or panics.
//
// This should only be called from private functions within the db package
// where we ensure an implicit or explicit transaction always exists.
func mustGetContextTxn(ctx context.Context) datastore.Txn {
	return ctx.Value(txnContextKey{}).(datastore.Txn)
}

// TryGetContextTxn returns a transaction and a bool indicating if the
// txn was retrieved from the given context.
func TryGetContextTxn(ctx context.Context) (datastore.Txn, bool) {
	txn, ok := ctx.Value(txnContextKey{}).(datastore.Txn)
	return txn, ok
}

// SetContextTxn returns a new context with the txn value set.
//
// This will overwrite any previously set transaction value.
func SetContextTxn(ctx context.Context, txn datastore.Txn) context.Context {
	return context.WithValue(ctx, txnContextKey{}, txn)
}

// GetContextIdentity returns the identity from the given context.
//
// If an identity does not exist `NoIdentity` is returned.
func GetContextIdentity(ctx context.Context) immutable.Option[acpIdentity.Identity] {
	identity, ok := ctx.Value(identityContextKey{}).(acpIdentity.Identity)
	if ok {
		return immutable.Some(identity)
	}
	return acpIdentity.None
}

// SetContextTxn returns a new context with the identity value set.
//
// This will overwrite any previously set identity value.
func SetContextIdentity(ctx context.Context, identity immutable.Option[acpIdentity.Identity]) context.Context {
	if identity.HasValue() {
		return context.WithValue(ctx, identityContextKey{}, identity.Value())
	}
	return context.WithValue(ctx, identityContextKey{}, nil)
}
