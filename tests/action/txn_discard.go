// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package action

// DiscardTransaction represents a discard request for a transaction of the given id.
type DiscardTransaction struct {
	stateful

	// Used to identify the transaction to discard.
	TransactionID int
}

var _ Action = (*DiscardTransaction)(nil)
var _ Stateful = (*DiscardTransaction)(nil)

func (a *DiscardTransaction) Execute() {
	a.s.Txns[a.TransactionID].Discard()
}
