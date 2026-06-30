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
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/state"
)

// ListActions is an action that lists all action information.
type ListActions struct {
	stateful

	// NodeID may hold the ID (index) of a node to list actions from.
	//
	// If a value is not provided the actions will be listed from all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	Identity immutable.Option[state.Identity]

	ExpectedInfo []client.ActionExecution

	// Used to identify the transaction for this to be executed in. Optional.
	TransactionID immutable.Option[int]
}

var (
	_ Action   = (*ListActions)(nil)
	_ Stateful = (*ListActions)(nil)
)

func (a *ListActions) Execute() {
	nodeIDs, nodes := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		opts := options.ListActions()
		identOption := getIdentityForRequestSpecificToNode(a.s, a.Identity, nodeID)
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}

		// Check if a transaction is attached to this action. If so, we will be using it.
		var txn client.Txn
		var err error
		var info []client.ActionExecution
		if a.TransactionID.HasValue() {
			txn, err = a.s.GetTransaction(node, a.TransactionID)
			require.NoError(a.s.T, err)
			info, err = txn.ListActions(a.s.Ctx, opts)
		} else {
			info, err = node.ListActions(a.s.Ctx, opts)
		}

		require.NoError(a.s.T, err)

		require.Equal(a.s.T, len(a.ExpectedInfo), len(info),
			"expected %d actions, got %d", len(a.ExpectedInfo), len(info))

		for i, expectedAction := range a.ExpectedInfo {
			actualAction := info[i]

			require.Equal(a.s.T, expectedAction.CollectionID, actualAction.CollectionID,
				"collectionID mismatch for action index %v", i)
			require.Equal(a.s.T, expectedAction.Action, actualAction.Action,
				"action mismatch for action index %v", i)
			require.Equal(a.s.T, expectedAction.Subject, actualAction.Subject,
				"subject mismatch for action index %v", i)
			require.Equal(a.s.T, expectedAction.Status, actualAction.Status,
				"status mismatch for action index %v", i)
		}
	}
}
