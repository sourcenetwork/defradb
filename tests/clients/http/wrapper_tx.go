// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package http

import (
	"context"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/datastore"
)

var _ datastore.Txn = (*TxWrapper)(nil)

// TxWrapper combines a client and server transaction into
// a single struct that implements the datastore.Txn interface.
type TxWrapper struct {
	server datastore.Txn
	client datastore.Txn
}

func (w *TxWrapper) Store() corekv.Store {
	return w.server.Store()
}

func (w *TxWrapper) ID() uint64 {
	return w.client.ID()
}

func (w *TxWrapper) Commit(ctx context.Context) error {
	return w.client.Commit(ctx)
}

func (w *TxWrapper) Discard(ctx context.Context) {
	w.client.Discard(ctx)
}

func (w *TxWrapper) OnSuccess(fn func()) {
	w.server.OnSuccess(fn)
}

func (w *TxWrapper) OnError(fn func()) {
	w.server.OnError(fn)
}

func (w *TxWrapper) OnDiscard(fn func()) {
	w.server.OnDiscard(fn)
}

func (w *TxWrapper) OnSuccessAsync(fn func()) {
	w.server.OnSuccessAsync(fn)
}

func (w *TxWrapper) OnErrorAsync(fn func()) {
	w.server.OnErrorAsync(fn)
}

func (w *TxWrapper) OnDiscardAsync(fn func()) {
	w.server.OnDiscardAsync(fn)
}
