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

func TestOneToManyToOneWithSumOfDeepOrderBySubTypeAndDeepOrderBySubtypeDescDirections(t *testing.T) {
	test := testUtils.TestCase{
		Description: "1-N-1 sum of deep orderby subtypes and non-sum deep orderby, desc. directions.",
		Actions: []any{
			gqlSchemaOneToManyToOne(),
			createDocsWith6BooksAnd5Publishers(),
			testUtils.Request{
				Request: `query {
					Author {
						name
						s1: _sum(book: {field: rating, order: {publisher: {yearOpened: DESC}}, limit: 2})
						NewestPublishersBook: book(order: {publisher: {yearOpened: DESC}}, limit: 2) {
							name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John Grisham",
						"s1":   4.8 + 4.5, // Because in descending order years for John are [2020, 2013].
						"NewestPublishersBook": []map[string]any{
							{
								"name": "Theif Lord",
							},
							{
								"name": "A Time for Mercy",
							},
						},
					},
					{
						"name":                 "Not a Writer",
						"s1":                   0.0,
						"NewestPublishersBook": []map[string]any{},
					},
					{
						"name": "Cornelia Funke",
						"s1":   4.0,
						"NewestPublishersBook": []map[string]any{
							{
								"name": "The Rooster Bar",
							},
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTEMP(t, test)
}

func TestOneToManyToOneWithSumOfDeepOrderBySubTypeAndDeepOrderBySubtypeAscDirections(t *testing.T) {
	test := testUtils.TestCase{
		Description: "1-N-1 sum of deep orderby subtypes and non-sum deep orderby, asc. directions.",
		Actions: []any{
			gqlSchemaOneToManyToOne(),
			createDocsWith6BooksAnd5Publishers(),
			testUtils.Request{
				Request: `query {
					Author {
						name
						s1: _sum(book: {field: rating, order: {publisher: {yearOpened: ASC}}, limit: 2})
						NewestPublishersBook: book(order: {publisher: {yearOpened: ASC}}, limit: 2) {
							name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John Grisham",
						// Because in ascending order years for John are:
						// 'The Associate' as it has no Publisher (4.2 rating), then 'Painted House' 1995 (4.9 rating).
						"s1": float64(4.2) + float64(4.9),
						"NewestPublishersBook": []map[string]any{
							{
								"name": "The Associate",
							},
							{
								"name": "Painted House",
							},
						},
					},
					{
						"name":                 "Not a Writer",
						"s1":                   0.0,
						"NewestPublishersBook": []map[string]any{},
					},
					{
						"name": "Cornelia Funke",
						"s1":   4.0,
						"NewestPublishersBook": []map[string]any{
							{
								"name": "The Rooster Bar",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTEMP(t, test)
}

func TestOneToManyToOneWithSumOfDeepOrderBySubTypeOfBothDescAndAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "1-N-1 sums of deep orderby subtypes of both descending and ascending.",
		Actions: []any{
			gqlSchemaOneToManyToOne(),
			createDocsWith6BooksAnd5Publishers(),
			testUtils.Request{
				Request: `query {
					Author {
						name
						s1: _sum(book: {field: rating, order: {publisher: {yearOpened: DESC}}, limit: 2})
						s2: _sum(book: {field: rating, order: {publisher: {yearOpened: ASC}}, limit: 2})
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John Grisham",
						// 'Theif Lord' (4.8 rating) 2020, then 'A Time for Mercy' 2013 (4.5 rating).
						"s1": 4.8 + 4.5,
						// 'The Associate' as it has no Publisher (4.2 rating), then 'Painted House' 1995 (4.9 rating).
						"s2": float64(4.2) + float64(4.9),
					},
					{
						"name": "Not a Writer",
						"s1":   0.0,
						"s2":   0.0,
					},
					{
						"name": "Cornelia Funke",
						"s1":   4.0,
						"s2":   4.0,
					},
				},
			},
		},
	}

	testUtils.ExecuteTEMP(t, test)
}

func TestOneToManyToOneWithSumOfDeepOrderBySubTypeAndDeepOrderBySubtypeOppositeDirections(t *testing.T) {
	test := testUtils.TestCase{
		Description: "1-N-1 sum of deep orderby subtypes and non-sum deep orderby, opposite directions.",
		Actions: []any{
			gqlSchemaOneToManyToOne(),
			createDocsWith6BooksAnd5Publishers(),
			testUtils.Request{
				Request: `query {
					Author {
						name
						s1: _sum(book: {field: rating, order: {publisher: {yearOpened: DESC}}, limit: 2})
						OldestPublishersBook: book(order: {publisher: {yearOpened: ASC}}, limit: 2) {
							name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John Grisham",
						// 'Theif Lord' (4.8 rating) 2020, then 'A Time for Mercy' 2013 (4.5 rating).
						"s1": 4.8 + 4.5,
						"OldestPublishersBook": []map[string]any{
							{
								"name": "The Associate",
							},
							{
								"name": "Painted House",
							},
						},
					},
					{
						"name":                 "Not a Writer",
						"s1":                   0.0,
						"OldestPublishersBook": []map[string]any{},
					},
					{
						"name": "Cornelia Funke",
						"s1":   4.0,
						"OldestPublishersBook": []map[string]any{
							{
								"name": "The Rooster Bar",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTEMP(t, test)
}
