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
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/state"
)

// ActivePeers is an action that will get the active peers from the given node(s)
// and assert that the given expect result matches the actual.
type ActivePeers struct {
	stateful

	// NodeID holds the ID (index) of a node to get active peers for.
	NodeID int

	// The expected set of results.
	//
	// Respects `replace`, and should typically be provided a string similar to
	// `{{.Peer1_Address0}}`.
	//
	// The order of elements in the given slice is not asserted.
	Expected []string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

var _ Action = (*ActivePeers)(nil)
var _ Stateful = (*ActivePeers)(nil)

func (a *ActivePeers) Execute() {
	nodeIDs, nodes := getNodesWithIDs(immutable.Some(a.NodeID), a.s.Nodes)
	for index, node := range nodes {
		actual, err := node.ActivePeers(a.s.Ctx)

		if a.ExpectedError != "" {
			require.Contains(a.s.T, err.Error(), a.ExpectedError)
			continue
		}

		require.NoError(a.s.T, err)

		expected := cloneAndReplacePeerInfos(a.s, nodeIDs[index], a.Expected)

		require.ElementsMatch(a.s.T, expected, actual)
	}
}

func cloneAndReplacePeerInfos(s *state.State, nodeID int, addresses []string) []string {
	result := make([]string, len(addresses))
	for i, address := range addresses {
		result[i] = replace(s, nodeID, address)
	}
	return result
}
