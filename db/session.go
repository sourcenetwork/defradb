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

// Session wraps a context to make it easier to pass request scoped
// parameters such as transactions.
type Session struct {
	context.Context
}

// NewSession returns a new session that wraps the given context.
func NewSession(ctx context.Context) *Session {
	return &Session{ctx}
}

// WithTxn returns a new session with the transaction value set.
func (s *Session) WithTxn(txn datastore.Txn) *Session {
	return &Session{setContextTxn(s, txn)}
}
