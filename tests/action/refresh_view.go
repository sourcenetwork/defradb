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

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/state"
)

// RefreshViews action will execute a call to `store.RefreshViews` using the provided options.
type RefreshViews struct {
	stateful

	// NodeID may hold the ID (index) of a node to refresh views on.
	//
	// If a value is not provided the views will be refreshed on all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// The set of fetch options for the views.
	FilterOptions *options.RefreshViewsOptionsBuilder

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

var _ Action = (*RefreshViews)(nil)
var _ Stateful = (*RefreshViews)(nil)

// Execute executes the refresh views action.
func (a *RefreshViews) Execute() {
	nodeIDs, nodes := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		identOpts := options.RefreshViews()
		identOption := getIdentityForRequestSpecificToNode(a.s, a.Identity, nodeID)
		if identOption.HasValue() {
			identOpts.SetIdentity(identOption.Value())
		}

		var allOpts []options.Lister[options.RefreshViewsOptions]
		if a.FilterOptions != nil {
			a.FilterOptions.Reset()
			allOpts = append(allOpts, a.FilterOptions)
		}
		allOpts = append(allOpts, identOpts)

		err := node.RefreshViews(a.s.Ctx, allOpts...)

		expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)
		assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)
	}
}

// refreshViews refreshes views for all collection names in state.
// This is used by the Request action when view type is materialized.
func refreshViews(
	s *state.State,
	node *state.NodeState,
	identity immutable.Option[acpIdentity.Identity],
	expectedError string,
) bool {
	if s.ViewType != state.MaterializedViewType {
		return false
	}
	for _, colName := range s.CollectionNames {
		opts := options.RefreshViews().SetCollectionName(colName)
		if identity.HasValue() {
			opts.SetIdentity(identity.Value())
		}
		err := node.RefreshViews(s.Ctx, opts)
		if assertError(s.T, err, expectedError) {
			return true
		}
	}
	return false
}
