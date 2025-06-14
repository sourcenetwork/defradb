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
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/db/txnctx"
)

// InitContext returns a new context with all caches initialized and linked to
// the given transaction.
//
// This will overwrite any previously set cached values - this is desirable as
// the cached values must be tied to the transaction, otherwise we risk leaking
// information between transactions.
func InitContext(ctx context.Context, txn datastore.Txn) context.Context {
	ctx = txnctx.Set(ctx, txn)
	ctx = id.InitCollectionShortIDCache(ctx)
	ctx = id.InitFieldShortIDCache(ctx)

	return ctx
}
