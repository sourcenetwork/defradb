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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToTwoManyWithNilUnnamedRelationship_FromOneSide(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
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
					"_reviewedByID": testUtils.NewDocIndex(1, 1),
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
					Book {
						name
						rating
						author {
							name
						}
						reviewedBy {
							name
							age
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":   "Painted House",
							"rating": 4.9,
							"author": map[string]any{
								"name": "John Grisham",
							},
							"reviewedBy": map[string]any{
								"name": "Cornelia Funke",
								"age":  int64(62),
							},
						},
						{
							"name":   "Theif Lord",
							"rating": 4.8,
							"author": map[string]any{
								"name": "Cornelia Funke",
							},
							"reviewedBy": map[string]any{
								"name": "John Grisham",
								"age":  int64(65),
							},
						},
						{
							"name":   "A Time for Mercy",
							"rating": 4.5,
							"author": map[string]any{
								"name": "John Grisham",
							},
							"reviewedBy": map[string]any{
								"name": "Cornelia Funke",
								"age":  int64(62),
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToTwoManyWithNilUnnamedRelationship_FromManySide(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
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
					"_reviewedByID": testUtils.NewDocIndex(1, 1),
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
						age
						written {
							name
						}
						reviewed {
							name
							rating
						}
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John Grisham",
							"age":  int64(65),
							"reviewed": []map[string]any{
								{
									"name":   "Theif Lord",
									"rating": 4.8,
								},
							},
							"written": []map[string]any{
								{
									"name": "Painted House",
								},
								{
									"name": "A Time for Mercy",
								},
							},
						},
						{
							"name": "Cornelia Funke",
							"age":  int64(62),
							"reviewed": []map[string]any{
								{
									"name":   "Painted House",
									"rating": 4.9,
								},
								{
									"name":   "A Time for Mercy",
									"rating": 4.5,
								},
							},
							"written": []map[string]any{
								{
									"name": "Theif Lord",
								},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToTwoManyWithNamedAndUnnamedRelationships(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Book {
						name: String
						rating: Float
						price: Price
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

					type Price {
						currency: String
						value: Float
						books: [Book]
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 2,
				Doc: `{
						"currency": "GBP",
						"value": 12.99
					}`,
			},
			&action.AddDoc{
				CollectionID: 2,
				Doc: `{
						"currency": "SEK",
						"value": 129
					}`,
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
					"_priceID":      testUtils.NewDocIndex(2, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":          "A Time for Mercy",
					"rating":        4.5,
					"_authorID":     testUtils.NewDocIndex(1, 0),
					"_reviewedByID": testUtils.NewDocIndex(1, 1),
					"_priceID":      testUtils.NewDocIndex(2, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":          "Theif Lord",
					"rating":        4.8,
					"_authorID":     testUtils.NewDocIndex(1, 1),
					"_reviewedByID": testUtils.NewDocIndex(1, 0),
					"_priceID":      testUtils.NewDocIndex(2, 0),
				},
			},
			&action.Request{
				Request: `query {
					Book {
						name
						rating
						author {
							name
						}
						reviewedBy {
							name
							age
						}
						price {
							currency
							value
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":   "A Time for Mercy",
							"rating": 4.5,
							"author": map[string]any{
								"name": "John Grisham",
							},
							"reviewedBy": map[string]any{
								"name": "Cornelia Funke",
								"age":  int64(62),
							},
							"price": map[string]any{
								"currency": "SEK",
								"value":    float64(129),
							},
						},
						{
							"name":   "Theif Lord",
							"rating": 4.8,
							"author": map[string]any{
								"name": "Cornelia Funke",
							},
							"reviewedBy": map[string]any{
								"name": "John Grisham",
								"age":  int64(65),
							},
							"price": map[string]any{
								"currency": "GBP",
								"value":    12.99,
							},
						},
						{
							"name":   "Painted House",
							"rating": 4.9,
							"author": map[string]any{
								"name": "John Grisham",
							},
							"reviewedBy": map[string]any{
								"name": "Cornelia Funke",
								"age":  int64(62),
							},
							"price": map[string]any{
								"currency": "GBP",
								"value":    12.99,
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToTwoManyWithNamedAndUnnamedRelationships_FromManySide(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
						type Book {
							name: String
							rating: Float
							price: Price
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

						type Price {
							currency: String
							value: Float
							books: [Book]
						}
					`,
			},
			&action.AddDoc{
				CollectionID: 2,
				Doc: `{
						"currency": "GBP",
						"value": 12.99
					}`,
			},
			&action.AddDoc{
				CollectionID: 2,
				Doc: `{
						"currency": "SEK",
						"value": 129
					}`,
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
					"_priceID":      testUtils.NewDocIndex(2, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":          "A Time for Mercy",
					"rating":        4.5,
					"_authorID":     testUtils.NewDocIndex(1, 0),
					"_reviewedByID": testUtils.NewDocIndex(1, 1),
					"_priceID":      testUtils.NewDocIndex(2, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":          "Theif Lord",
					"rating":        4.8,
					"_authorID":     testUtils.NewDocIndex(1, 1),
					"_reviewedByID": testUtils.NewDocIndex(1, 0),
					"_priceID":      testUtils.NewDocIndex(2, 0),
				},
			},
			&action.Request{
				Request: `query {
					Author {
						name
						age
						written {
							name
							price {
								value
							}
						}
						reviewed {
							name
							rating
						}
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John Grisham",
							"age":  int64(65),
							"reviewed": []map[string]any{
								{
									"name":   "Theif Lord",
									"rating": 4.8,
								},
							},
							"written": []map[string]any{
								{
									"name": "A Time for Mercy",
									"price": map[string]any{
										"value": float64(129),
									},
								},
								{
									"name": "Painted House",
									"price": map[string]any{
										"value": 12.99,
									},
								},
							},
						},
						{
							"name": "Cornelia Funke",
							"age":  int64(62),
							"reviewed": []map[string]any{
								{
									"name":   "A Time for Mercy",
									"rating": 4.5,
								},
								{
									"name":   "Painted House",
									"rating": 4.9,
								},
							},
							"written": []map[string]any{
								{
									"name": "Theif Lord",
									"price": map[string]any{
										"value": 12.99,
									},
								},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
