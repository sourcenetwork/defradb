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

package index

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

// TestIndexCreate_WithExplicitTxn_BackfillRunsOnCommit verifies that when NewIndex is called
// inside an explicit transaction:
//   - The index is not yet visible outside the transaction before commit.
//   - After commit the index is present and all pre-existing documents are queryable through it.
func TestIndexCreate_WithExplicitTxn_BackfillRunsOnCommit(t *testing.T) {
	indexName := "user_name_idx"

	explainReq := makeExplainQuery(`query {
		User(filter: {name: {_eq: "Alice"}}) {
			name
		}
	}`)

	test := testUtils.TestCase{
		// LevelDB does not support concurrent transactions: this test holds an explicit
		// transaction open while the backfill runs on commit. Tracked by
		// https://github.com/sourcenetwork/defradb/issues/4959.
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age:  Int
					}
				`,
			},
			// Documents added before the transaction are committed data to backfill.
			&action.AddDoc{
				Doc: `{"name": "Alice", "age": 30}`,
			},
			&action.AddDoc{
				Doc: `{"name": "Bob", "age": 25}`,
			},
			&action.AddDoc{
				Doc: `{"name": "Carol", "age": 40}`,
			},
			&action.NewIndex{
				IndexName:     indexName,
				FieldName:     "name",
				TransactionID: immutable.Some(0),
			},
			// Before commit: index must not be visible outside the transaction.
			&action.ListIndexes{
				ExpectedIndexes: []client.IndexDescription{},
			},
			// Commit the transaction — backfill runs synchronously inside Commit.
			&action.CommitTransaction{
				TransactionID: 0,
			},
			// After commit: index must appear.
			&action.ListIndexes{
				ExpectedIndexes: []client.IndexDescription{
					{
						Name: indexName,
						ID:   1,
						Fields: []client.IndexedFieldDescription{
							{Name: "name"},
						},
					},
				},
			},
			&action.Request{
				Request: `query {
					User(filter: {name: {_eq: "Alice"}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Alice"},
					},
				},
			},
			// The planner must use the index (index fetches > 0).
			&action.Request{
				Request:  explainReq,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestIndexCreate_WithExplicitTxnDiscard_NoIndexNoBackfill verifies that when the transaction
// is discarded instead of committed:
//   - No index appears in ListIndexes.
//   - A filtered query still returns correct results (via full collection scan, 0 index fetches).
func TestIndexCreate_WithExplicitTxnDiscard_NoIndexNoBackfill(t *testing.T) {
	explainReq := makeExplainQuery(`query {
		User(filter: {name: {_eq: "Alice"}}) {
			name
		}
	}`)

	test := testUtils.TestCase{
		// LevelDB does not support concurrent transactions: this test holds an explicit
		// transaction open while the backfill runs on commit. Tracked by
		// https://github.com/sourcenetwork/defradb/issues/4959.
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age:  Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{"name": "Alice", "age": 30}`,
			},
			&action.AddDoc{
				Doc: `{"name": "Bob", "age": 25}`,
			},
			&action.NewIndex{
				IndexName:     "user_name_idx",
				FieldName:     "name",
				TransactionID: immutable.Some(0),
			},
			&action.DiscardTransaction{
				TransactionID: 0,
			},
			&action.ListIndexes{
				ExpectedIndexes: []client.IndexDescription{},
			},
			// The query still works via a full collection scan.
			&action.Request{
				Request: `query {
					User(filter: {name: {_eq: "Alice"}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Alice"},
					},
				},
			},
			// No index fetches because the index was discarded.
			&action.Request{
				Request:  explainReq,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
