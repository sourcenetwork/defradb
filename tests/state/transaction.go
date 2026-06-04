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

package state

import (
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/immutable"
)

// GetTransaction returns the transaction, creating one if needed.
func (s *State) GetTransaction(
	db client.TxnStore,
	transactionSpecifier immutable.Option[int],
) (client.Txn, error) {
	if !transactionSpecifier.HasValue() {
		return nil, nil
	}

	transactionID := transactionSpecifier.Value()
	if transactionID >= len(s.Txns) {
		// Extend the txn slice so this txn can fit and be accessed by TransactionID
		s.Txns = append(s.Txns, make([]client.Txn, transactionID-len(s.Txns)+1)...)
	}

	if s.Txns[transactionID] == nil {
		txn, err := db.NewTxn(false)
		if err != nil {
			txn.Discard()
			return nil, err
		}

		s.Txns[transactionID] = txn
	}

	return s.Txns[transactionID], nil
}
