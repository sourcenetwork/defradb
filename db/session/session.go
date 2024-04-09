package session

import (
	"context"

	"github.com/sourcenetwork/defradb/datastore"
)

type contextKey string

const (
	TxnContextKey = contextKey("txn")
)

// Session wraps a context to make it easier to pass request scoped
// parameters such as transactions.
type Session struct {
	context.Context
}

// New returns a new session that wraps the given context.
func New(ctx context.Context) *Session {
	return &Session{ctx}
}

// WithTxn returns a new session with the transaction value set.
func (s *Session) WithTxn(txn datastore.Txn) *Session {
	return &Session{context.WithValue(s, TxnContextKey, txn)}
}
