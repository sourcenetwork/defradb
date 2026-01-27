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
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

type Truncate struct {
	stateful

	// NodeID may hold the ID (index) of a node to truncate.
	//
	// If a value is not provided all nodes will be truncated.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// CollectionIndex is the index of the collection to truncate.
	CollectionIndex int

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

var _ Action = (*Truncate)(nil)
var _ Stateful = (*Truncate)(nil)

func (a *Truncate) Execute() {
	nodeIDs, nodes := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for index := range nodes {
		nodeID := nodeIDs[index]
		collection := a.s.Nodes[nodeID].Collections[a.CollectionIndex]

		ctx := getContextWithIdentity(a.s.Ctx, a.s, a.Identity, nodeID)
		err := collection.Truncate(ctx)

		expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)
		assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)
	}
}
