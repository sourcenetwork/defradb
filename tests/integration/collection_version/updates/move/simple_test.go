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

package move

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCollectionVersionUpdatesMoveCollectionDoesNothing(t *testing.T) {
	collectionVersionID := "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			&action.UpdateDoc{
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
