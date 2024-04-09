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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
)

type contextKey string

const (
	txnContextKey = contextKey("txn")
)

// Session wraps a context to make it easier to pass request scoped
// parameters such as transactions.
type Session struct {
	context.Context
}

// NewSession returns a session that wraps the given context.
func NewSession(ctx context.Context) *Session {
	return &Session{ctx}
}

// WithTxn returns a new session with the transaction value set.
func (s *Session) WithTxn(txn datastore.Txn) *Session {
	return &Session{context.WithValue(s, txnContextKey, txn)}
}

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

// getContextTxn returns the explicit transaction from
// the context or creates a new implicit one.
func getContextTxn(ctx context.Context, db client.DB, readOnly bool) (datastore.Txn, error) {
	txn, ok := ctx.Value(txnContextKey).(datastore.Txn)
	if ok {
		return &explicitTxn{txn}, nil
	}
	return db.NewTxn(ctx, readOnly)
}
