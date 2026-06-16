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
				ExpectedStatuses: map[string]client.IndexDescriptionStatus{
					"User_name_ASC": {
						Status: client.IndexStatusReady,
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
				ExpectedStatuses: map[string]client.IndexDescriptionStatus{
					"User_name_ASC": {
						Status: client.IndexStatusReady,
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
