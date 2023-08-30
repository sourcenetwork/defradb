// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"context"
	"fmt"

	"github.com/sourcenetwork/defradb/datastore"
)

var _ datastore.Txn = (*TxWrapper)(nil)

type TxWrapper struct {
	tx  datastore.Txn
	cmd *cliWrapper
}

func (w *TxWrapper) ID() uint64 {
	return w.tx.ID()
}

func (w *TxWrapper) Commit(ctx context.Context) error {
	args := []string{"client", "tx", "commit"}
	args = append(args, fmt.Sprintf("%d", w.tx.ID()))

	_, err := w.cmd.execute(ctx, args)
	return err
}

func (w *TxWrapper) Discard(ctx context.Context) {
	args := []string{"client", "tx", "discard"}
	args = append(args, fmt.Sprintf("%d", w.tx.ID()))

	w.cmd.execute(ctx, args)
}

func (w *TxWrapper) OnSuccess(fn func()) {
	w.tx.OnSuccess(fn)
}

func (w *TxWrapper) OnError(fn func()) {
	w.tx.OnError(fn)
}

func (w *TxWrapper) OnDiscard(fn func()) {
	w.tx.OnDiscard(fn)
}

func (w *TxWrapper) Rootstore() datastore.DSReaderWriter {
	return w.tx.Rootstore()
}

func (w *TxWrapper) Datastore() datastore.DSReaderWriter {
	return w.tx.Datastore()
}

func (w *TxWrapper) Headstore() datastore.DSReaderWriter {
	return w.tx.Headstore()
}

func (w *TxWrapper) DAGstore() datastore.DAGStore {
	return w.tx.DAGstore()
}

func (w *TxWrapper) Systemstore() datastore.DSReaderWriter {
	return w.tx.Systemstore()
}
