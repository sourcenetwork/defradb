// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_many

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// This documents unwanted behaviour, see https://github.com/sourcenetwork/defradb/issues/1520
func TestQueryOneToManyWithIdFieldOnPrimary(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation primary direction, id field with name clash on primary side",
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
						published: [Book]
					}
				`,
			},
			testUtils.CreateDoc{
				// bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"author_id": 123456
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"author_id": "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
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
						"name":      "A Time for Mercy",
						"author_id": "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed",
						"author": map[string]any{
							"name": "John Grisham",
						},
					},
					{
						"name":      "Painted House",
						"author_id": int64(123456),
						"author":    nil,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
