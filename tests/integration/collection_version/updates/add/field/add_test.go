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

package field

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCollectionVersionUpdatesAddFieldWithAdd(t *testing.T) {
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
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			&action.Request{
				Request: `query {
					Users {
						_docID
						name
						email
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_docID": "bae-9b4d35b6-00f0-50df-8627-44cea1dbcf11",
							"name":   "John",
							"email":  nil,
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesAddFieldWithAddAfterCollectionUpdate(t *testing.T) {
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
			// We want to make sure that this works across database versions, so we tell
			// the change detector to split here.
			testUtils.SetupComplete{},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Shahzad",
					"email": "sqlizded@yahoo.ca"
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						email
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":  "John",
							"email": nil,
						},
						{
							"name":  "Shahzad",
							"email": "sqlizded@yahoo.ca",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// This test covers a bug that was found as part of https://github.com/sourcenetwork/defradb/issues/4707
// it only occurred when adding a field to a collection that already has a secondary relationship.
// The bug has been fixed, but the test remains as coverage of this case is important.
func TestCollectionVersionUpdatesAddField_WithExistingSecondaryOneToOneRelationship(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						publisher: Publisher @primary
					}

					type Publisher {
						book: Book
					}
				`,
			},
			&action.PatchCollection{
				Patch: `[{"op":"add","path":"/Publisher/Fields/-","value":{"Name":"name","Kind":"String"}}]`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name": "Penguin Books",
				},
			},
			&action.Request{
				Request: `query {
					Publisher {
						name
					}
				}`,
				Results: map[string]any{
					"Publisher": []map[string]any{
						{
							"name": "Penguin Books",
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
