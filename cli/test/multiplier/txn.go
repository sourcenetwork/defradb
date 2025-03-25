// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package multiplier

import (
	"github.com/sourcenetwork/testo/multiplier"

	"github.com/sourcenetwork/defradb/cli/test/action"
)

func init() {
	multiplier.Register(&txnCommit{})
}

const TxnCommit Name = "txn-commit"

type txnCommit struct{}

var _ Multiplier = (*txnCommit)(nil)

func (n *txnCommit) Name() Name {
	return TxnCommit
}

func (n *txnCommit) Apply(source action.Actions) action.Actions {
	lastStartIndex := 0
	lastWriteIndex := 0

	for i, a := range source {
		switch a.(type) {
		case *action.StartCli:
			lastStartIndex = i

		case *action.SchemaAdd:
			lastWriteIndex = i

		case *action.TxCreate:
			// If the action set already contains txns we should not adjust it
			return source
		}
	}

	result := make(action.Actions, 0, len(source)+2)

	for i, a := range source {
		switch a.(type) {
		case *action.StartCli:
			result = append(result, a)
			result = append(result, &action.TxCreate{})

		default:
			if augmentedAction, ok := a.(action.AugmentedAction); ok && i > lastStartIndex {
				result = append(result, action.WithTxn(augmentedAction))
			} else {
				result = append(result, a)
			}
		}
	}

	result = append(result, &action.TxCommit{})

	for i, a := range source {
		if i <= lastStartIndex {
			continue
		}

		if i <= lastWriteIndex {
			continue
		}

		switch a.(type) {
		// Append orginal, read-only actions that occured after the last write action,
		// after the transaction has been commited.  This allows the automatic of whether
		// or not the transaction-state has been persisted.
		case *action.CollectionDescribe:
			result = append(result, a)
		}
	}

	return result
}
