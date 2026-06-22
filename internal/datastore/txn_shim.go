// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package datastore

import (
	"time"

	"github.com/sourcenetwork/corekv"
)

// TxnShim is a storeless implementation of the [Txn] interface.
//
// Calling any of the child store accessors will result in a panic.
// The majority of Defra does not check for nil being returned from these
// functions, so we might as well panic within this type where the stack
// trace should be a little more obvious.
type TxnShim struct {
	id uint64
	ts time.Time

	successFns []func()
	errorFns   []func()
	discardFns []func()

	successAsyncFns []func()
	errorAsyncFns   []func()
	discardAsyncFns []func()
}

var _ Txn = (*TxnShim)(nil)

func NewTxnShim(
	id uint64,
) *TxnShim {
	return &TxnShim{
		id: id,
		ts: time.Now(),
	}
}

func (t *TxnShim) Blockstore() Blockstore {
	panic("unimplemented")
}

func (t *TxnShim) Commit() error {
	for _, fn := range t.successAsyncFns {
		go fn()
	}
	for _, fn := range t.successFns {
		fn()
	}
	return nil
}

func (t *TxnShim) Datastore() Keyedstore {
	panic("unimplemented")
}

func (t *TxnShim) Discard() {
	for _, fn := range t.discardAsyncFns {
		go fn()
	}
	for _, fn := range t.discardFns {
		fn()
	}
}

func (t *TxnShim) Encstore() Blockstore {
	panic("unimplemented")
}

func (t *TxnShim) Headstore() corekv.ReaderWriter {
	panic("unimplemented")
}

func (t *TxnShim) ID() uint64 {
	return t.id
}

func (t *TxnShim) OnSuccess(fn func()) {
	t.successFns = append(t.successFns, fn)
}

func (t *TxnShim) OnError(fn func()) {
	t.errorFns = append(t.errorFns, fn)
}

func (t *TxnShim) OnDiscard(fn func()) {
	t.discardFns = append(t.discardFns, fn)
}

func (t *TxnShim) OnSuccessAsync(fn func()) {
	t.successAsyncFns = append(t.successAsyncFns, fn)
}

func (t *TxnShim) OnErrorAsync(fn func()) {
	t.errorAsyncFns = append(t.errorAsyncFns, fn)
}

func (t *TxnShim) OnDiscardAsync(fn func()) {
	t.discardAsyncFns = append(t.discardAsyncFns, fn)
}

func (t *TxnShim) Peerstore() corekv.ReaderWriter {
	panic("unimplemented")
}

func (t *TxnShim) Rootstore() corekv.ReaderWriter {
	panic("unimplemented")
}

func (t *TxnShim) StartTS() time.Time {
	return t.ts
}

func (t *TxnShim) Systemstore() corekv.ReaderWriter {
	panic("unimplemented")
}
