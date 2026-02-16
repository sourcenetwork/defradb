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
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John"}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"rating": 4.5,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"rating": 4.9,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
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
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John"}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"rating": 4.5,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"rating": 4.9,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
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
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John"}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"rating": 4.5,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"rating": 4.9,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
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
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob"}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book A1",
					"rating": 3.5,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book A2",
					"rating": 4.8,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book B1",
					"rating": 4.0,
					"author": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.CreateDoc{
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
			&action.AddSchema{
				Schema: `
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
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob"}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book A1",
					"rating": 3.5,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book A2",
					"rating": 4.8,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book B1",
					"rating": 4.0,
					"author": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.CreateDoc{
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
			&action.AddSchema{
				Schema: `
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
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John"}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Fred"}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book1",
					"rating": 3.5,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book2",
					"rating": 4.8,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book3",
					"rating": 4.2,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
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
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob"}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book A1",
					"rating": 4.8,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book A2",
					"rating": 3.5,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book B1",
					"rating": 2.5,
					"author": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.CreateDoc{
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
			&action.AddSchema{
				Schema: `
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
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John"}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book2020",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book2010",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book2000",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "OrphanBook",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":            "Publisher2020",
					"establishedYear": 2020,
					"book":            testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":            "Publisher2010",
					"establishedYear": 2010,
					"book":            testUtils.NewDocIndex(1, 1),
				},
			},
			&action.CreateDoc{
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
			&action.AddSchema{
				Schema: `
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
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John"}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book2020",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book2010",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book2000",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":            "Publisher2020",
					"establishedYear": 2020,
					"book":            testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":            "Publisher2010",
					"establishedYear": 2010,
					"book":            testUtils.NewDocIndex(1, 1),
				},
			},
			&action.CreateDoc{
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

func TestQueryWithOrderByRelationField_ExhaustiveWithParentSecondaryASC_ShouldIncludeOrphans(t *testing.T) {
	req := `query @exhaustive {
		Book(order: {publisher: {establishedYear: ASC}}) {
			title
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
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
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"title": "Book1"}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"title": "Book2"}`,
			},
			&action.CreateDoc{
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
				// Secondary parent: join fetches 1 publisher + 1 book, orphanNode scans 2 books to find orphans
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1).WithDocFetches(4),
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
			&action.AddSchema{
				Schema: `
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
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"title": "Book1"}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"title": "Book2"}`,
			},
			&action.CreateDoc{
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
				// Secondary parent: join fetches 1 publisher + 1 book, orphanNode scans 2 books to find orphans
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1).WithDocFetches(4),
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
			&action.AddSchema{
				Schema: `
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
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"title":  "Book1",
					"rating": 5,
				},
			},
			&action.CreateDoc{
				CollectionID: 1,
				Doc:          `{"name": "OrphanPublisher"}`,
			},
			&action.CreateDoc{
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
				// Primary parent: join fetches 1 book + 1 publisher, orphanNode uses index to find 1 orphan
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3).WithDocFetches(3),
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
			&action.AddSchema{
				Schema: `
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
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"title":  "Book1",
					"rating": 5,
				},
			},
			&action.CreateDoc{
				CollectionID: 1,
				Doc:          `{"name": "OrphanPublisher"}`,
			},
			&action.CreateDoc{
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
				// Primary parent: join fetches 1 book + 1 publisher, orphanNode uses index to find 1 orphan
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3).WithDocFetches(3),
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
			&action.AddSchema{
				Schema: `
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
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"title": "Book1"}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"title": "Book2"}`,
			},
			&action.CreateDoc{
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
				// Without @exhaustive: join fetches 1 publisher + 1 book, no orphan scanning
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1).WithDocFetches(2),
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
			&action.AddSchema{
				Schema: `
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
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"title":  "Book1",
					"rating": 5,
				},
			},
			&action.CreateDoc{
				CollectionID: 1,
				Doc:          `{"name": "OrphanPublisher"}`,
			},
			&action.CreateDoc{
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
				// Without @exhaustive: join fetches 1 book + 1 publisher, no orphan scanning
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2).WithDocFetches(2),
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
			&action.AddSchema{
				Schema: `
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
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John"}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book2020",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book2010",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book2000",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "OrphanBook",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":            "Publisher2020",
					"establishedYear": 2020,
					"book":            testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":            "Publisher2010",
					"establishedYear": 2010,
					"book":            testUtils.NewDocIndex(1, 1),
				},
			},
			&action.CreateDoc{
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
				// With recursive aggregation, the indexFetches from the Publisher scanNode are now included.
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
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
			&action.AddSchema{
				Schema: `
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
			&action.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "John"}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book2020",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book2010",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "Book2000",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			// OrphanBook has no publisher - would come first in ASC ordering if included
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"title":  "OrphanBook",
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":            "Publisher2020",
					"establishedYear": 2020,
					"book":            testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":            "Publisher2010",
					"establishedYear": 2010,
					"book":            testUtils.NewDocIndex(1, 1),
				},
			},
			&action.CreateDoc{
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
				// docFetches: 1 author + 2 books + 2 publishers = 5 (recursively aggregated)
				// indexFetches: 2 from Publisher index (recursively aggregated from nested join)
				Asserter: testUtils.NewExplainAsserter().WithDocFetches(5).WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
