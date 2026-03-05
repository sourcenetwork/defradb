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

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/state"
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

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

var _ Action = (*DeleteIndex)(nil)
var _ Stateful = (*DeleteIndex)(nil)

func (a *DeleteIndex) Execute() {
	var expectedErrorRaised bool

	nodeIDs, _ := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for _, nodeID := range nodeIDs {
		collection := a.s.Nodes[nodeID].Collections[a.CollectionID]

		opts := options.DeleteCollectionIndex()
		identOption := getIdentityForRequestSpecificToNode(a.s, a.Identity, nodeID)
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}

		err := collection.DeleteIndex(a.s.Ctx, a.IndexName, opts)

		expectedErrorRaised = assertError(a.s.T, err, a.ExpectedError)
	}

	assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)
}
