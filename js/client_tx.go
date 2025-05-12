// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build js

package js

import (
	"context"
	"syscall/js"

	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/goji"
)

type transaction struct {
	txn datastore.Txn
}

func newTransaction(txn datastore.Txn) js.Value {
	wrapper := &transaction{txn}
	return js.ValueOf(map[string]any{
		"id":      txn.ID(),
		"commit":  goji.Async(wrapper.commit),
		"discard": goji.Async(wrapper.discard),
	})
}

func (t *transaction) commit(this js.Value, args []js.Value) (js.Value, error) {
	err := t.txn.Commit(context.Background())
	return js.Undefined(), err
}

func (t *transaction) discard(this js.Value, args []js.Value) (js.Value, error) {
	t.txn.Discard(context.Background())
	return js.Undefined(), nil
}
