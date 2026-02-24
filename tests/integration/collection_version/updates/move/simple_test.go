// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package move

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesMoveCollectionDoesNothing(t *testing.T) {
	collectionVersionID := "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John"
				}`,
			},
			&action.PatchCollection{
				// This just moves an object to a new key in a temporary dictionary, it doesn't actually do
				// anything
				Patch: `
					[
						{ "op": "move", "from": "/Users", "path": "/Books" }
					]
				`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"name": "Johnnn"
				}`,
			},
			&action.Request{
				// Assert that Users is still Users
				Request: `query {
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Johnnn",
						},
					},
				},
			},
			&action.Request{
				// Assert that the version ID remains the same
				Request: `query {
					_commits (filter: {fieldName: {_eq: "_C"}}) {
						collectionVersionId
					}
				}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							// Update commit
							"collectionVersionId": collectionVersionID,
						},
						{
							// Create commit
							"collectionVersionId": collectionVersionID,
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
