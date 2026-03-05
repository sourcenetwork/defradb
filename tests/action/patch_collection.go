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
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/state"
)

// PatchCollection executes a patch collection command, updating 0 to many collections and applying
// a migration if one is provided.
type PatchCollection struct {
	stateful

	// NodeID may hold the ID (index) of a node to apply this patch to.
	//
	// If a value is not provided the patch will be applied to all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// The Patch to apply to the collection version.
	Patch string

	// An optional migration that will be set if the patch creates any new CollectionVersions.
	Lens immutable.Option[model.Lens]

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

var _ Action = (*PatchCollection)(nil)
var _ Stateful = (*PatchCollection)(nil)

// Execute executes the patch collection action.
func (a *PatchCollection) Execute() {
	// The lens IDs are consistent across nodes, so we can patch once for all nodes.
	// This will need to change if patches want to replace more than just lens IDs.
	patch := replace(a.s, 0, a.Patch)

	nodeIDs, nodes := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		opts := options.PatchCollection()
		identOption := getIdentityForRequestSpecificToNode(a.s, a.Identity, nodeID)
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}

		err := node.PatchCollection(a.s.Ctx, patch, a.Lens, opts)
		expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)

		assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)
	}

	// If the collection was updated we need to refresh the collection definitions.
	refreshCollections(a.s)
}
