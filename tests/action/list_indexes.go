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

	// ExpectedCollectionName is the collection name expected on every returned result.
	//
	// Only asserted when set.
	ExpectedCollectionName string

	// ExpectedStatuses maps index name to the expected execution for that index.
	// When set for a name, asserts Status, Action and Reason (partial match) instead of the
	// default ready assertion.
	// When not set for a name, asserts Status == CompletedActionStatus (ready).
	ExpectedStatuses map[string]client.ActionExecution

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

	nodeIDs, _ := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for _, nodeID := range nodeIDs {
		node := a.s.Nodes[nodeID]

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

		collections, err := GetCollectionsCanonically(a.s, node, txnOption, a.Identity)
		if err != nil {
			expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)
			assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)
			continue
		}
		collection := collections[a.CollectionID]

		opts := options.ListCollectionIndexes()
		identOption := getIdentityForRequestSpecificToNode(a.s, a.Identity, nodeID)
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}

		actualStatuses, err := collection.ListIndexes(a.s.Ctx, opts)

		expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)
		assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)
		if err != nil {
			continue
		}

		assertIndexesListsEqual(a.ExpectedIndexes, actualStatuses, a.s.T)
		assertIndexStatuses(a.ExpectedStatuses, actualStatuses, a.s.T)
		assertIndexCollectionNames(a.ExpectedCollectionName, actualStatuses, a.s.T)
	}
}

func assertIndexesListsEqual(
	expectedIndexes []client.IndexDescription,
	actualResults []client.ListIndexesResult,
	t require.TestingT,
) {
	toNamesFromExpected := func(indexes []client.IndexDescription) []string {
		names := make([]string, len(indexes))
		for i, index := range indexes {
			names[i] = index.Name
		}
		return names
	}

	toNamesFromActual := func(results []client.ListIndexesResult) []string {
		names := make([]string, len(results))
		for i, r := range results {
			names[i] = r.Description.Name
		}
		return names
	}

	require.ElementsMatch(t, toNamesFromExpected(expectedIndexes), toNamesFromActual(actualResults))

	actualMap := map[string]client.IndexDescription{}
	for _, r := range actualResults {
		actualMap[r.Description.Name] = r.Description
	}

	expectedMap := map[string]client.IndexDescription{}
	for _, index := range expectedIndexes {
		expectedMap[index.Name] = index
	}

	for key := range expectedMap {
		assertIndexesEqual(expectedMap[key], actualMap[key], t)
	}
}

// assertIndexStatuses checks Status, Action and Reason for each returned index.
//
// When expectedStatuses is nil, no status assertions are made.
// When expectedStatuses is non-nil:
//   - For names present in the map, asserts the given Status and Action, and that Reason
//     contains the expected substring.
//   - For all other names, asserts Status == CompletedActionStatus (ready).
func assertIndexStatuses(
	expectedStatuses map[string]client.ActionExecution,
	actualResults []client.ListIndexesResult,
	t require.TestingT,
) {
	if expectedStatuses == nil {
		return
	}
	actualByName := make(map[string]struct{}, len(actualResults))
	for _, actual := range actualResults {
		name := actual.Description.Name
		actualByName[name] = struct{}{}
		if expected, ok := expectedStatuses[name]; ok {
			assert.Equal(t, expected.Status, actual.Execution.Status, "index %s status mismatch", name)
			assert.Equal(t, expected.Action, actual.Execution.Action, "index %s action mismatch", name)
			if expected.Reason != "" {
				assert.Contains(t, actual.Execution.Reason, expected.Reason, "index %s reason mismatch", name)
			}
		} else {
			assert.Equal(t, client.CompletedActionStatus, actual.Execution.Status, "index %s expected ready status", name)
		}
	}
	// Every configured expectation must match a returned index, so a misspelled or missing
	// index name fails rather than silently passing.
	for name := range expectedStatuses {
		assert.Contains(t, actualByName, name, "expected status configured for missing index %s", name)
	}
}

// assertIndexCollectionNames checks that every returned result names its owning collection.
//
// When expectedName is empty, no assertions are made.
func assertIndexCollectionNames(
	expectedName string,
	actualResults []client.ListIndexesResult,
	t require.TestingT,
) {
	if expectedName == "" {
		return
	}
	for _, actual := range actualResults {
		assert.Equal(t, expectedName, actual.CollectionName,
			"index %s collection name mismatch", actual.Description.Name)
	}
}

func assertIndexesEqual(expectedIndex, actualIndex client.IndexDescription, t require.TestingT) {
	assert.Equal(t, expectedIndex.Name, actualIndex.Name, "index name mismatch")
	assert.Equal(t, expectedIndex.ID, actualIndex.ID, "index id mismatch")
	assert.Equal(t, expectedIndex.Unique, actualIndex.Unique, "index unique mismatch")
	assert.Equal(t, expectedIndex.Kind, actualIndex.Kind, "index kind mismatch")
	assert.Equal(t, expectedIndex.Options, actualIndex.Options, "index options mismatch")

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
