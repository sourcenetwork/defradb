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

// TestIndexStatus_NewIndex_IsReady verifies that a freshly created index is listed as ready.
func TestIndexStatus_NewIndex_IsReady(t *testing.T) {
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
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "name",
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "User_name_ASC",
						ID:     1,
						Fields: []client.IndexedFieldDescription{{Name: "name"}},
					},
				},
				ExpectedStatuses: map[string]client.ActionExecution{
					"User_name_ASC": {
						Status: client.CompletedActionStatus,
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// TestIndexStatus_ReadyAndFailedInOneCollection_EachReportsOwnStatus verifies that a listing
// of a collection holding both a ready index (on "name") and a failed index (on "age", from
// duplicate values) reports each index's own status.
func TestIndexStatus_ReadyAndFailedInOneCollection_EachReportsOwnStatus(t *testing.T) {
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
			// Add two docs with the same age so that a unique index on age will fail.
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice", "age": 21}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob", "age": 21}`,
			},
			// This index succeeds and is ready.
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "name",
			},
			// This unique index fails because of the duplicate age values.
			&action.NewIndex{
				CollectionID:  0,
				FieldName:     "age",
				Unique:        true,
				ExpectedError: "can not index a doc's field(s) that violates unique index",
			},
			// ListIndexes must return both indexes: name=ready, age=failed.
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "User_name_ASC",
						ID:     1,
						Fields: []client.IndexedFieldDescription{{Name: "name"}},
					},
					{
						Name:   "User_age_ASC",
						ID:     2,
						Unique: true,
						Fields: []client.IndexedFieldDescription{{Name: "age"}},
					},
				},
				ExpectedStatuses: map[string]client.ActionExecution{
					"User_name_ASC": {
						Status: client.CompletedActionStatus,
					},
					"User_age_ASC": {
						Status: client.ErroredActionStatus,
						Action: client.BackfillIndexAction,
						Reason: "can not index",
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// TestIndexStatus_SDLIndex_IsReady verifies that an index declared inline in the SDL
// (the @index path, which writes no state record) lists as ready.
// A missing state record must be treated as ready.
func TestIndexStatus_SDLIndex_IsReady(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String @index
						age:  Int
					}
				`,
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "User_name_ASC",
						ID:     1,
						Fields: []client.IndexedFieldDescription{{Name: "name"}},
					},
				},
				ExpectedStatuses: map[string]client.ActionExecution{
					"User_name_ASC": {
						Status: client.CompletedActionStatus,
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
