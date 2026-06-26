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
)

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

		results, err := collection.ListIndexes(a.s.Ctx)
		require.NoError(a.s.T, err)
		for _, r := range results {
			waitForIndexBuilt(a.s, collection, r.Description.ID)
		}
	}
}
