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

// NewIndex will attempt to make a new secondary index for the given collection
// using the collection api.
type NewIndex struct {
	stateful

	// NodeID may hold the ID (index) of a node to make the new secondary index on.
	//
	// If a value is not provided the index will be made on all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// The collection for which this index should be added.
	CollectionID int

	// The name of the index to add. If not provided, one will be generated.
	IndexName string

	// The name of the field to index. Used only for single field indexes.
	// It's a convenience field so that tests don't have to add a slice
	// of [IndexedField] when only a single field index is needed.
	FieldName string

	// The fields to index. Used only for composite indexes.
	Fields []client.IndexedFieldDescription

	// If Unique is true, the index will be added as a unique index.
	Unique bool

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// Used to identify the transaction for this to be executed in. Optional.
	TransactionID immutable.Option[int]
}

var _ Action = (*NewIndex)(nil)
var _ Stateful = (*NewIndex)(nil)

func (a *NewIndex) Execute() {
	nodeIDs, _ := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for _, nodeID := range nodeIDs {
		node := a.s.Nodes[nodeID]

		// Check if a transaction is attached to this action. If so, we will be using it.
		var err error
		var txn client.Txn
		var collections []client.Collection
		if a.TransactionID.HasValue() {
			txn, err = a.s.GetTransaction(node, a.TransactionID)
			require.NoError(a.s.T, err)
			collections, err = txn.GetCollections(a.s.Ctx, options.GetCollections())
		} else {
			collections, err = node.GetCollections(a.s.Ctx, options.GetCollections())
		}

		if err != nil {
			return
		}

		collection := collections[a.CollectionID]

		indexDesc := client.NewIndexRequest{
			Name: a.IndexName,
		}

		if a.FieldName != "" {
			indexDesc.Fields = []client.IndexedFieldDescription{
				{
					Name: a.FieldName,
				},
			}
		} else if len(a.Fields) > 0 {
			for i := range a.Fields {
				indexDesc.Fields = append(indexDesc.Fields, client.IndexedFieldDescription{
					Name:       a.Fields[i].Name,
					Descending: a.Fields[i].Descending,
				})
			}
		}

		indexDesc.Unique = a.Unique

		opts := options.NewCollectionIndex()
		identOption := getIdentityForRequestSpecificToNode(a.s, a.Identity, nodeID)
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}

		_, err = collection.NewIndex(a.s.Ctx, indexDesc, opts)

		expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)
		if expectedErrorRaised {
			return
		}
	}

	assertExpectedErrorRaised(a.s.T, a.ExpectedError, false)
}
