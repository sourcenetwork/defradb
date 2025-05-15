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

	"github.com/sourcenetwork/defradb/datastore"
)

type key struct{}

// MustGet returns the transaction from the context or panics.
func MustGet(ctx context.Context) datastore.Txn {
	return ctx.Value(key{}).(datastore.Txn) //nolint:forcetypeassert
}

// TryGet returns a transaction and a bool indicating if the
// txn was retrieved from the given context.
func TryGet(ctx context.Context) (datastore.Txn, bool) {
	txn, ok := ctx.Value(key{}).(datastore.Txn)
	return txn, ok
}

// Set returns a new context with the txn value set.
//
// This will overwrite any previously set transaction value.
func Set(ctx context.Context, txn datastore.Txn) context.Context {
	return context.WithValue(ctx, key{}, txn)
}
