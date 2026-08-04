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

type PurgeDocs struct {
	stateful

	NodeID          immutable.Option[int]
	Identity        immutable.Option[state.Identity]
	CollectionIndex int
	DocIndexes      []int
	PruneHistory    bool
	ExpectedError   string
	TransactionID   immutable.Option[int]
}

var _ Action = (*PurgeDocs)(nil)
var _ Stateful = (*PurgeDocs)(nil)

func (a *PurgeDocs) Execute() {
	nodeIDs, nodes := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for index, node := range nodes {
		txnOption := immutable.None[client.Txn]()
		if a.TransactionID.HasValue() {
			txn, err := a.s.GetTransaction(node, a.TransactionID)
			require.NoError(a.s.T, err)
			txnOption = immutable.Some(txn)
		}

		collections, err := GetCollectionsCanonically(a.s, node, txnOption, a.Identity)
		if err != nil {
			expected := assertError(a.s.T, err, a.ExpectedError)
			assertExpectedErrorRaised(a.s.T, a.ExpectedError, expected)
			continue
		}

		docIDs := make([]client.DocID, len(a.DocIndexes))
		a.s.DocIDsLock.RLock()
		for i, docIndex := range a.DocIndexes {
			docIDs[i] = a.s.DocIDs[a.CollectionIndex][docIndex]
		}
		a.s.DocIDsLock.RUnlock()

		opts := options.PurgeByDocIDs()
		identity := getIdentityForRequestSpecificToNode(a.s, a.Identity, nodeIDs[index])
		if identity.HasValue() {
			opts.SetIdentity(identity.Value())
		}

		err = collections[a.CollectionIndex].PurgeByDocIDs(
			a.s.Ctx,
			docIDs,
			a.PruneHistory,
			opts,
		)
		expected := assertError(a.s.T, err, a.ExpectedError)
		assertExpectedErrorRaised(a.s.T, a.ExpectedError, expected)
	}
}
