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
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "John"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"rating": 4.5,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"rating": 4.9,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"rating": 4.2,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
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
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter("subType").WithIndexFetches(3),
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
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "John"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"rating": 4.5,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"rating": 4.9,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"rating": 4.2,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
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
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter("subType").WithIndexFetches(3),
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
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "John"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"rating": 4.5,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"rating": 4.9,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"rating": 4.2,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
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
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter("subType").WithIndexFetches(1),
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
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book A1",
					"rating": 3.5,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book A2",
					"rating": 4.8,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book B1",
					"rating": 4.0,
					"author": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book B2",
					"rating": 2.5,
					"author": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.Request{
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
			&action.Request{
				Request: makeExplainQuery(req),
				// index fetches 8: 4 for ordering all books for each author
				Asserter: testUtils.NewExplainAsserter("subType").WithIndexFetches(8),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithOrderOnOneToMany_WithMultipleAuthorsAndIndexOnRelation_ShouldOrderEachAuthorsBooks(t *testing.T) {
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
			&action.AddCollection{
				SDL: `
					type Author {
						name: String
						published: [Book]
					}
					type Book {
						title: String
						rating: Float @index
						author: Author @index
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book A1",
					"rating": 3.5,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book A2",
					"rating": 4.8,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book B1",
					"rating": 4.0,
					"author": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book B2",
					"rating": 2.5,
					"author": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.Request{
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
			&action.Request{
				Request: makeExplainQuery(req),
				// index fetches 4: relation ID index fetches 2 books per author, then sorts in memory
				Asserter: testUtils.NewExplainAsserter("subType").WithIndexFetches(4),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithOrderOnOneToMany_WithSubFilterAndOrderAndRelationIndex_ShouldFilterThenOrder(t *testing.T) {
	req := `query {
		Author {
			name
			published(filter: {rating: {_geq: 4.0}}, order: {rating: DESC}) {
				title
				rating
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Author {
						name: String
						published: [Book]
					}
					type Book {
						title: String
						rating: Float @index
						author: Author @index
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "John"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Fred"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book1",
					"rating": 3.5,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book2",
					"rating": 4.8,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book3",
					"rating": 4.2,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book4",
					"rating": 4.4,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John",
							"published": []map[string]any{
								{"title": "Book2", "rating": 4.8},
								{"title": "Book4", "rating": 4.4},
								{"title": "Book3", "rating": 4.2},
							},
						},
						{
							"name":      "Fred",
							"published": []map[string]any{},
						},
					},
				},
			},
			&action.Request{
				Request: makeExplainQuery(req),
				// 6 indexFetches: sub-filter uses rating index (3 books match filter rating _geq: 4.0) for 2 authors,
				// DESC instructs the index to iterate in reverse order, so no in-memory sort needed
				Asserter: testUtils.NewExplainAsserter("subType").WithIndexFetches(6),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithOrderOnOneToMany_WithParentFilterOnRelationAndSubOrder_ShouldOrderChildren(t *testing.T) {
	req := `query {
		Author(filter: {published: {rating: {_geq: 4.0}}}) {
			name
			published(order: {rating: DESC}) {
				title
				rating
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book A1",
					"rating": 4.8,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book A2",
					"rating": 3.5,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book B1",
					"rating": 2.5,
					"author": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book B2",
					"rating": 3.0,
					"author": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "Alice",
							"published": []map[string]any{
								{"title": "Book A1", "rating": 4.8},
								{"title": "Book A2", "rating": 3.5},
							},
						},
					},
				},
			},
			&action.Request{
				Request: makeExplainQuery(req),
				// 5 indexFetch: parent filter uses rating index via inverted join (1 book matches _ge: 4.0)
				// For the matched author full index scan is done to get all 4 books
				Asserter: testUtils.NewExplainAsserter("subType").WithIndexFetches(5),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithNestedOrderByRelationField_WithDESCAndLimit_RecursiveExplain(t *testing.T) {
	req := `query {
		Author {
			name
			published(order: {publisher: {establishedYear: DESC}}, limit: 2) {
				title
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Author {
						name: String
						published: [Book]
					}
					type Book {
						title: String
						author: Author
						publisher: Publisher
					}
					type Publisher {
						name: String
						establishedYear: Int @index
						book: Book @primary
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "John"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book2020",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book2010",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book2000",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "OrphanBook",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":            "Publisher2020",
					"establishedYear": 2020,
					"book":            testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":            "Publisher2010",
					"establishedYear": 2010,
					"book":            testUtils.NewDocIndex(1, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":            "Publisher2000",
					"establishedYear": 2000,
					"book":            testUtils.NewDocIndex(1, 2),
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John",
							"published": []map[string]any{
								{"title": "Book2020"},
								{"title": "Book2010"},
							},
						},
					},
				},
			},
			&action.Request{
				Request: makeExplainQuery(req),
				// The index on Publisher.establishedYear is used by the nested Book->Publisher join.
				// Publisher is at subType/subType (nested inside Book which is at subType).
				Asserter: testUtils.NewExplainAsserter("subType", "subType").WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithNestedOrderByRelationField_WithASCAndLimit_RecursiveExplain(t *testing.T) {
	req := `query {
		Author {
			name
			published(order: {publisher: {establishedYear: ASC}}, limit: 2) {
				title
			}
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Author {
						name: String
						published: [Book]
					}
					type Book {
						title: String
						author: Author
						publisher: Publisher
					}
					type Publisher {
						name: String
						establishedYear: Int @index
						book: Book @primary
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "John"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book2020",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book2010",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book2000",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":            "Publisher2020",
					"establishedYear": 2020,
					"book":            testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":            "Publisher2010",
					"establishedYear": 2010,
					"book":            testUtils.NewDocIndex(1, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":            "Publisher2000",
					"establishedYear": 2000,
					"book":            testUtils.NewDocIndex(1, 2),
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John",
							"published": []map[string]any{
								{"title": "Book2000"},
								{"title": "Book2010"},
							},
						},
					},
				},
			},
			&action.Request{
				Request: makeExplainQuery(req),
				// Author root: 1 docFetch
				// Book (subType): 2 docFetches
				// Publisher (subType/subType): 2 docFetches, 2 indexFetches
				Asserter: testUtils.NewExplainAsserter("root").WithDocFetches(1).
					WithLevel("subType").WithDocFetches(2).
					WithLevel("subType", "subType").WithDocFetches(2).WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
