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
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/state"
)

// listIndexesOptions returns the ListIndexes options carrying the node's identity, so the call is
// authorized when NAC is enabled. Waiting for a build is test infrastructure, not a user operation,
// so it uses the node identity to bypass NAC rather than the test's client identity. ListIndexes
// reads the identity from its options, not the context.
func listIndexesOptions(s *state.State, node *state.NodeState) options.Enumerable[options.ListCollectionIndexesOptions] {
	nodeID := -1
	for i, n := range s.Nodes {
		if n == node {
			nodeID = i
			break
		}
	}
	builder := options.ListCollectionIndexes()
	identOption := getIdentityForRequestSpecificToNode(s, NodeIdentity(nodeID), nodeID)
	if identOption.HasValue() {
		builder.SetIdentity(identOption.Value())
	}
	return builder
}

// WaitForNodeIndexesBuilt blocks until every index on every collection of the node has left the
// building state. A collection patch or version switch that reindexes stages the builds and returns;
// the worker runs them in the background, so a following query would otherwise race the build. Call
// this after such an action so the next query sees a built index.
func WaitForNodeIndexesBuilt(s *state.State, node *state.NodeState) {
	if node.Closed {
		return
	}
	opts := listIndexesOptions(s, node)
	// Fetch collections outside any transaction: the caller's txn may already be committed, and a
	// collection bound to it would fail ListIndexes with "transaction not found".
	for _, collection := range MustGetCanonicallyOrderedCollections(s, node, immutable.None[client.Txn]()) {
		if collection == nil {
			continue
		}
		results, err := collection.ListIndexes(s.Ctx, opts)
		require.NoError(s.T, err)
		for _, r := range results {
			waitForIndexBuilt(s, collection, r.Description.ID, opts)
		}
	}
}

// WaitForIndexReady blocks until every index on the collection has finished building (ready or
// failed). It is for tests that create an index on an explicit transaction, where NewIndex cannot
// wait (the build runs only after the caller commits), and then need it built before querying.
type WaitForIndexReady struct {
	stateful

	// NodeID may hold the ID (index) of a node to wait on. If not provided, the first node is used.
	NodeID immutable.Option[int]

	// The collection whose indexes to wait on.
	CollectionID int
}

var _ Action = (*WaitForIndexReady)(nil)
var _ Stateful = (*WaitForIndexReady)(nil)

func (a *WaitForIndexReady) Execute() {
	if len(a.s.Nodes) == 0 {
		return
	}

	nodeIDs, _ := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for _, nodeID := range nodeIDs {
		node := a.s.Nodes[nodeID]
		if node.Closed {
			continue
		}
		collections := MustGetCanonicallyOrderedCollections(a.s, node, immutable.None[client.Txn]())
		collection := collections[a.CollectionID]

		opts := listIndexesOptions(a.s, node)
		results, err := collection.ListIndexes(a.s.Ctx, opts)
		require.NoError(a.s.T, err)
		for _, r := range results {
			waitForIndexBuilt(a.s, collection, r.Description.ID, opts)
		}
	}
}
