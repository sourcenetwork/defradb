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

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/datastore"
)

var _ datastore.Txn = (*Transaction)(nil)

type Transaction struct {
	tx  datastore.Txn
	cmd *cliWrapper
}

func (w *Transaction) Store() corekv.Store {
	return w.tx.Store()
}

func (w *Transaction) ID() uint64 {
	return w.tx.ID()
}

func (w *Transaction) Commit(ctx context.Context) error {
	args := []string{"client", "tx", "commit"}
	args = append(args, fmt.Sprintf("%d", w.tx.ID()))

	_, err := w.cmd.execute(ctx, args)
	return err
}

func (w *Transaction) Discard(ctx context.Context) {
	args := []string{"client", "tx", "discard"}
	args = append(args, fmt.Sprintf("%d", w.tx.ID()))

	w.cmd.execute(ctx, args) //nolint:errcheck
}

func (w *Transaction) OnSuccess(fn func()) {
	w.tx.OnSuccess(fn)
}

func (w *Transaction) OnError(fn func()) {
	w.tx.OnError(fn)
}

func (w *Transaction) OnDiscard(fn func()) {
	w.tx.OnDiscard(fn)
}

func (w *Transaction) OnSuccessAsync(fn func()) {
	w.tx.OnSuccessAsync(fn)
}

func (w *Transaction) OnErrorAsync(fn func()) {
	w.tx.OnErrorAsync(fn)
}

func (w *Transaction) OnDiscardAsync(fn func()) {
	w.tx.OnDiscardAsync(fn)
}
