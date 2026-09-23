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

import (
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/tests/state"
)

// GetCollectionByName is an action that fetches a single collection by name.
//
// ID, RootID and CollectionVersionID will only be asserted on if an expected value is provided.
type GetCollectionByName struct {
	stateful

	// NodeID may hold the ID (index) of a node to get the collection from.
	//
	// If a value is not provided the collection will be gotten from all nodes.
	NodeID immutable.Option[int]

	// Used to identify the transaction for this to run against. Optional.
	TransactionID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// The name of the collection to get.
	Name string

	// The expected result.
	//
	// If CollectionID, VersionID, or FieldIDs are default they will not be compared with the actual.
	ExpectedResult client.CollectionVersion

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

var _ Action = (*GetCollectionByName)(nil)
var _ Stateful = (*GetCollectionByName)(nil)

// Execute executes the get collection by name action.
func (a *GetCollectionByName) Execute() {
	nodeIDs, nodes := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]
		// Check if a transaction is attached to this action. If so, we will be using it.
		var txn client.Txn
		var err error
		hadTxn := a.TransactionID.HasValue()
		if hadTxn {
			txn, err = a.s.GetTransaction(node, a.TransactionID)
			if assertError(a.s.T, err, a.ExpectedError) {
				return
			}
		}
		ctx := db.InitContext(a.s.Ctx, txn)

		opts := options.GetCollectionByName()
		identOption := getIdentityForRequestSpecificToNode(a.s, a.Identity, nodeID)
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}

		// If we have a transaction, we will use it here. Otherwise we use the node.
		var result client.Collection
		if hadTxn {
			result, err = txn.GetCollectionByName(ctx, a.Name, opts)
		} else {
			result, err = node.GetCollectionByName(ctx, a.Name, opts)
		}

		expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)
		assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)

		if !expectedErrorRaised {
			assertCollectionVersions(
				a.s,
				[]client.CollectionVersion{a.ExpectedResult},
				[]client.CollectionVersion{result.Version()},
			)
		}
	}
}
