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

// CommitTransaction represents a commit request for a transaction of the given id.
type CommitTransaction struct {
	stateful

	// Used to identify the transaction to commit.
	TransactionID int

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

var _ Action = (*CommitTransaction)(nil)
var _ Stateful = (*CommitTransaction)(nil)

func (a *CommitTransaction) Execute() {
	err := a.s.Txns[a.TransactionID].Commit()
	if err != nil {
		a.s.Txns[a.TransactionID].Discard()
	}

	RefreshCollections(a.s)

	expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)

	assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)
}
