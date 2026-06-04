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
	"fmt"
	"strings"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/request/graphql/schema/types"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

// AddView is an action that will add a new View.
type AddView struct {
	stateful

	// NodeID may hold the ID (index) of a node to create this View on.
	//
	// If a value is not provided the view will be created on all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// The query that this View is to be based off of. Required.
	Query string

	// The SDL containing all types used by the view output.
	SDL string

	// An optional CID of an existing lens transform.
	// Use AddLens action first to store the lens and get its CID.
	TransformCID immutable.Option[string]

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// Used to identify the transaction for this to be executed in. Optional.
	TransactionID immutable.Option[int]
}

var _ Action = (*AddView)(nil)
var _ Stateful = (*AddView)(nil)

// Execute executes the create view action.
func (a *AddView) Execute() {
	sdl := a.SDL

	switch {
	case strings.Contains(sdl, "@materialized(if: false)"):
		if a.s.ViewType == state.MaterializedViewType {
			sdl = strings.ReplaceAll(sdl, "@materialized(if: false)", "@materialized(if: true)")
		}

	case strings.Contains(sdl, "@materialized(if: true)"):
		if a.s.ViewType == state.CachelessViewType {
			sdl = strings.ReplaceAll(sdl, "@materialized(if: true)", "@materialized(if: false)")
		}

	default:
		typeIndex := strings.Index(sdl, "\ttype ")
		if typeIndex == -1 {
			a.s.T.Fatal("materialized view SDL must contain '\ttype ' declaration")
			return
		}

		subStrSquigglyIndex := strings.Index(sdl[typeIndex:], "{")
		if subStrSquigglyIndex == -1 {
			a.s.T.Fatal("materialized view SDL type declaration must contain '{'")
			return
		}

		squigglyIndex := typeIndex + subStrSquigglyIndex
		sdl = strings.Join([]string{
			sdl[:squigglyIndex],
			"@",
			types.MaterializedDirectiveLabel,
			"(if: ",
			fmt.Sprint(a.s.ViewType == state.MaterializedViewType),
			") ",
			sdl[squigglyIndex:],
			"",
		}, "")
	}

	nodeIDs, nodes := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for i, node := range nodes {
		nodeID := nodeIDs[i]

		opts := options.AddView()
		identOption := getIdentityForRequestSpecificToNode(a.s, a.Identity, nodeID)
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}

		if a.TransformCID.HasValue() {
			transformCID := replace(a.s, nodeID, a.TransformCID.Value())
			opts.SetTransformCID(transformCID)
		}

		// Check if a transaction is attached to this action. If so, we will be using it.
		var txn client.Txn
		var results []client.CollectionVersion
		var err error
		if a.TransactionID.HasValue() {
			txn, err = a.s.GetTransaction(node, a.TransactionID)
			require.NoError(a.s.T, err)
			results, err = txn.AddView(a.s.Ctx, a.Query, sdl, opts)
		} else {
			results, err = node.AddView(a.s.Ctx, a.Query, sdl, opts)
		}

		for _, result := range results {
			appendCollectionVersion(a.s, result.VersionID)
		}

		expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)

		assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)
	}

	if !a.TransactionID.HasValue() {
		RefreshCollections(a.s)
	}
}
