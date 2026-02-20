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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToManyWithSumWithLimitWithOffsetWithOrderAsc(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
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
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Painted House",
					"rating":    4.9,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"rating":    4.5,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Associate",
					"rating":    4.2,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Sooley",
					"rating":    3.2,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Rooster Bar",
					"rating":    4,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"rating":    4.8,
					"_authorID": testUtils.NewDocIndex(1, 1),
				},
			},
			&action.Request{
				Request: `query {
					Author {
						name
					SUM(published: {field: rating, offset: 1, limit: 2, order: {name: ASC}})
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "Cornelia Funke",
							"SUM":  float64(0),
						},
						{
							"name": "John Grisham",
							// 4.9 + 3.2
							// ...00001 is float math artifact
							"SUM": 8.100000000000001,
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
		Actions: []any{
			&action.AddSchema{
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
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Painted House",
					"rating":    4.9,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"rating":    4.5,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Associate",
					"rating":    4.2,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Sooley",
					"rating":    3.2,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Rooster Bar",
					"rating":    4,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"rating":    4.8,
					"_authorID": testUtils.NewDocIndex(1, 1),
				},
			},
			&action.Request{
				Request: `query {
					Author {
						name
						SUM(published: {field: rating, offset: 1, limit: 2, order: {name: DESC}})
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "Cornelia Funke",
							"SUM":  float64(0),
						},
						{
							"name": "John Grisham",
							// 4.2 + 3.2
							"SUM": 7.4,
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
		Actions: []any{
			&action.AddSchema{
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
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Painted House",
					"rating":    4.9,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"rating":    4.5,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Associate",
					"rating":    4.2,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Sooley",
					"rating":    3.2,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Rooster Bar",
					"rating":    4,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"rating":    4.8,
					"_authorID": testUtils.NewDocIndex(1, 1),
				},
			},
			&action.Request{
				Request: `query {
					Author {
						name
						asc: SUM(published: {field: rating, offset: 1, limit: 2, order: {name: ASC}})
						desc: SUM(published: {field: rating, offset: 1, limit: 2, order: {name: DESC}})
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
		Actions: []any{
			&action.AddSchema{
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
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Painted House",
					"rating":    4.9,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"rating":    4.5,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Associate",
					"rating":    4.2,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Sooley",
					"rating":    3.2,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Rooster Bar",
					"rating":    4,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"rating":    4.8,
					"_authorID": testUtils.NewDocIndex(1, 1),
				},
			},
			&action.Request{
				Request: `query {
					Author {
						name
						byName: SUM(published: {field: rating, offset: 1, limit: 2, order: {name: DESC}})
						byRating: SUM(published: {field: rating, offset: 1, limit: 2, order: {rating: DESC}})
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
		Actions: []any{
			&action.AddSchema{
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
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Painted House",
					"rating":    4.9,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"rating":    4.5,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Associate",
					"rating":    4.2,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Sooley",
					"rating":    3.2,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Rooster Bar",
					"rating":    4,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"rating":    4.8,
					"_authorID": testUtils.NewDocIndex(1, 1),
				},
			},
			&action.Request{
				Request: `query {
					Author {
						name
						SUM(published: {field: rating, offset: 1, limit: 2, order: {name: DESC}})
						published(offset: 1, limit: 2, order: {name: ASC}) {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name":      "Cornelia Funke",
							"SUM":       float64(0),
							"published": []map[string]any{},
						},
						{
							"name": "John Grisham",
							// 4.2 + 3.2
							"SUM": 7.4,
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
