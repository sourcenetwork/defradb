// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_two_many

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToTwoManyWithOrder(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from one side, order in opposite directions on children",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Book {
						name: String
						rating: Float
						author: Author @relation(name: "written_books")
						reviewedBy: Author @relation(name: "reviewed_books")
					}

					type Author {
						name: String
						age: Int
						verified: Boolean
						written: [Book] @relation(name: "written_books")
						reviewed: [Book] @relation(name: "reviewed_books")
					}
				`,
			},

			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":          "Painted House",
					"rating":        4.9,
					"author_id":     testUtils.NewDocIndex(1, 0),
					"reviewedBy_id": testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":          "A Time for Mercy",
					"rating":        4.5,
					"author_id":     testUtils.NewDocIndex(1, 0),
					"reviewedBy_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":          "Theif Lord",
					"rating":        4.8,
					"author_id":     testUtils.NewDocIndex(1, 1),
					"reviewedBy_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.Request{
				Request: `query {
					Author {
						name
						written (order: {rating: ASC}) {
							name
						}
						reviewed (order: {rating: DESC}){
							name
							rating
						}
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Cornelia Funke",
						"reviewed": []map[string]any{
							{
								"name":   "Painted House",
								"rating": 4.9,
							},
						},
						"written": []map[string]any{
							{
								"name": "Theif Lord",
							},
						},
					},
					{
						"name": "John Grisham",
						"reviewed": []map[string]any{
							{
								"name":   "Theif Lord",
								"rating": 4.8,
							},
							{
								"name":   "A Time for Mercy",
								"rating": 4.5,
							},
						},
						"written": []map[string]any{
							{
								"name": "A Time for Mercy",
							},
							{
								"name": "Painted House",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
