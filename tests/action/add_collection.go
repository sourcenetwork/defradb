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
	"github.com/sourcenetwork/defradb/tests/state"
)

// AddCollection is an action that will add the given GQL SDL to the Defra nodes.
type AddCollection struct {
	stateful

	// NodeID may hold the ID (index) of a node to apply this update to.
	//
	// If a value is not provided the update will be applied to all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// The collection definition SDL to add.
	SDL string

	// Optionally, the expected results.
	//
	// Each item will be compared individually, if ID, RootID, CollectionVersionID or Fields on the
	// expected item are default they will not be compared with the actual.
	//
	// Assertions on Indexes and Sources will not distinguish between nil and empty (in order
	// to allow their ommission in most cases).
	ExpectedResults []client.CollectionVersion

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

var _ Action = (*AddCollection)(nil)
var _ Stateful = (*AddCollection)(nil)

func (a *AddCollection) Execute() {
	nodeIDs, nodes := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		sdl := replace(a.s, nodeID, a.SDL)

		opts := options.AddCollection()
		identOption := getIdentityForRequestSpecificToNode(a.s, a.Identity, nodeID)
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}
		results, err := node.AddCollection(a.s.Ctx, sdl, opts)

		for _, result := range results {
			appendCollectionVersion(a.s, result.VersionID)
		}

		expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)

		assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)

		if a.ExpectedResults != nil {
			assertCollectionVersions(a.s, a.ExpectedResults, results)
		}
	}

	// If the collection was updated we need to refresh the collection definitions.
	refreshCollections(a.s)
}
