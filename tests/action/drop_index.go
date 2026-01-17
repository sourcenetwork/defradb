// Copyright 2026 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/tests/state"
)

// DropIndex will attempt to drop the given secondary index from the given collection
// using the collection api.
type DropIndex struct {
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

var _ Action = (*DropIndex)(nil)
var _ Stateful = (*DropIndex)(nil)

func (a *DropIndex) Execute() {
	var expectedErrorRaised bool

	nodeIDs, _ := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for _, nodeID := range nodeIDs {
		collection := a.s.Nodes[nodeID].Collections[a.CollectionID]

		ctx := getContextWithIdentity(a.s.Ctx, a.s, a.Identity, nodeID)
		err := collection.DropIndex(ctx, a.IndexName)

		expectedErrorRaised = assertError(a.s.T, err, a.ExpectedError)
	}

	assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)
}
