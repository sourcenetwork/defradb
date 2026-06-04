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
				// sub-filter uses rating index: 4 books for John + 2 for Fred = 6 indexFetches.
				// Books are fetched per author. DESC uses reverse index iteration, no in-memory sort needed.
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
				// subType=Book: parent filter uses rating index via inverted join (1 book matches _geq: 4.0).
				// For the matched author, a clone scan fetches all her books (4 books = 4 index fetches).
				// Total: 1 (filter match) + 4 (clone for Alice's books) = 5 index fetches, 5 doc fetches.
				// The rating index also satisfies the sub-ordering (canSatisfyOrder=true),
				// so the clone skips the redundant in-memory sort — saving 3 iterations.
				// Without the optimization: iterations would be 8 (5 scan + 3 orderNode).
				Asserter: testUtils.NewExplainAsserter("subType").
					WithDocFetches(5).WithFieldFetches(15).WithIndexFetches(5).WithIterations(5),
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
				// subType/subType=Publisher: 2 index fetches (via establishedYear).
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
				// root=Author: 1 doc.
				// subType=Book: 2 docs.
				// subType/subType=Publisher: 2 docs, 2 index (via establishedYear).
				Asserter: testUtils.NewExplainAsserter("root").WithDocFetches(1).
					WithLevel("subType").WithDocFetches(2).
					WithLevel("subType", "subType").WithDocFetches(2).WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithOrderByRelationField_ExhaustiveWithParentSecondaryASC_ShouldIncludeOrphans(t *testing.T) {
	req := `query @exhaustive {
		Book(order: {publisher: {establishedYear: ASC}}) {
			title
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						title: String
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
				Doc:          `{"title": "Book1"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"title": "Book2"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":            "Publisher1",
					"establishedYear": 2020,
					"book":            testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"Book": []map[string]any{
						{"title": "Book2"},
						{"title": "Book1"},
					},
				},
			},
			&action.Request{
				Request: makeExplainQuery(req),
				// root=Book: 1 doc (orphan scan, no index).
				// subType=Publisher: 1 doc, 1 index (via establishedYear).
				Asserter: testUtils.NewExplainAsserter("root").WithDocFetches(1).WithIndexFetches(0).
					WithLevel("subType").WithDocFetches(1).WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithOrderByRelationField_ExhaustiveWithParentSecondaryDESC_ShouldIncludeOrphans(t *testing.T) {
	req := `query @exhaustive {
		Book(order: {publisher: {establishedYear: DESC}}) {
			title
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						title: String
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
				Doc:          `{"title": "Book1"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"title": "Book2"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":            "Publisher1",
					"establishedYear": 2020,
					"book":            testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"Book": []map[string]any{
						{"title": "Book1"},
						{"title": "Book2"},
					},
				},
			},
			&action.Request{
				Request: makeExplainQuery(req),
				// root=Book: 1 doc (orphan scan, no index).
				// subType=Publisher: 1 doc, 1 index (via establishedYear).
				Asserter: testUtils.NewExplainAsserter("root").WithDocFetches(1).WithIndexFetches(0).
					WithLevel("subType").WithDocFetches(1).WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithOrderByRelationField_ExhaustiveWithParentPrimaryASC_ShouldIncludeOrphans(t *testing.T) {
	req := `query @exhaustive {
		Publisher(order: {book: {rating: ASC}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						title: String
						rating: Int @index
						publisher: Publisher
					}
					type Publisher {
						name: String
						book: Book @primary
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"title":  "Book1",
					"rating": 5,
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc:          `{"name": "OrphanPublisher"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name": "LinkedPublisher",
					"book": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"Publisher": []map[string]any{
						{"name": "OrphanPublisher"},
						{"name": "LinkedPublisher"},
					},
				},
			},
			&action.Request{
				Request: makeExplainQuery(req),
				// root=Publisher: 1 doc, 1 index (orphan via book_id IS NULL).
				// subType=Book: 1 doc, 1 index (via rating).
				Asserter: testUtils.NewExplainAsserter("root").WithDocFetches(1).WithIndexFetches(1).
					WithLevel("subType").WithDocFetches(1).WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithOrderByRelationField_ExhaustiveWithParentPrimaryDESC_ShouldIncludeOrphans(t *testing.T) {
	req := `query @exhaustive {
		Publisher(order: {book: {rating: DESC}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						title: String
						rating: Int @index
						publisher: Publisher
					}
					type Publisher {
						name: String
						book: Book @primary
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"title":  "Book1",
					"rating": 5,
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc:          `{"name": "OrphanPublisher"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name": "LinkedPublisher",
					"book": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"Publisher": []map[string]any{
						{"name": "LinkedPublisher"},
						{"name": "OrphanPublisher"},
					},
				},
			},
			&action.Request{
				Request: makeExplainQuery(req),
				// root=Publisher: 1 doc, 1 index (orphan via book_id IS NULL).
				// subType=Book: 1 doc, 1 index (via rating).
				Asserter: testUtils.NewExplainAsserter("root").WithDocFetches(1).WithIndexFetches(1).
					WithLevel("subType").WithDocFetches(1).WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithOrderByRelationField_WithParentSecondaryASC_ExcludesOrphans(t *testing.T) {
	// No @exhaustive directive - orphans should be excluded for performance
	req := `query {
		Book(order: {publisher: {establishedYear: ASC}}) {
			title
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						title: String
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
				Doc:          `{"title": "Book1"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"title": "Book2"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":            "Publisher1",
					"establishedYear": 2020,
					"book":            testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"Book": []map[string]any{
						{"title": "Book1"},
					},
				},
			},
			&action.Request{
				Request: makeExplainQuery(req),
				// root=Book: 1 doc, no index. subType=Publisher: 1 doc, 1 index (via establishedYear).
				Asserter: testUtils.NewExplainAsserter("root").WithDocFetches(1).WithIndexFetches(0).
					WithLevel("subType").WithDocFetches(1).WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithOrderByRelationField_WithParentPrimaryASC_ExcludesOrphans(t *testing.T) {
	// No @exhaustive directive - orphans should be excluded for performance
	req := `query {
		Publisher(order: {book: {rating: ASC}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						title: String
						rating: Int @index
						publisher: Publisher
					}
					type Publisher {
						name: String
						book: Book @primary
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"title":  "Book1",
					"rating": 5,
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc:          `{"name": "OrphanPublisher"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name": "LinkedPublisher",
					"book": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"Publisher": []map[string]any{
						{"name": "LinkedPublisher"},
					},
				},
			},
			&action.Request{
				Request: makeExplainQuery(req),
				// root=Publisher: 1 doc, 1 index (via book_id). subType=Book: 1 doc, 1 index (via rating).
				Asserter: testUtils.NewExplainAsserter("root").WithDocFetches(1).WithIndexFetches(1).
					WithLevel("subType").WithDocFetches(1).WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithNestedOrderByRelationField_WithDESCAndLimit_ExcludesOrphans(t *testing.T) {
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
				// root=Author: 1 doc, no index.
				// subType/root=Book: 2 docs (limit 2, orphan excluded).
				// subType/subType=Publisher: 2 docs, 2 index (via establishedYear).
				Asserter: testUtils.NewExplainAsserter("root").WithDocFetches(1).WithIndexFetches(0).
					WithLevel("subType", "root").WithDocFetches(2).WithIndexFetches(0).
					WithLevel("subType", "subType").WithDocFetches(2).WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Tests that orphan children (without the ordering relation) are excluded in subquery ordering
// when using an index-based inverted join with ASC order. This documents the expected behavior
// where orphans would come first in ASC (NULLS FIRST) but are excluded due to index-based join.
func TestQueryWithNestedOrderByRelationField_WithASCAndLimit_ExcludesOrphans(t *testing.T) {
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
			// OrphanBook has no publisher - would come first in ASC ordering if included
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
			// With ASC ordering and no @exhaustive OrphanBook is excluded.
			// Otherwise result would be OrphanBook, Book2000.
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
				// root=Author: 1 doc, no index.
				// subType/root=Book: 2 docs (limit 2, orphan excluded).
				// subType/subType=Publisher: 2 docs, 2 index (via establishedYear).
				Asserter: testUtils.NewExplainAsserter("root").WithDocFetches(1).WithIndexFetches(0).
					WithLevel("subType", "root").WithDocFetches(2).WithIndexFetches(0).
					WithLevel("subType", "subType").WithDocFetches(2).WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithNestedOrderByRelationField_ExhaustiveWithASCAndLimit_ShouldIncludeOrphansFirst(t *testing.T) {
	req := `query @exhaustive {
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
			// ASC + @exhaustive: OrphanBook comes first (null), then Book2000 (earliest year).
			// Limit 2 is respected after orphan merging.
			&action.Request{
				Request: req,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John",
							"published": []map[string]any{
								{"title": "OrphanBook"},
								{"title": "Book2000"},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithNestedOrderByRelationField_ExhaustiveWithDESCAndLimit_ShouldAppendOrphansLast(t *testing.T) {
	req := `query @exhaustive {
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
			// DESC + @exhaustive: Book2020 first (latest year), then Book2010.
			// Limit 2 is full from join results, orphan not needed.
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithOrderByRelationField_WithSomeDocsWithoutRelation_ShouldIncludeAll(t *testing.T) {
	req := `query @exhaustive {
		Book(order: {publisher: {year: ASC}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						name: String
						publisher: Publisher
					}
					type Publisher {
						year: Int @index
						book: Book @primary
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Book1"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Book2"}`, // No publisher - orphan
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"year": 2020,
					"book": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"Book": []map[string]any{
						{"name": "Book2"}, // null year first in ASC
						{"name": "Book1"},
					},
				},
			},
			&action.Request{
				Request: makeExplainQuery(req),
				// root=Book: 1 doc, no index.
				// subType=Publisher: 1 doc, 1 index (via year).
				Asserter: testUtils.NewExplainAsserter("root").WithDocFetches(1).WithIndexFetches(0).
					WithLevel("subType").WithDocFetches(1).WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithNestedOrderByRelationField_ExhaustiveWithPrimaryParentASC_ShouldIncludeOrphans(t *testing.T) {
	req := `query @exhaustive {
		Author {
			name
			book(order: {publisher: {yearOpened: ASC}}) {
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
						book: [Book]
					}
					type Book {
						title: String
						author: Author
						publisher: Publisher
					}
					type Publisher {
						name: String
						yearOpened: Int @index
						book: [Book]
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "John"}`,
			},
			&action.AddDoc{
				CollectionID: 2,
				Doc:          `{"name": "Publisher2020", "yearOpened": 2020}`,
			},
			// Book with a publisher
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":        "LinkedBook",
					"author":       testUtils.NewDocIndex(0, 0),
					"_publisherID": testUtils.NewDocIndex(2, 0),
				},
			},
			// Book without a publisher (orphan)
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "OrphanBook",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: req,
				// ASC + @exhaustive: OrphanBook (null publisher) comes first, LinkedBook after.
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John",
							"book": []map[string]any{
								{"title": "OrphanBook"},
								{"title": "LinkedBook"},
							},
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithFilterOnNullRelation_SecondaryDocWithoutRelation_ShouldReturnOrphans(t *testing.T) {
	// Book is the secondary side (Publisher stores _bookID via @primary).
	// Querying with order on publisher.establishedYear + @exhaustive triggers orphan detection
	// for Books that have no Publisher.
	req := `query @exhaustive {
		Book(order: {publisher: {establishedYear: ASC}}) {
			title
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						title: String
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
				Doc:          `{"title": "Book With Publisher"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"title": "Orphan Book 1"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"title": "Orphan Book 2"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":            "Publisher2020",
					"establishedYear": 2020,
					"book":            testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"Book": []map[string]any{
						{"title": "Orphan Book 2"},
						{"title": "Orphan Book 1"},
						{"title": "Book With Publisher"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithOrderOnOneToMany_WithSubFilterAndOrderOnSameIndexedField_ShouldFilterThenOrderASC(t *testing.T) {
	req := `query {
		Author {
			name
			published(filter: {rating: {_gt: 3.0}}, order: {rating: ASC}) {
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
				Doc:          `{"name": "John"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Low Rated",
					"rating": 2.0,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Mid Rated",
					"rating": 4.5,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "High Rated",
					"rating": 4.9,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				// Filter and order on the same field (rating) at the child level.
				// Only books with rating > 3.0 should be returned, ordered ASC.
				Request: req,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John",
							"published": []map[string]any{
								{"title": "Mid Rated", "rating": 4.5},
								{"title": "High Rated", "rating": 4.9},
							},
						},
					},
				},
			},
			&action.Request{
				Request: makeExplainQuery(req),
				// root=Author: 1 doc fetched (John), no index on Author.
				// subType=Book: the rating index satisfies both filter and order.
				//   2 index fetches (only entries with rating > 3.0), 2 doc fetches for matched books.
				Asserter: testUtils.NewExplainAsserter("root").WithDocFetches(1).WithIndexFetches(0).
					WithLevel("subType").WithDocFetches(2).WithIndexFetches(2),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithOrderOnOneToMany_WithSubFilterAndOrderOnSameIndexedField_ShouldFilterThenOrderDESC(t *testing.T) {
	req := `query {
		Author {
			name
			published(filter: {rating: {_geq: 4.5}}, order: {rating: DESC}) {
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
				Doc:          `{"name": "John"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Low Rated",
					"rating": 2.0,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Mid Rated",
					"rating": 4.5,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "High Rated",
					"rating": 4.9,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Top Rated",
					"rating": 5.0,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				// Filter and order on the same field (rating) at the child level.
				// Only books with rating >= 4.5 should be returned, ordered DESC.
				Request: req,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John",
							"published": []map[string]any{
								{"title": "Top Rated", "rating": 5.0},
								{"title": "High Rated", "rating": 4.9},
								{"title": "Mid Rated", "rating": 4.5},
							},
						},
					},
				},
			},
			&action.Request{
				Request: makeExplainQuery(req),
				// root=Author: 1 doc fetched (John), no index on Author.
				// subType=Book: the rating index satisfies both filter and order.
				//   3 index fetches (only entries with rating >= 4.5), 3 doc fetches for matched books.
				Asserter: testUtils.NewExplainAsserter("root").WithDocFetches(1).WithIndexFetches(0).
					WithLevel("subType").WithDocFetches(3).WithIndexFetches(3),
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
