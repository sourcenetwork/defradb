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
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/tests/state"
)

// AddLens is an action that adds a lens to the lens store and stores its CID.
type AddLens struct {
	stateful

	// NodeID may hold the ID (index) of a node to add the lens to.
	//
	// If a value is not provided the lens will be added to all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	Identity immutable.Option[state.Identity]

	// The lens configuration to add.
	Lens model.Lens

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

var _ Action = (*AddLens)(nil)
var _ Stateful = (*AddLens)(nil)

func (a *AddLens) Execute() {
	var lensID string

	nodeIDs, nodes := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		ctx := getContextWithIdentity(a.s.Ctx, a.s, a.Identity, nodeID)

		var err error
		lensID, err = node.AddLens(ctx, a.Lens)

		if a.ExpectedError != "" {
			expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)
			assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)
			continue
		}

		if err != nil {
			a.s.T.Fatalf("failed to add lens: %v", err)
		}
	}

	if a.ExpectedError == "" {
		a.s.LensIDs = append(a.s.LensIDs, lensID)
	}
}
