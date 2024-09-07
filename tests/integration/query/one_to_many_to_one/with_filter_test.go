// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_many_to_one

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryComplexWithDeepFilterOnRenderedChildren(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many-to-one deep filter on rendered children.",
		Actions: []any{
			gqlSchemaOneToManyToOne(),
			// Authors
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// Has written 1 Book
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// Has written no Book
				Doc: `{
					"name": "Not a Writer",
					"age": 6,
					"verified": false
				}`,
			},
			// Books
			testUtils.CreateDoc{
				CollectionID: 1,
				// Has 1 Publisher
				DocMap: map[string]any{
					"name":      "The Rooster Bar",
					"rating":    4,
					"author_id": testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// Has 1 Publisher
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"rating":    4.8,
					"author_id": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// Has no Publisher.
				DocMap: map[string]any{
					"name":      "The Associate",
					"rating":    4.2,
					"author_id": testUtils.NewDocIndex(0, 0),
				},
			},
			// Publishers
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":       "Only Publisher of The Rooster Bar",
					"address":    "1 Rooster Ave., Waterloo, Ontario",
					"yearOpened": 2022,
					"book_id":    testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":       "Only Publisher of Theif Lord",
					"address":    "1 Theif Lord, Waterloo, Ontario",
					"yearOpened": 2020,
					"book_id":    testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.Request{
				Request: `query {
					Author (filter: {book: {publisher: {yearOpened: {_gt: 2021}}}}) {
						name
						book {
							publisher {
								yearOpened
							}
						}
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "Cornelia Funke",
							"book": []map[string]any{
								{
									"publisher": map[string]any{
										"yearOpened": int64(2022),
									},
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

func TestOneToManyToOneWithSumOfDeepFilterSubTypeOfBothDescAndAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "1-N-1 sums of deep filter subtypes of both descending and ascending.",
		Actions: []any{
			gqlSchemaOneToManyToOne(),
			createDocsWith6BooksAnd5Publishers(),
			testUtils.Request{
				Request: `query {
					Author {
						name
						s1: _sum(book: {field: rating, filter: {publisher: {yearOpened: {_eq: 2013}}}})
						s2: _sum(book: {field: rating, filter: {publisher: {yearOpened: {_ge: 2020}}}})
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "Not a Writer",
							"s1":   0.0,
							"s2":   0.0,
						},
						{
							"name": "John Grisham",
							// 'Theif Lord' (4.8 rating) 2020, then 'A Time for Mercy' 2013 (4.5 rating).
							"s1": 4.5,
							// 'The Associate' as it has no Publisher (4.2 rating), then 'Painted House' 1995 (4.9 rating).
							"s2": 4.8,
						},
						{
							"name": "Cornelia Funke",
							"s1":   0.0,
							"s2":   4.0,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOneToManyToOneWithSumOfDeepFilterSubTypeAndDeepOrderBySubtypeOppositeDirections(t *testing.T) {
	test := testUtils.TestCase{
		Description: "1-N-1 sum of deep filter subtypes and non-sum deep filter",
		Actions: []any{
			gqlSchemaOneToManyToOne(),
			createDocsWith6BooksAnd5Publishers(),
			testUtils.Request{
				Request: `query {
					Author {
						name
						s1: _sum(book: {field: rating, filter: {publisher: {yearOpened: {_eq: 2013}}}})
						books2020: book(filter: {publisher: {yearOpened: {_ge: 2020}}}) {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name":      "Not a Writer",
							"s1":        0.0,
							"books2020": []map[string]any{},
						},
						{
							"name": "John Grisham",
							"s1":   4.5,
							"books2020": []map[string]any{
								{
									"name": "Theif Lord",
								},
							},
						},
						{
							"name": "Cornelia Funke",
							"s1":   0.0,
							"books2020": []map[string]any{
								{
									"name": "The Rooster Bar",
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

func TestOneToManyToOneWithTwoLevelDeepFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "1-N-1 two level deep filter",
		Actions: []any{
			gqlSchemaOneToManyToOne(),
			createDocsWith6BooksAnd5Publishers(),
			testUtils.Request{
				Request: `query {
					Author (filter: {book: {publisher: {yearOpened: { _ge: 2020}}}}){
						name
						book {
							name
							publisher {
								yearOpened
							}
						}
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"book": []map[string]any{
								{
									"name":      "The Associate",
									"publisher": nil,
								},
								{
									"name": "Theif Lord",
									"publisher": map[string]any{
										"yearOpened": int64(2020),
									},
								},
								{
									"name": "Painted House",
									"publisher": map[string]any{
										"yearOpened": int64(1995),
									},
								},
								{
									"name": "A Time for Mercy",
									"publisher": map[string]any{
										"yearOpened": int64(2013),
									},
								},
								{
									"name": "Sooley",
									"publisher": map[string]any{
										"yearOpened": int64(1999),
									},
								},
							},
							"name": "John Grisham",
						},
						{
							"book": []map[string]any{
								{
									"name": "The Rooster Bar",
									"publisher": map[string]any{
										"yearOpened": int64(2022),
									},
								},
							},
							"name": "Cornelia Funke",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOneToManyToOneWithCompoundOperatorInFilterAndRelation(t *testing.T) {
	test := testUtils.TestCase{
		Description: "1-N-1 two level deep filter with compound operator and relation",
		Actions: []any{
			gqlSchemaOneToManyToOne(),
			createDocsWith6BooksAnd5Publishers(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John Tolkien",
					"age": 70,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "The Lord of the Rings",
					"rating":    5.0,
					"author_id": testUtils.NewDocIndex(0, 3),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":       "Allen & Unwin",
					"address":    "1 Allen Ave., Sydney, Australia",
					"yearOpened": 1954,
					"book_id":    testUtils.NewDocIndex(1, 6),
				},
			},
			testUtils.Request{
				Request: `query {
					Author (filter: {_and: [
						{age: {_gt: 50}},
						{_or: [
							{book: {publisher: {yearOpened: {_gt: 2020}}}},
							{book: {publisher: {yearOpened: {_lt: 1960}}}}
						]}
					]}){
						name
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John Tolkien",
						},
						{
							"name": "Cornelia Funke",
						},
					},
				},
			},
			testUtils.Request{
				Request: `query {
					Author (filter: {_and: [
						{_not: {age: {_ge: 70}}},
						{book: {rating: {_gt: 2.5}}},
						{_or: [
							{book: {publisher: {yearOpened: {_le: 2020}}}},
							{_not: {book: {rating: {_le: 4.0}}}}
						]}
					]}){
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

func TestOneToManyToOneWithCompoundOperatorInSubFilterAndRelation(t *testing.T) {
	test := testUtils.TestCase{
		Description: "1-N-1 with sub filter with compound operator and relation",
		Actions: []any{
			gqlSchemaOneToManyToOne(),
			createDocsWith6BooksAnd5Publishers(),
			testUtils.Request{
				Request: `query {
					Author (filter: {_and: [
						{age: {_gt: 20}},
						{_or: [
							{book: {publisher: {yearOpened: {_lt: 2020}}}},
							{book: {rating: { _lt: 1}}}
						]}
					]}){
						name
						book (filter: {_and: [
							{publisher: {yearOpened: {_lt: 2020}}},
							{_or: [
								{rating: { _lt: 3.4}},
								{publisher: {name: {_eq: "Not existing publisher"}}}
							]}
						]}){
							name
						}
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John Grisham",
							"book": []map[string]any{{
								"name": "Sooley",
							}},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
