// Copyright 2022 Democratized Data Foundation
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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToManyMultipleWithCount(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from many side with count",
		Actions: []any{
			testUtils.SchemaUpdate{
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
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":     "John Grisham",
					"age":      65,
					"verified": true,
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":     "Cornelia Funke",
					"age":      62,
					"verified": false,
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "After Guantánamo, Another Injustice",
					"author_id": testUtils.NewDocIndex(2, 0),
					"rating":    3,
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "To my dear readers",
					"author_id": testUtils.NewDocIndex(2, 1),
					"rating":    2,
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Twinklestar's Favourite Xmas Cookie",
					"author_id": testUtils.NewDocIndex(2, 1),
					"rating":    1,
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "Painted House",
					"author_id": testUtils.NewDocIndex(2, 0),
					"score":     1,
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"author_id": testUtils.NewDocIndex(2, 0),
					"score":     2,
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"author_id": testUtils.NewDocIndex(2, 1),
					"score":     4,
				},
			},
			testUtils.Request{
				Request: `query {
						Author {
							name
							numberOfBooks: _count(books: {})
							numberOfArticles: _count(articles: {})
						}
					}`,
				Results: []map[string]any{
					{
						"name":             "Cornelia Funke",
						"numberOfBooks":    1,
						"numberOfArticles": 2,
					},
					{
						"name":             "John Grisham",
						"numberOfBooks":    2,
						"numberOfArticles": 1,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyMultipleWithCountOnMultipleJoins(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from many side with count",
		Actions: []any{
			testUtils.SchemaUpdate{
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
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":     "John Grisham",
					"age":      65,
					"verified": true,
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":     "Cornelia Funke",
					"age":      62,
					"verified": false,
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "After Guantánamo, Another Injustice",
					"author_id": testUtils.NewDocIndex(2, 0),
					"rating":    3,
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "To my dear readers",
					"author_id": testUtils.NewDocIndex(2, 1),
					"rating":    2,
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Twinklestar's Favourite Xmas Cookie",
					"author_id": testUtils.NewDocIndex(2, 1),
					"rating":    1,
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "Painted House",
					"author_id": testUtils.NewDocIndex(2, 0),
					"score":     1,
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"author_id": testUtils.NewDocIndex(2, 0),
					"score":     2,
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "Sooley",
					"author_id": testUtils.NewDocIndex(2, 0),
					"score":     3,
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"author_id": testUtils.NewDocIndex(2, 1),
					"score":     4,
				},
			},
			testUtils.Request{
				Request: `query {
						Author {
							name
							_count(books: {}, articles: {})
						}
					}`,
				Results: []map[string]any{
					{
						"name":   "Cornelia Funke",
						"_count": 3,
					},
					{
						"name":   "John Grisham",
						"_count": 4,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
