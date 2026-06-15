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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// TestUniqueIndexBackfill_WithFailedIndex_QueryDoesFullScan verifies that a query
// filtered on a field whose unique index failed backfill returns correct results and
// that the planner performs a full scan rather than using the incomplete index.
func TestUniqueIndexBackfill_WithFailedIndex_QueryDoesFullScan(t *testing.T) {
	req := `query {
		User(filter: {age: {_eq: 21}}) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
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
				CollectionID: 0,
				Doc:          `{"name": "Alice", "age": 21}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob", "age": 21}`,
			},
			// Backfill fails: two documents share age 21.
			&action.NewIndex{
				CollectionID:  0,
				FieldName:     "age",
				Unique:        true,
				ExpectedError: "can not index a doc's field(s) that violates unique index",
			},
			// Both docs are returned correctly — the failed index is not used.
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Bob", "age": int64(21)},
						{"name": "Alice", "age": int64(21)},
					},
				},
			},
			// The explain must show 0 index fetches: the planner did a full scan.
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestUniqueIndexBackfill_WithDuplicateValues_FailsAndPersistsDefinition verifies that
// creating a unique index over a field that contains duplicate values:
//   - Returns an error.
//   - Leaves the index definition in place with a failed status (visible in ListIndexes).
//   - Allows the failed index to be removed with DeleteIndex.
//   - Allows a fresh unique index to be created successfully once the duplicate is resolved.
func TestUniqueIndexBackfill_WithDuplicateValues_FailsAndPersistsDefinition(t *testing.T) {
	test := testUtils.TestCase{
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
				CollectionID: 0,
				Doc:          `{"name": "Alice", "age": 21}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob", "age": 21}`,
			},
			// Backfill fails: two documents share age 21.
			&action.NewIndex{
				CollectionID:  0,
				FieldName:     "age",
				Unique:        true,
				ExpectedError: "can not index a doc's field(s) that violates unique index",
			},
			// The definition persists even though the backfill failed.
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "User_age_ASC",
						ID:     1,
						Unique: true,
						Fields: []client.IndexedFieldDescription{
							{Name: "age"},
						},
					},
				},
			},
			// Remove the failed index.
			&action.DeleteIndex{
				CollectionID: 0,
				IndexName:    "User_age_ASC",
			},
			&action.ListIndexes{
				CollectionID:    0,
				ExpectedIndexes: []client.IndexDescription{},
			},
			// Fix the duplicate: update doc 1 (Bob) to a different age.
			&action.UpdateDoc{
				CollectionID: 0,
				DocID:        1,
				Doc:          `{"age": 22}`,
			},
			// Now the unique index can be created successfully.
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "age",
				Unique:       true,
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						// The sequence advances even after a delete, so the new index gets ID 2.
						Name:   "User_age_ASC",
						ID:     2,
						Unique: true,
						Fields: []client.IndexedFieldDescription{
							{Name: "age"},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
