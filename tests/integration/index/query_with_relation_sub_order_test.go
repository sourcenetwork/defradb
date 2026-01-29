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
				Asserter: testUtils.NewExplainAsserter().WithOrder().WithIndexFetches(8),
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
				Asserter: testUtils.NewExplainAsserter().WithOrder().WithIndexFetches(4),
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
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(6),
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
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(5),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithOrderByRelationField_WithParentSecondaryASC_ShouldIncludeOrphans(t *testing.T) {
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
				Doc:          `{"title": "Book2"}`, // No publisher - orphan
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
				Request: `query {
					Book(order: {publisher: {establishedYear: ASC}}) {
						title
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{"title": "Book2"}, // null establishedYear first in ASC
						{"title": "Book1"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithOrderByRelationField_WithParentSecondaryDESC_ShouldIncludeOrphans(t *testing.T) {
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
				Doc:          `{"title": "Book2"}`, // No publisher - orphan
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
				Request: `query {
					Book(order: {publisher: {establishedYear: DESC}}) {
						title
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{"title": "Book1"},
						{"title": "Book2"}, // null establishedYear last in DESC
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithOrderByRelationField_WithParentPrimaryASC_ShouldIncludeOrphans(t *testing.T) {
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
				Doc: `{"name": "OrphanPublisher"}`, // No book - orphan
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name": "LinkedPublisher",
					"book": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: `query {
					Publisher(order: {book: {rating: ASC}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Publisher": []map[string]any{
						{"name": "OrphanPublisher"}, // null rating first in ASC (orphan - no book)
						{"name": "LinkedPublisher"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithOrderByRelationField_WithParentPrimaryDESC_ShouldIncludeOrphans(t *testing.T) {
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
				Doc: `{"name": "OrphanPublisher"}`, // No book - orphan
			},
			&action.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name": "LinkedPublisher",
					"book": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: `query {
					Publisher(order: {book: {rating: DESC}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Publisher": []map[string]any{
						{"name": "LinkedPublisher"},
						{"name": "OrphanPublisher"}, // null rating last in DESC (orphan - no book)
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
