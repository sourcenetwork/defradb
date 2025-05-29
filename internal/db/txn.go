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
//
// The returned context will contain the transaction
// along with the copied values from the input context.
func ensureContextTxn(ctx context.Context, db transactionDB, readOnly bool) (context.Context, datastore.Txn, error) {
	// explicit transaction
	txn, ok := datastore.TryGetTxn(ctx)
	if ok {
		return InitContext(ctx, &explicitTxn{txn}), &explicitTxn{txn}, nil
	}
	txn, err := db.NewTxn(ctx, readOnly)
	if err != nil {
		return nil, txn, err
	}
	return InitContext(ctx, txn), txn, nil
}
