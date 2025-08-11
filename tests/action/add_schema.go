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
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
)

// AddSchema is an action that will add the given GQL schema to the Defra nodes.
type AddSchema struct {
	stateful

	// NodeID may hold the ID (index) of a node to apply this update to.
	//
	// If a value is not provided the update will be applied to all nodes.
	NodeID immutable.Option[int]

	// The schema to add.
	Schema string

	// Optionally, the expected results.
	//
	// Each item will be compared individually, if ID, RootID, SchemaVersionID or Fields on the
	// expected item are default they will not be compared with the actual.
	//
	// Assertions on Indexes and Sources will not distinguish between nil and empty (in order
	// to allow their ommission in most cases).
	ExpectedResults []client.CollectionVersion

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

var _ Action = (*AddSchema)(nil)
var _ Stateful = (*AddSchema)(nil)

func (a *AddSchema) Execute() {
	nodeIDs, nodes := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		schema := replace(a.s, nodeID, a.Schema)

		results, err := node.AddSchema(a.s.Ctx, schema)
		expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)

		assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)

		if a.ExpectedResults != nil {
			assertCollectionVersions(a.s, a.ExpectedResults, results)
		}
	}

	// If the schema was updated we need to refresh the collection definitions.
	refreshCollections(a.s)
}
