// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package action

import (
	"fmt"

	"github.com/sourcenetwork/defradb/cli/test/state"
)

// TxnAction wraps an action within the context of a transaction.
//
// Executing this TxnAction will execute the given action within the scope
// of the given transaction.
type TxnAction[T ArgmentedAction] struct {
	s *state.State
	argmented

	TxnIndex int
	Action   T
}

var _ Action = (*TxnAction[ArgmentedAction])(nil)

// WithTxn wraps the given action within the scope of the default (ID: 0) transaction.
//
// If a transaction with the default ID (0) has not been created by the time this action
// executes, executing the transaction will panic.
func WithTxn[T ArgmentedAction](action T) *TxnAction[T] {
	return &TxnAction[T]{
		Action: action,
	}
}

// WithTxn wraps the given action within the scope of given transaction.
//
// If a transaction with the given ID has not been created by the time this action
// executes, executing the transaction will panic.
func WithTxnI[T ArgmentedAction](action T, txnIndex int) *TxnAction[T] {
	return &TxnAction[T]{
		TxnIndex: txnIndex,
		Action:   action,
	}
}

func (a *TxnAction[T]) SetState(s *state.State) {
	a.s = s
	if inner, ok := any(a.Action).(Stateful); ok {
		inner.SetState(s)
	}
}

func (a *TxnAction[T]) Execute() {
	a.Action.AddArgs("tx", fmt.Sprint(a.s.Txns[a.TxnIndex]))
	a.Action.Execute()
}
