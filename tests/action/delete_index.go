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
	"slices"
	"strconv"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

// DeleteIndex will attempt to delete the given secondary index from the given collection
// using the collection api.
type DeleteIndex struct {
	stateful

	// NodeID may hold the ID (index) of a node to delete the secondary index from.
	//
	// If a value is not provided the index will be deleted from all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// The collection from which the index should be deleted.
	CollectionID int

	// The index name of the secondary index within the collection.
	IndexName string

	// Async, when true, returns without waiting for the entry GC. The index leaves ListIndexes at
	// once (the definition is removed synchronously), but a dropping record persists until the
	// worker finishes collecting the entries.
	//
	// By default the action waits for the GC, so a following raw-entry assertion sees them gone.
	Async bool

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// Used to identify the transaction for this to be executed in. Optional.
	TransactionID immutable.Option[int]
}

var _ Action = (*DeleteIndex)(nil)
var _ Stateful = (*DeleteIndex)(nil)

func (a *DeleteIndex) Execute() {
	nodeIDs, _ := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for _, nodeID := range nodeIDs {
		node := a.s.Nodes[nodeID]

		// Check if a transaction is attached to this action. If so, we will be using it.
		txnOption := immutable.None[client.Txn]()
		if a.TransactionID.HasValue() {
			txn, err := a.s.GetTransaction(node, a.TransactionID)
			require.NoError(a.s.T, err)
			txnOption = immutable.Some(txn)
		}

		collections, err := GetCollectionsCanonically(a.s, node, txnOption, a.Identity)
		if err != nil {
			expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)
			assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)
			continue
		}

		collection := collections[a.CollectionID]

		opts := options.DeleteCollectionIndex()
		identOption := getIdentityForRequestSpecificToNode(a.s, a.Identity, nodeID)
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}

		// Capture the index ID before deletion so the wait below can target its drop record.
		indexID, hadIndex := indexIDByName(a.s, node, collection, a.IndexName)

		err = collection.DeleteIndex(a.s.Ctx, a.IndexName, opts)

		expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)
		assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)

		// Unless the test wants to observe the dropping window, wait for the drop record to clear
		// so a following raw-entry assertion sees them gone. With an explicit transaction the record
		// is not committed until the caller commits, so there is nothing to wait for yet.
		if err == nil && !a.Async && !a.TransactionID.HasValue() && hadIndex {
			waitForIndexDropped(a.s, node, collection.Version().CollectionID, indexID)
		}
	}
}

// indexIDByName returns the ID of the named index on the collection, and whether it was found.
// It lists as the node identity so the lookup is authorized when NAC is enabled (it is bookkeeping
// for the drop-wait, not the operation under test).
func indexIDByName(s *state.State, node *state.NodeState, collection client.Collection, name string) (uint32, bool) {
	results, err := collection.ListIndexes(s.Ctx, listIndexesOptions(s, node))
	require.NoError(s.T, err)
	for _, r := range results {
		if r.Description.Name == name {
			return r.Description.ID, true
		}
	}
	return 0, false
}

// waitForIndexDropped blocks until no in-progress drop record remains for the index, i.e. the GC has
// finished and cleared it. It polls ListActions so it composes with explicit Wait actions.
func waitForIndexDropped(s *state.State, node *state.NodeState, collectionID string, indexID uint32) {
	// Poll as the node identity so the ListActions call is authorized when NAC is enabled; waiting
	// for the drop is test infrastructure, not the operation under test.
	nodeID := slices.Index(s.Nodes, node)
	opts := options.ListActions()
	identOption := getIdentityForRequestSpecificToNode(s, NodeIdentity(nodeID), nodeID)
	if identOption.HasValue() {
		opts.SetIdentity(identOption.Value())
	}
	subject := strconv.FormatUint(uint64(indexID), 10)
	require.Eventually(s.T, func() bool {
		actions, err := node.ListActions(s.Ctx, opts)
		if err != nil {
			return false
		}
		for _, ex := range actions {
			if ex.CollectionID == collectionID &&
				ex.Action == client.DropIndexAction &&
				ex.Subject == subject &&
				ex.Status == client.InProgressActionStatus {
				return false
			}
		}
		return true
	}, indexBuildTimeout, time.Millisecond, "timed out waiting for index %d drop to finish", indexID)
}
