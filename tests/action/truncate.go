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

type Truncate struct {
	stateful

	// NodeID may hold the ID (index) of a node to truncate.
	//
	// If a value is not provided all nodes will be truncated.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// CollectionIndex is the index of the collection to truncate.
	CollectionIndex int

	// Filter limits the truncate to matching documents.
	Filter any

	// DocIndexes builds a DocID filter from documents in the test state.
	DocIndexes []int

	// PruneHistory removes unshared history for matching documents.
	PruneHistory bool

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// TransactionID to use for the action. Optional.
	TransactionID immutable.Option[int]
}

var _ Action = (*Truncate)(nil)
var _ Stateful = (*Truncate)(nil)

func (a *Truncate) Execute() {
	hadTxn := a.TransactionID.HasValue()

	nodeIDs, nodes := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for index, node := range nodes {
		// Check if a transaction is attached to this action. If so, we will be using it.
		var txn client.Txn
		txnOption := immutable.None[client.Txn]()
		if hadTxn {
			var err error
			txn, err = a.s.GetTransaction(node, a.TransactionID)
			require.NoError(a.s.T, err)
			txnOption = immutable.Some(txn)
		}

		nodeID := nodeIDs[index]

		collections, err := GetCollectionsCanonically(a.s, node, txnOption, a.Identity)
		if err != nil {
			expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)
			assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)
			continue
		}
		collection := collections[a.CollectionIndex]

		opts := options.TruncateCollection()
		filter := a.Filter
		if len(a.DocIndexes) > 0 {
			values := make([]any, len(a.DocIndexes))
			a.s.DocIDsLock.RLock()
			for i, docIndex := range a.DocIndexes {
				values[i] = a.s.DocIDs[a.CollectionIndex][docIndex].String()
			}
			a.s.DocIDsLock.RUnlock()
			filter = map[string]any{"_docID": map[string]any{"_in": values}}
		}
		if filter != nil {
			opts.SetFilter(filter).SetPruneHistory(a.PruneHistory)
		}
		identOption := getIdentityForRequestSpecificToNode(a.s, a.Identity, nodeID)
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}
		err = collection.Truncate(a.s.Ctx, opts)

		expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)
		assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)
	}
}
