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

	"github.com/sourcenetwork/defradb/datastore"
)

var _ datastore.Txn = (*TxWrapper)(nil)

type TxWrapper struct {
	server datastore.Txn
	client datastore.Txn
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

func (w *TxWrapper) Rootstore() datastore.DSReaderWriter {
	return w.server.Rootstore()
}

func (w *TxWrapper) Datastore() datastore.DSReaderWriter {
	return w.server.Datastore()
}

func (w *TxWrapper) Headstore() datastore.DSReaderWriter {
	return w.server.Headstore()
}

func (w *TxWrapper) DAGstore() datastore.DAGStore {
	return w.server.DAGstore()
}

func (w *TxWrapper) Systemstore() datastore.DSReaderWriter {
	return w.server.Systemstore()
}
