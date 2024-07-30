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

func TestQueryOneToManyWithSumWithLimitWithOffsetWithOrderAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from many side with sum with limit and offset and order",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Book {
						name: String
						rating: Float
						author: Author
					}

					type Author {
						name: String
						age: Int
						verified: Boolean
						published: [Book]
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
					"name":      "Painted House",
					"rating":    4.9,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"rating":    4.5,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Associate",
					"rating":    4.2,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Sooley",
					"rating":    3.2,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Rooster Bar",
					"rating":    4,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"rating":    4.8,
					"author_id": testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.Request{
				Request: `query {
					Author {
						name
					_sum(published: {field: rating, offset: 1, limit: 2, order: {name: ASC}})
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "Cornelia Funke",
							"_sum": float64(0),
						},
						{
							"name": "John Grisham",
							// 4.9 + 3.2
							// ...00001 is float math artifact
							"_sum": 8.100000000000001,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithSumWithLimitWithOffsetWithOrderDesc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from many side with sum with limit and offset and order",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Book {
						name: String
						rating: Float
						author: Author
					}

					type Author {
						name: String
						age: Int
						verified: Boolean
						published: [Book]
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
					"name":      "Painted House",
					"rating":    4.9,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"rating":    4.5,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Associate",
					"rating":    4.2,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Sooley",
					"rating":    3.2,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Rooster Bar",
					"rating":    4,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"rating":    4.8,
					"author_id": testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.Request{
				Request: `query {
					Author {
						name
						_sum(published: {field: rating, offset: 1, limit: 2, order: {name: DESC}})
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "Cornelia Funke",
							"_sum": float64(0),
						},
						{
							"name": "John Grisham",
							// 4.2 + 3.2
							"_sum": 7.4,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithSumWithLimitWithOffsetWithOrderAscAndDesc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from many side with sum with limit and offset and order",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Book {
						name: String
						rating: Float
						author: Author
					}

					type Author {
						name: String
						age: Int
						verified: Boolean
						published: [Book]
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
					"name":      "Painted House",
					"rating":    4.9,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"rating":    4.5,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Associate",
					"rating":    4.2,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Sooley",
					"rating":    3.2,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Rooster Bar",
					"rating":    4,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"rating":    4.8,
					"author_id": testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.Request{
				Request: `query {
					Author {
						name
						asc: _sum(published: {field: rating, offset: 1, limit: 2, order: {name: ASC}})
						desc: _sum(published: {field: rating, offset: 1, limit: 2, order: {name: DESC}})
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "Cornelia Funke",
							"asc":  float64(0),
							"desc": float64(0),
						},
						{
							"name": "John Grisham",
							// 4.9 + 3.2
							// ...00001 is float math artifact
							"asc": 8.100000000000001,
							// 4.2 + 3.2
							"desc": 7.4,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithSumWithLimitWithOffsetWithOrderOnDifferentFields(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from many side with sum with limit and offset and order",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Book {
						name: String
						rating: Float
						author: Author
					}

					type Author {
						name: String
						age: Int
						verified: Boolean
						published: [Book]
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
					"name":      "Painted House",
					"rating":    4.9,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"rating":    4.5,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Associate",
					"rating":    4.2,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Sooley",
					"rating":    3.2,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Rooster Bar",
					"rating":    4,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"rating":    4.8,
					"author_id": testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.Request{
				Request: `query {
					Author {
						name
						byName: _sum(published: {field: rating, offset: 1, limit: 2, order: {name: DESC}})
						byRating: _sum(published: {field: rating, offset: 1, limit: 2, order: {rating: DESC}})
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name":     "Cornelia Funke",
							"byName":   float64(0),
							"byRating": float64(0),
						},
						{
							"name": "John Grisham",
							// 4.2 + 3.2
							"byName": 7.4,
							// 4.5 + 4.2
							"byRating": 8.7,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithSumWithLimitWithOffsetWithOrderDescAndRenderedChildrenOrderedAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from many side with sum with limit and offset and order",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Book {
						name: String
						rating: Float
						author: Author
					}

					type Author {
						name: String
						age: Int
						verified: Boolean
						published: [Book]
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
					"name":      "Painted House",
					"rating":    4.9,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"rating":    4.5,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Associate",
					"rating":    4.2,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Sooley",
					"rating":    3.2,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Rooster Bar",
					"rating":    4,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"rating":    4.8,
					"author_id": testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.Request{
				Request: `query {
					Author {
						name
						_sum(published: {field: rating, offset: 1, limit: 2, order: {name: DESC}})
						published(offset: 1, limit: 2, order: {name: ASC}) {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name":      "Cornelia Funke",
							"_sum":      float64(0),
							"published": []map[string]any{},
						},
						{
							"name": "John Grisham",
							// 4.2 + 3.2
							"_sum": 7.4,
							"published": []map[string]any{
								{
									"name": "Painted House",
								},
								{
									"name": "Sooley",
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
