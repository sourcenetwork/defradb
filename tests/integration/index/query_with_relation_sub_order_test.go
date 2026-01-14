// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package index

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryWithOrderOnOneToMany_WithIndexOnOrderFieldDescending_ShouldOrder(t *testing.T) {
	req := `query {
		Author {
			name
			published(order: {rating: DESC}) {
				rating
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Author {
						name: String
						published: [Book]
					}
					type Book {
						rating: Float @index
						author: Author
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John"}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"rating": 4.5,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"rating": 4.9,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"rating": 4.2,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John",
							"published": []map[string]any{
								{"rating": 4.9},
								{"rating": 4.5},
								{"rating": 4.2},
							},
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithOrderOnOneToMany_WithIndexOnOrderFieldAscending_ShouldOrder(t *testing.T) {
	req := `query {
		Author {
			name
			published(order: {rating: ASC}) {
				rating
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Author {
						name: String
						published: [Book]
					}
					type Book {
						rating: Float @index
						author: Author
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John"}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"rating": 4.5,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"rating": 4.9,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"rating": 4.2,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John",
							"published": []map[string]any{
								{"rating": 4.2},
								{"rating": 4.5},
								{"rating": 4.9},
							},
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithOrderOnOneToMany_WithIndexOnOrderFieldAscendingWithLimit_ShouldOrderAndLimit(t *testing.T) {
	req := `query {
		Author {
			name
			published(order: {rating: ASC}, limit: 1) {
				rating
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Author {
						name: String
						published: [Book]
					}
					type Book {
						rating: Float @index
						author: Author
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John"}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"rating": 4.5,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"rating": 4.9,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"rating": 4.2,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John",
							"published": []map[string]any{
								{"rating": 4.2},
							},
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithOrderOnOneToMany_WithMultipleAuthors_ShouldOrderEachAuthorsBooks(t *testing.T) {
	req := `query {
		Author(order: {name: ASC}) {
			name
			published(order: {rating: DESC}) {
				title
				rating
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Author {
						name: String
						published: [Book]
					}
					type Book {
						title: String
						rating: Float @index
						author: Author
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob"}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book A1",
					"rating": 3.5,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book A2",
					"rating": 4.8,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book B1",
					"rating": 4.0,
					"author": testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book B2",
					"rating": 2.5,
					"author": testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "Alice",
							"published": []map[string]any{
								{"title": "Book A2", "rating": 4.8},
								{"title": "Book A1", "rating": 3.5},
							},
						},
						{
							"name": "Bob",
							"published": []map[string]any{
								{"title": "Book B1", "rating": 4.0},
								{"title": "Book B2", "rating": 2.5},
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req),
				// index fetches 8: 4 for ordering all books for each author
				Asserter: testUtils.NewExplainAsserter().WithOrder().WithIndexFetches(8),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
