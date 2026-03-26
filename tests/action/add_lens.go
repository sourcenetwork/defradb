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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"
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

	// Used to identify the transaction for this to be executed in. Optional.
	TransactionID immutable.Option[int]
}

var _ Action = (*AddLens)(nil)
var _ Stateful = (*AddLens)(nil)

func (a *AddLens) Execute() {
	var lensID string

	nodeIDs, nodes := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		opts := options.AddLens()
		identOption := getIdentityForRequestSpecificToNode(a.s, a.Identity, nodeID)
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}

		// Check if a transaction is attached to this action. If so, we will be using it.
		var txn client.Txn
		var err error
		hadTxn := a.TransactionID.HasValue()
		if hadTxn {
			txn, err = a.s.GetTransaction(node, a.TransactionID)
			require.NoError(a.s.T, err)
			lensID, err = txn.AddLens(a.s.Ctx, a.Lens, opts)
		} else {
			lensID, err = node.AddLens(a.s.Ctx, a.Lens, opts)
		}

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
