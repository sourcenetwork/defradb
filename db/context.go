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
	"github.com/sourcenetwork/defradb/db/session"
)

// explicitTxn is a transaction that is managed outside of the session.
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

// getContextTxn returns the explicit transaction from
// the context or creates a new implicit one.
func getContextTxn(ctx context.Context, db transactionDB, readOnly bool) (datastore.Txn, error) {
	txn, ok := ctx.Value(session.TxnContextKey).(datastore.Txn)
	if ok {
		return &explicitTxn{txn}, nil
	}
	return db.NewTxn(ctx, readOnly)
}
