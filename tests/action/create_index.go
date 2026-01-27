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
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/state"
)

// CreateIndex will attempt to create the given secondary index for the given collection
// using the collection api.
type CreateIndex struct {
	stateful

	// NodeID may hold the ID (index) of a node to create the secondary index on.
	//
	// If a value is not provided the index will be created in all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// The collection for which this index should be created.
	CollectionID int

	// The name of the index to create. If not provided, one will be generated.
	IndexName string

	// The name of the field to index. Used only for single field indexes.
	// It's a convenience field so that tests don't have to create a slice
	// of [IndexedField] when only a single field index is needed.
	FieldName string

	// The fields to index. Used only for composite indexes.
	Fields []client.IndexedFieldDescription

	// If Unique is true, the index will be created as a unique index.
	Unique bool

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

var _ Action = (*CreateIndex)(nil)
var _ Stateful = (*CreateIndex)(nil)

func (a *CreateIndex) Execute() {
	nodeIDs, _ := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for _, nodeID := range nodeIDs {
		collection := a.s.Nodes[nodeID].Collections[a.CollectionID]

		indexDesc := client.IndexCreateRequest{
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

		ctx := getContextWithIdentity(a.s.Ctx, a.s, a.Identity, nodeID)
		_, err := collection.CreateIndex(ctx, indexDesc)

		expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)
		if expectedErrorRaised {
			return
		}
	}

	assertExpectedErrorRaised(a.s.T, a.ExpectedError, false)
}
