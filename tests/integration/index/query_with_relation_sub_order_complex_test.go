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

func TestQueryWithOrderByRelationField_ExhaustiveASCWithLimit_ManyOrphansEarlyTermination(t *testing.T) {
	req := `query @exhaustive {
		Book(order: {publisher: {establishedYear: ASC}}, limit: 3) {
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
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Linked-2020"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Linked-2010"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Linked-2000"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Orphan-A"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Orphan-B"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Orphan-C"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Orphan-D"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Orphan-E"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Orphan-F"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Orphan-G"}`},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":            "Pub2020",
					"establishedYear": 2020,
					"book":            testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":            "Pub2010",
					"establishedYear": 2010,
					"book":            testUtils.NewDocIndex(0, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":            "Pub2000",
					"establishedYear": 2000,
					"book":            testUtils.NewDocIndex(0, 2),
				},
			},
			&action.Request{
				Request:           req,
				NonOrderedResults: true,
				Results: map[string]any{
					"Book": []map[string]any{
						{"title": "Orphan-A"},
						{"title": "Orphan-F"},
						{"title": "Orphan-B"},
					},
				},
			},
			// orphanNode scans only 4 of 10 Books to find 3 orphans (early termination).
			// Each parent gets a Has() call on the child FK index (4 indexFetches).
			// Source phase (ordered join) never entered — root/subType scanNodes show 0 fetches.
			&action.Request{
				Request: makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter("orphanNode").
					WithDocFetches(4).
					WithIndexFetches(4).
					WithLevel("root").
					WithDocFetches(0).
					WithIndexFetches(0).
					WithLevel("subType").
					WithDocFetches(0).
					WithIndexFetches(0),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithOrderByRelationField_ExhaustiveASC_ManyBooksShowsFullPipeline(t *testing.T) {
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
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Linked-2020"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Linked-2010"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Linked-2000"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Orphan-A"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Orphan-B"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Orphan-C"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Orphan-D"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Orphan-E"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Orphan-F"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Orphan-G"}`},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":            "Pub2020",
					"establishedYear": 2020,
					"book":            testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":            "Pub2010",
					"establishedYear": 2010,
					"book":            testUtils.NewDocIndex(0, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":            "Pub2000",
					"establishedYear": 2000,
					"book":            testUtils.NewDocIndex(0, 2),
				},
			},
			&action.Request{
				Request:           req,
				NonOrderedResults: true,
				Results: map[string]any{
					"Book": []map[string]any{
						{"title": "Orphan-A"},
						{"title": "Orphan-F"},
						{"title": "Orphan-B"},
						{"title": "Orphan-E"},
						{"title": "Orphan-D"},
						{"title": "Orphan-C"},
						{"title": "Orphan-G"},
						{"title": "Linked-2000"},
						{"title": "Linked-2010"},
						{"title": "Linked-2020"},
					},
				},
			},
			// Without limit, orphanNode scans all 10 Books and does a Has() for each.
			// docFetches=10 (parent scans), indexFetches=10 (one Has() per parent).
			// Source phase also runs: root fetches 3 linked Books, subType fetches 3 Publishers via index.
			&action.Request{
				Request: makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter("orphanNode").
					WithDocFetches(10).
					WithIndexFetches(10).
					WithLevel("root").
					WithDocFetches(3).
					WithIndexFetches(0).
					WithLevel("subType").
					WithDocFetches(3).
					WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithOrderByRelationField_ExhaustiveDESCWithLimit_ManyOrphansSkipsOrphanPhase(t *testing.T) {
	req := `query @exhaustive {
		Book(order: {publisher: {establishedYear: DESC}}, limit: 3) {
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
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Linked-2020"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Linked-2010"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Linked-2000"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Orphan-A"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Orphan-B"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Orphan-C"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Orphan-D"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Orphan-E"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Orphan-F"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"title": "Orphan-G"}`},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":            "Pub2020",
					"establishedYear": 2020,
					"book":            testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":            "Pub2010",
					"establishedYear": 2010,
					"book":            testUtils.NewDocIndex(0, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":            "Pub2000",
					"establishedYear": 2000,
					"book":            testUtils.NewDocIndex(0, 2),
				},
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"Book": []map[string]any{
						{"title": "Linked-2020"},
						{"title": "Linked-2010"},
						{"title": "Linked-2000"},
					},
				},
			},
			// DESC puts source (ordered join) first. Limit 3 satisfied by 3 linked books.
			// Orphan phase never entered — 0 doc/index fetches on orphanNode.
			// Source phase fetches all 3: root scans 3 Books, subType looks up 3 Publishers via index.
			&action.Request{
				Request: makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter("orphanNode").
					WithDocFetches(0).
					WithIndexFetches(0).
					WithLevel("root").
					WithDocFetches(3).
					WithIndexFetches(0).
					WithLevel("subType").
					WithDocFetches(3).
					WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
