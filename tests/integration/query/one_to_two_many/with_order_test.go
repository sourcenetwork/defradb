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

package one_to_two_many

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToTwoManyWithOrder(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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

			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":          "Painted House",
					"rating":        4.9,
					"_authorID":     testUtils.NewDocIndex(1, 0),
					"_reviewedByID": testUtils.NewDocIndex(1, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":          "A Time for Mercy",
					"rating":        4.5,
					"_authorID":     testUtils.NewDocIndex(1, 0),
					"_reviewedByID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":          "Theif Lord",
					"rating":        4.8,
					"_authorID":     testUtils.NewDocIndex(1, 1),
					"_reviewedByID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.Request{
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
				Results: map[string]any{
					"Author": []map[string]any{
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
