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
)

// DeleteCollection deletes one or more collections by name.
//
// All named collections are removed atomically in a single operation. This can be used
// to delete collections that reference each other via relations, since removing them
// one at a time would leave a dangling reference and be rolled back.
//
// By default, every version of each named collection is deleted (active head and all
// earlier versions). Set ActiveOnly to true to delete only the latest (head) version
// and keep earlier versions intact.
//
// The collections must not contain any documents - they must be deleted first.
type DeleteCollection struct {
	stateful

	// NodeID may hold the ID (index) of a node to apply this delete to.
	//
	// If a value is not provided the delete will be applied to all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// Names is the list of collection names to delete atomically.
	//
	// A single-name delete is just a one-element slice; the underlying API handles
	// one or many names uniformly.
	Names []string

	// ActiveOnly limits the delete to only the active head version of each named
	// collection. When false (the default) every version of each named collection
	// is removed.
	ActiveOnly bool

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// Used to identify the transaction for this to be executed in. Optional.
	TransactionID immutable.Option[int]
}

var _ Action = (*DeleteCollection)(nil)
var _ Stateful = (*DeleteCollection)(nil)

// Execute executes the delete collection action.
func (a *DeleteCollection) Execute() {
	nodeIDs, nodes := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		opts := options.DeleteCollection()
		identOption := getIdentityForRequestSpecificToNode(a.s, a.Identity, nodeID)
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}
		opts.SetActiveOnly(a.ActiveOnly)

		var txn client.Txn
		var err error
		if a.TransactionID.HasValue() {
			txn, err = a.s.GetTransaction(node, a.TransactionID)
			require.NoError(a.s.T, err)
			err = txn.DeleteCollection(a.s.Ctx, a.Names, opts)
		} else {
			err = node.DeleteCollection(a.s.Ctx, a.Names, opts)
		}

		expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)

		assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)
	}

	if !a.TransactionID.HasValue() {
		RefreshCollections(a.s)
	}
}
