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
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/tests/state"
)

// GetCollections is an action that fetches collections using the provided options.
//
// ID, RootID and CollectionVersionID will only be asserted on if an expected value is provided.
type GetCollections struct {
	stateful

	// NodeID may hold the ID (index) of a node to get collections from.
	//
	// If a value is not provided collections will be gotten from all nodes.
	NodeID immutable.Option[int]

	// Used to identify the transaction for this to run against. Optional.
	TransactionID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// The expected results.
	//
	// Each item will be compared individually, if CollectionID, VersionID, or FieldIDs on the
	// expected item are default they will not be compared with the actual.
	//
	// Assertions on Indexes and Sources will not distinguish between nil and empty (in order
	// to allow their omission in most cases).
	ExpectedResults []client.CollectionVersion

	// An optional set of fetch options for the collections.
	FilterOptions client.CollectionFetchOptions

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

var _ Action = (*GetCollections)(nil)
var _ Stateful = (*GetCollections)(nil)

// ReplaceMapFunc is a callback for template replacement (lens IDs).
// Wired up from tests/integration/utils.go.
var ReplaceMapFunc func(s *state.State, nodeID int, inputSet []string) map[string]string

// Execute executes the get collections action.
func (a *GetCollections) Execute() {
	// Collect transform strings from expected results for lens ID replacement
	transformSet := []string{}
	for _, col := range a.ExpectedResults {
		if col.PreviousVersion.HasValue() && col.PreviousVersion.Value().Transform.HasValue() {
			transformSet = append(transformSet, col.PreviousVersion.Value().Transform.Value())
		}
	}

	// The lens IDs are consistent across nodes, so we can patch once for all nodes.
	// This will need to change if patches want to replace more than just lens IDs.
	if ReplaceMapFunc != nil && len(transformSet) > 0 {
		transformMap := ReplaceMapFunc(a.s, 0, transformSet)

		for i, col := range a.ExpectedResults {
			if col.PreviousVersion.HasValue() && col.PreviousVersion.Value().Transform.HasValue() {
				a.ExpectedResults[i].PreviousVersion = immutable.Some(
					client.CollectionSource{
						SourceCollectionID: a.ExpectedResults[i].PreviousVersion.Value().SourceCollectionID,
						Transform:          immutable.Some(transformMap[col.PreviousVersion.Value().Transform.Value()]),
					},
				)
			}
		}
	}

	nodeIDs, nodes := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]
		txn := a.getTransaction(node)
		ctx := db.InitContext(a.s.Ctx, txn)
		ctx = getContextWithIdentity(ctx, a.s, a.Identity, nodeID)

		results, err := node.GetCollections(ctx, a.FilterOptions)
		resultDescriptions := make([]client.CollectionVersion, len(results))
		for i, col := range results {
			resultDescriptions[i] = col.Version()
		}

		expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)
		assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)

		if !expectedErrorRaised {
			assertCollectionVersions(a.s, a.ExpectedResults, resultDescriptions)
		}
	}
}

// getTransaction returns the transaction for this action, creating one if needed.
func (a *GetCollections) getTransaction(db client.TxnStore) client.Txn {
	if !a.TransactionID.HasValue() {
		return nil
	}

	transactionID := a.TransactionID.Value()

	if transactionID >= len(a.s.Txns) {
		// Extend the txn slice so this txn can fit and be accessed by TransactionId
		a.s.Txns = append(a.s.Txns, make([]client.Txn, transactionID-len(a.s.Txns)+1)...)
	}

	if a.s.Txns[transactionID] == nil {
		txn, err := db.NewTxn(false)
		if assertError(a.s.T, err, a.ExpectedError) {
			txn.Discard()
			return nil
		}

		a.s.Txns[transactionID] = txn
	}

	return a.s.Txns[transactionID]
}
