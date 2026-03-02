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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/state"
)

// AddIndex will attempt to add the given secondary index for the given collection
// using the collection api.
type AddIndex struct {
	stateful

	// NodeID may hold the ID (index) of a node to add the secondary index on.
	//
	// If a value is not provided the index will be added on all nodes.
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

var _ Action = (*AddIndex)(nil)
var _ Stateful = (*AddIndex)(nil)

func (a *AddIndex) Execute() {
	fmt.Println("Entering AddIndex Execute")
	nodeIDs, _ := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for index, nodeID := range nodeIDs {

		node := a.s.Nodes[nodeID]

		// Check if a transaction is attached to this action. If so, we will be using it.
		var hadTxn bool
		var txn client.Txn
		if a.TransactionID.HasValue() {
			hadTxn = true
			txn, _ = a.s.GetTransaction(node, a.TransactionID)
		}

		nodeID := nodeIDs[index]
		var collections []client.Collection
		var err error
		if hadTxn {
			collections, err = txn.GetCollections(a.s.Ctx, options.GetCollections())
			if err != nil {
				return
			}
		} else {
			collections, err = node.GetCollections(a.s.Ctx, options.GetCollections())
			if err != nil {
				return
			}
		}

		collection := collections[a.CollectionID]
		fmt.Println("Collection that was found in AddIndex: ", collection.Name())

		indexDesc := client.IndexAddRequest{
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

		opts := options.CollectionAddIndex()
		identOption := getIdentityForRequestSpecificToNode(a.s, a.Identity, nodeID)
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}

		_, err = collection.AddIndex(a.s.Ctx, indexDesc, opts)

		expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)
		if expectedErrorRaised {
			return
		}
	}

	assertExpectedErrorRaised(a.s.T, a.ExpectedError, false)
}
