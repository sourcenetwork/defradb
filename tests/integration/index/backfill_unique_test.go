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

// TestUniqueIndexBackfill_WithDuplicateValues_FailsAndPersistsDefinition verifies the full
// lifecycle of a unique index whose backfill fails on duplicate values:
//   - Returns an error and leaves the definition listed with a failed status + reason.
//   - A query on the field still returns correct results via a full scan (planner ignores it).
//   - The failed index can be removed with DeleteIndex.
//   - Once the duplicate is fixed, a fresh unique index builds to ready.
func TestUniqueIndexBackfill_WithDuplicateValues_FailsAndPersistsDefinition(t *testing.T) {
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
			// The definition persists even though the backfill failed, with a failed status.
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
				ExpectedStatuses: map[string]client.IndexDescriptionStatus{
					"User_age_ASC": {
						Status: client.ErroredActionStatus,
						Action: client.BackfillIndexAction,
						Reason: "can not index",
					},
				},
			},
			// The planner ignores the failed index: both docs come back via a full scan.
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Bob", "age": int64(21)},
						{"name": "Alice", "age": int64(21)},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
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
				ExpectedStatuses: map[string]client.IndexDescriptionStatus{
					"User_age_ASC": {
						Status: client.CompletedActionStatus,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
