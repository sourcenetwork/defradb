// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_many_multiple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToManyMultipleWithMultipleManyFilters(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Article {
						name: String
						author: Author
						rating: Int
					}

					type Book {
						name: String
						author: Author
						score: Int
					}

					type Author {
						name: String
						age: Int
						verified: Boolean
						books: [Book]
						articles: [Article]
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":     "John Grisham",
					"age":      65,
					"verified": true,
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":     "Cornelia Funke",
					"age":      62,
					"verified": false,
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "After Guantánamo, Another Injustice",
					"_authorID": testUtils.NewDocIndex(2, 0),
					"rating":    3,
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "To my dear readers",
					"_authorID": testUtils.NewDocIndex(2, 1),
					"rating":    2,
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Twinklestar's Favourite Xmas Cookie",
					"_authorID": testUtils.NewDocIndex(2, 1),
					"rating":    1,
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "Painted House",
					"_authorID": testUtils.NewDocIndex(2, 0),
					"score":     1,
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"_authorID": testUtils.NewDocIndex(2, 0),
					"score":     2,
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "Sooley",
					"_authorID": testUtils.NewDocIndex(2, 0),
					"score":     3,
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"_authorID": testUtils.NewDocIndex(2, 1),
					"score":     4,
				},
			},
			&action.Request{
				Request: `query {
					Author(filter: {name: {_eq: "John Grisham"}, books: {score: {_eq: 1}}, articles: {rating: {_eq: 3}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John Grisham",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
