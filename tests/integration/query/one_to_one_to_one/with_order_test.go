// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_one_to_one

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToOneToOneWithNestedOrder(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one-to-one relation primary direction",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Publisher {
						name: String
						printed: Book @primary
					}

					type Book {
						name: String
						publisher: Publisher
						author: Author @primary
					}

					type Author {
						name: String
						published: Book
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Old Publisher"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "New Publisher"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Painted House",
					"publisher_id": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Theif Lord",
					"publisher_id": testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":         "John Grisham",
					"published_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":         "Cornelia Funke",
					"published_id": testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.Request{
				Request: `query {
					Publisher(order: {printed: {author: {name: ASC}}}) {
						name
						printed {
							name
							author {
								name
							}
						}
					}
				}`,
				Results: []map[string]any{
					{
						"name": "New Publisher",
						"printed": map[string]any{
							"name": "Theif Lord",
							"author": map[string]any{
								"name": "Cornelia Funke",
							},
						},
					},
					{
						"name": "Old Publisher",
						"printed": map[string]any{
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
