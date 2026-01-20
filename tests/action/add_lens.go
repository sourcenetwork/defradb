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
}

var _ Action = (*AddLens)(nil)
var _ Stateful = (*AddLens)(nil)

func (a *AddLens) Execute() {
	var lensID string

	nodeIDs, nodes := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		a.s.Ctx = getContextWithIdentity(a.s.Ctx, a.s, a.Identity, nodeID)

		var err error
		lensID, err = node.AddLens(a.s.Ctx, a.Lens)

		resetStateContext(a.s)

		if err != nil {
			a.s.T.Fatalf("failed to add lens: %v", err)
		}
	}

	a.s.LensIDs = append(a.s.LensIDs, lensID)
}
