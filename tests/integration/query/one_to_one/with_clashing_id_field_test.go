// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_one

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// This documents unwanted behaviour, see https://github.com/sourcenetwork/defradb/issues/1520
func TestQueryOneToOneWithClashingIdFieldOnSecondary(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation secondary direction, id field with name clash on secondary side",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Book {
						name: String
						author_id: Int
						author: Author
					}

					type Author {
						name: String
						published: Book @primary
					}
				`,
			},
			testUtils.CreateDoc{
				// bae-d82dbe47-9df1-5e33-bd87-f92e9c378161
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"author_id": 123456
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"published_id": "bae-d82dbe47-9df1-5e33-bd87-f92e9c378161"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book {
						name
						author_id
						author {
							name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"name":      "Painted House",
						"author_id": "bae-9d67a886-64e3-520b-8cd5-1ca7b098fabe",
						"author": map[string]any{
							"name": "John Grisham",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This documents unwanted behaviour, see https://github.com/sourcenetwork/defradb/issues/1520
func TestQueryOneToOneWithClashingIdFieldOnPrimary(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation primary direction, id field with name clash on primary side",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Book {
						name: String
						author_id: Int
						author: Author @primary
					}

					type Author {
						name: String
						published: Book
					}
				`,
			},
			testUtils.CreateDoc{
				// bae-d82dbe47-9df1-5e33-bd87-f92e9c378161
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"author_id": 123456
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"published_id": "bae-d82dbe47-9df1-5e33-bd87-f92e9c378161"
				}`,
				ExpectedError: "target document is already linked to another document.",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
