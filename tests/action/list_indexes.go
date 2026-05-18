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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/state"
)

// ListIndexes will attempt to list the indexes from the given collection
// using the collection api.
type ListIndexes struct {
	stateful

	// NodeID may hold the ID (index) of a node to list the indexes from.
	//
	// If a value is not provided the indexes will be retrieved from the first node.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// The collection for which the indexes should be retrieved.
	CollectionID int

	// The expected indexes to be returned.
	ExpectedIndexes []client.IndexDescription

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// Used to identify the transaction for this to be executed in. Optional.
	TransactionID immutable.Option[int]
}

var _ Action = (*ListIndexes)(nil)
var _ Stateful = (*ListIndexes)(nil)

func (a *ListIndexes) Execute() {
	if len(a.s.Nodes) == 0 {
		return
	}

	var expectedErrorRaised bool

	nodeIDs, _ := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for index, nodeID := range nodeIDs {
		node := a.s.Nodes[index]

		// Check if a transaction is attached to this action. If so, we will be using it.
		var txn client.Txn
		var err error
		txnOption := immutable.None[client.Txn]()
		hadTxn := a.TransactionID.HasValue()
		if hadTxn {
			txn, err = a.s.GetTransaction(node, a.TransactionID)
			require.NoError(a.s.T, err)
			txnOption = immutable.Some(txn)
		}

		collections := MustGetCanonicallyOrderedCollections(a.s, node, txnOption)
		collection := collections[a.CollectionID]

		opts := options.ListCollectionIndexes()
		identOption := getIdentityForRequestSpecificToNode(a.s, a.Identity, nodeID)
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}

		actualIndexes, err := collection.ListIndexes(a.s.Ctx, opts)

		if assertError(a.s.T, err, a.ExpectedError) {
			expectedErrorRaised = true
			continue
		}

		assertIndexesListsEqual(a.ExpectedIndexes, actualIndexes, a.s.T)
	}

	assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)
}

func assertIndexesListsEqual(
	expectedIndexes []client.IndexDescription,
	actualIndexes []client.IndexDescription,
	t require.TestingT,
) {
	toNames := func(indexes []client.IndexDescription) []string {
		names := make([]string, len(indexes))
		for i, index := range indexes {
			names[i] = index.Name
		}
		return names
	}

	require.ElementsMatch(t, toNames(expectedIndexes), toNames(actualIndexes))

	toMap := func(indexes []client.IndexDescription) map[string]client.IndexDescription {
		resultMap := map[string]client.IndexDescription{}
		for _, index := range indexes {
			resultMap[index.Name] = index
		}
		return resultMap
	}

	expectedMap := toMap(expectedIndexes)
	actualMap := toMap(actualIndexes)
	for key := range expectedMap {
		assertIndexesEqual(expectedMap[key], actualMap[key], t)
	}
}

func assertIndexesEqual(expectedIndex, actualIndex client.IndexDescription, t require.TestingT) {
	assert.Equal(t, expectedIndex.Name, actualIndex.Name, "index name mismatch")
	assert.Equal(t, expectedIndex.ID, actualIndex.ID, "index id mismatch")
	assert.Equal(t, expectedIndex.Unique, actualIndex.Unique, "index unique mismatch")

	toNames := func(fields []client.IndexedFieldDescription) []string {
		names := make([]string, len(fields))
		for i, field := range fields {
			names[i] = field.Name
		}
		return names
	}

	require.ElementsMatch(t, toNames(expectedIndex.Fields), toNames(actualIndex.Fields), "index fields' names mismatch")

	toMap := func(fields []client.IndexedFieldDescription) map[string]client.IndexedFieldDescription {
		resultMap := map[string]client.IndexedFieldDescription{}
		for _, field := range fields {
			resultMap[field.Name] = field
		}
		return resultMap
	}

	expectedFieldsMap := toMap(expectedIndex.Fields)
	actualFieldsMap := toMap(actualIndex.Fields)
	for fieldName := range expectedFieldsMap {
		assert.Equal(t, expectedFieldsMap[fieldName].Descending, actualFieldsMap[fieldName].Descending,
			"index field %s descending mismatch", fieldName)
	}
}
