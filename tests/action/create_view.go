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
	"fmt"
	"strings"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/internal/request/graphql/schema/types"
	"github.com/sourcenetwork/defradb/tests/state"
)

// CreateView is an action that will create a new View.
type CreateView struct {
	stateful

	// NodeID may hold the ID (index) of a node to create this View on.
	//
	// If a value is not provided the view will be created on all nodes.
	NodeID immutable.Option[int]

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
}

var _ Action = (*CreateView)(nil)
var _ Stateful = (*CreateView)(nil)

// Execute executes the create view action.
func (a *CreateView) Execute() {
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
		transformCID := a.TransformCID
		if transformCID.HasValue() {
			transformCID = immutable.Some(replace(a.s, nodeIDs[i], transformCID.Value()))
		}
		results, err := node.AddView(a.s.Ctx, a.Query, sdl, transformCID)

		for _, result := range results {
			appendCollectionVersion(a.s, result.VersionID)
		}

		expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)

		assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)
	}

	refreshCollections(a.s)
}
