// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package replace

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColDescrUpdateReplaceNameOneToMany_GivenExistingName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Author {
						name: String
						books: [Book]
					}

					type Book {
						name: String
						author: Author
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John Grisham",
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":   "Painted House",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "replace", "path": "/1/Name", "value": "Writer" }
					]
				`,
			},
			testUtils.Request{
				Request: `query {
					Book {
						name
						author {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name": "Painted House",
							"author": map[string]any{
								"name": "John Grisham",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColDescrUpdateReplaceNameOneToMany_GivenExistingNameReplacedBeforeAndAfterCreate(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Author {
						name: String
						books: [Book]
					}

					type Book {
						name: String
						author: Author
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John Grisham",
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":   "Painted House",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "replace", "path": "/1/Name", "value": "Writer" }
					]
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "Cornelia Funke",
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":   "Theif Lord",
					"author": testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.Request{
				Request: `query {
					Book {
						name
						author {
							name
						}
					}
				}`,
				// This test ensures that documents created before and after the collection rename
				// are correctly fetched together
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name": "Painted House",
							"author": map[string]any{
								"name": "John Grisham",
							},
						},
						{
							"name": "Theif Lord",
							"author": map[string]any{
								"name": "Cornelia Funke",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
