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

	"github.com/sourcenetwork/defradb/internal/planner"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// bm25Articles is the corpus the ranking tests run against. "database" is in three of the four
// documents and "indexing" in two, so the two terms carry different weight; the three documents
// holding "database" are of three different lengths, so length normalisation separates them.
var bm25Articles = []any{
	&action.AddDoc{Doc: `{"title": "both", "body": "database indexing"}`},
	&action.AddDoc{Doc: `{"title": "common", "body": "database systems"}`},
	&action.AddDoc{Doc: `{"title": "common and long", "body": "database design patterns"}`},
	&action.AddDoc{Doc: `{"title": "rare", "body": "indexing"}`},
}

func bm25ArticleActions(directive string, requests ...any) []any {
	actions := []any{
		&action.AddCollection{
			SDL: `
				type Article {
					title: String
					body: String ` + directive + `
				}
			`,
		},
	}
	actions = append(actions, bm25Articles...)
	return append(actions, requests...)
}

// The documents come back best match first. A document holding only the term two thirds of the
// collection holds scores below one holding only the term half of it holds, which is the whole
// point of weighting a term by how rare it is.
func TestBM25Query_RanksByRelevance(t *testing.T) {
	test := testUtils.TestCase{
		Actions: bm25ArticleActions(`@fullTextIndex`, &action.Request{
			Request: `query {
				Article(order: {_alias: {rank: DESC}}) {
					title
					rank: _bm25(query: "database indexing", fields: ["body"])
				}
			}`,
			Results: map[string]any{
				"Article": []map[string]any{
					{"title": "both", "rank": 1.0498221244986776},
					{"title": "rare", "rank": 0.8713850269896455},
					{"title": "common", "rank": 0.3566749439387324},
					{"title": "common and long", "rank": 0.29610750062838165},
				},
			},
		}),
	}

	testUtils.ExecuteTestCase(t, test)
}

// A limit returns exactly that many documents, taken from the top of the ranking.
func TestBM25Query_WithLimit_ReturnsTheBestThatMany(t *testing.T) {
	test := testUtils.TestCase{
		Actions: bm25ArticleActions(`@fullTextIndex`, &action.Request{
			Request: `query {
				Article(order: {_alias: {rank: DESC}}, limit: 2) {
					title
					rank: _bm25(query: "database indexing", fields: ["body"])
				}
			}`,
			Results: map[string]any{
				"Article": []map[string]any{
					{"title": "both", "rank": 1.0498221244986776},
					{"title": "rare", "rank": 0.8713850269896455},
				},
			},
		}),
	}

	testUtils.ExecuteTestCase(t, test)
}

// Filtering happens above the ranked candidate source. The limit therefore keeps pulling after a
// higher-scoring document is rejected and returns the best K documents that actually survive.
func TestBM25Query_WithFilterAndLimit_ReturnsTheBestSurvivingDocuments(t *testing.T) {
	test := testUtils.TestCase{
		Actions: bm25ArticleActions(`@fullTextIndex`, &action.Request{
			Request: `query {
				Article(
					filter: {title: {_neq: "both"}}
					order: {_alias: {rank: DESC}}
					limit: 2
				) {
					title
					rank: _bm25(query: "database indexing", fields: ["body"])
				}
			}`,
			Results: map[string]any{
				"Article": []map[string]any{
					{"title": "rare", "rank": 0.8713850269896455},
					{"title": "common", "rank": 0.3566749439387324},
				},
			},
		}),
	}

	testUtils.ExecuteTestCase(t, test)
}

// The scan yields in score order, so an ordering that is exactly the _bm25 field descending is
// already satisfied and no order node is added. BM25 posting reads and scoring are eagerly
// materialized by design, but this keeps document fetching lazy: an order node would drain every
// ranked candidate into a sort buffer before the limit could stop pulling documents.
//
// The plan shape is asserted rather than the results, because an order node re-sorting what the
// scan already sorted returns exactly the same documents. What it changes is docFetches, which is
// two of the four documents here and would be four with the order node in place.
func TestBM25Query_OrderedByRankDescending_DropsTheOrderNode(t *testing.T) {
	test := testUtils.TestCase{
		Actions: bm25ArticleActions(`@fullTextIndex`, &action.ExplainRequest{
			Request: `query @explain(type: execute) {
				Article(order: {_alias: {rank: DESC}}, limit: 2) {
					title
					rank: _bm25(query: "database indexing", fields: ["body"])
				}
			}`,
			ExpectedFullGraph: map[string]any{
				"explain": map[string]any{
					"executionSuccess": true,
					"sizeOfResult":     1,
					"planExecutions":   uint64(2),
					"operationNode": []map[string]any{
						{
							"selectTopNode": map[string]any{
								"limitNode": map[string]any{
									"iterations": uint64(3),
									"selectNode": map[string]any{
										"iterations":    uint64(2),
										"filterMatches": uint64(2),
										"scanNode": map[string]any{
											"iterations":   uint64(2),
											"docFetches":   uint64(2),
											"fieldFetches": uint64(4),
											"indexFetches": uint64(5),
										},
									},
								},
							},
						},
					},
				},
			},
		}),
	}

	testUtils.ExecuteTestCase(t, test)
}

// An ordering that is not the ranking still gets an order node, so the scan's own order is not
// mistaken for any order the caller asks for.
func TestBM25Query_OrderedBySomethingElse_KeepsTheOrderNode(t *testing.T) {
	test := testUtils.TestCase{
		Actions: bm25ArticleActions(`@fullTextIndex`, &action.ExplainRequest{
			Request: `query @explain(type: simple) {
				Article(order: {title: ASC}) {
					title
					rank: _bm25(query: "database indexing", fields: ["body"])
				}
			}`,
			ExpectedPatterns: map[string]any{
				"explain": map[string]any{
					"operationNode": []map[string]any{
						{
							"selectTopNode": map[string]any{
								"orderNode": map[string]any{
									"selectNode": map[string]any{
										"scanNode": map[string]any{},
									},
								},
							},
						},
					},
				},
			},
		}),
	}

	testUtils.ExecuteTestCase(t, test)
}

// b decides how much a field longer than average is penalised. Two documents holding the query
// term twice, one of them padded, change places when it is turned off.
func TestBM25Query_WithFieldLengthNormalisation_ChangesTheOrder(t *testing.T) {
	documents := []any{
		&action.AddDoc{Doc: `{"title": "long", "body": "alpha alpha beta gamma delta epsilon"}`},
		&action.AddDoc{Doc: `{"title": "short", "body": "alpha zeta"}`},
	}

	assert := func(directive string, expected []string) {
		actions := []any{
			&action.AddCollection{
				SDL: `type Article { title: String  body: String ` + directive + ` }`,
			},
		}
		actions = append(actions, documents...)

		results := make([]map[string]any, 0, len(expected))
		for _, title := range expected {
			// The scores themselves are not the point here, only the order they put the
			// documents in, and pinning them would only restate the formula.
			results = append(results, map[string]any{
				"title": title,
				"rank":  testUtils.BeNumerically(">", 0.0),
			})
		}
		actions = append(actions, &action.Request{
			Request: `query {
				Article(order: {_alias: {rank: DESC}}) {
					title
					rank: _bm25(query: "alpha", fields: ["body"])
				}
			}`,
			Results: map[string]any{"Article": results},
		})

		testUtils.ExecuteTestCase(t, testUtils.TestCase{Actions: actions})
	}

	// With normalisation off only the two occurrences count, so the longer document wins.
	assert(`@fullTextIndex(BM25: {b: 0})`, []string{"long", "short"})
	// With it on the longer document's two occurrences are worth less than the shorter one's one.
	assert(`@fullTextIndex(BM25: {b: 0.75})`, []string{"short", "long"})
}

// k1 decides how quickly a repeated term stops adding to the score. At zero every occurrence
// after the first is worth nothing, so matching both query terms beats repeating one of them; set
// high the repetition keeps paying and the order reverses.
func TestBM25Query_WithTermFrequencySaturation_ChangesTheOrder(t *testing.T) {
	documents := []any{
		&action.AddDoc{Doc: `{"title": "repeated", "body": "alpha alpha alpha alpha"}`},
		&action.AddDoc{Doc: `{"title": "both terms", "body": "alpha beta"}`},
		&action.AddDoc{Doc: `{"title": "filler one", "body": "beta gamma"}`},
		&action.AddDoc{Doc: `{"title": "filler two", "body": "beta delta"}`},
	}

	assert := func(directive string, expected []string) {
		actions := []any{
			&action.AddCollection{
				SDL: `type Article { title: String  body: String ` + directive + ` }`,
			},
		}
		actions = append(actions, documents...)

		results := make([]map[string]any, 0, len(expected))
		for _, title := range expected {
			results = append(results, map[string]any{
				"title": title,
				"rank":  testUtils.BeNumerically(">", 0.0),
			})
		}
		actions = append(actions, &action.Request{
			Request: `query {
				Article(order: {_alias: {rank: DESC}}, limit: 2) {
					title
					rank: _bm25(query: "alpha beta", fields: ["body"])
				}
			}`,
			Results: map[string]any{"Article": results},
		})

		testUtils.ExecuteTestCase(t, testUtils.TestCase{Actions: actions})
	}

	assert(`@fullTextIndex(BM25: {k1: 0})`, []string{"both terms", "repeated"})
	assert(`@fullTextIndex(BM25: {k1: 10})`, []string{"repeated", "both terms"})
}

// A document whose indexed field no longer holds a query term drops out of the ranking, and a
// deleted one goes with it. Both also change the collection totals every remaining score is
// measured against.
func TestBM25Query_AfterUpdateAndDelete_RanksWhatIsLeft(t *testing.T) {
	test := testUtils.TestCase{
		Actions: bm25ArticleActions(`@fullTextIndex`,
			// "both" stops holding either query term.
			&action.UpdateDoc{
				DocID: 0,
				Doc:   `{"body": "unrelated words entirely"}`,
			},
			&action.Request{
				Request: `mutation {
					delete_Article(docID: "{{.DocID0_3}}") { _docID }
				}`,
				Results: map[string]any{
					"delete_Article": []map[string]any{{"_docID": "{{.DocID0_3}}"}},
				},
			},
			&action.Request{
				Request: `query {
					Article(order: {_alias: {rank: DESC}}) {
						title
						rank: _bm25(query: "database indexing", fields: ["body"])
					}
				}`,
				Results: map[string]any{
					"Article": []map[string]any{
						{"title": "common", "rank": 0.5235483465015789},
						{"title": "common and long", "rank": 0.44713858782297006},
					},
				},
			},
		),
	}

	testUtils.ExecuteTestCase(t, test)
}

// A BM25 score is not computable from the document on its own, so a field with no BM25 index has
// no answer to give rather than a slower way of reaching one.
func TestBM25Query_WithoutIndex_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: bm25ArticleActions(``, &action.Request{
			Request: `query {
				Article {
					title
					rank: _bm25(query: "database indexing", fields: ["body"])
				}
			}`,
			ExpectedError: planner.NewErrNoBM25Index("Article", "body").Error(),
		}),
	}

	testUtils.ExecuteTestCase(t, test)
}

// An index of another kind on the field is not a BM25 index.
func TestBM25Query_WithIndexOfAnotherKind_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: bm25ArticleActions(`@trigramIndex`, &action.Request{
			Request: `query {
				Article {
					rank: _bm25(query: "database indexing", fields: ["body"])
				}
			}`,
			ExpectedError: planner.NewErrNoBM25Index("Article", "body").Error(),
		}),
	}

	testUtils.ExecuteTestCase(t, test)
}

// The child scan of a join runs once per parent row, over the documents that row relates to. A
// ranked scan produces its own document set instead, so the two do not compose and the request is
// refused rather than half served.
func TestBM25Query_OnTheChildSideOfAJoin_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Author {
						name: String
						articles: [Article]
					}
					type Article {
						title: String
						body: String @fullTextIndex
						author: Author
					}
				`,
			},
			&action.Request{
				Request: `query {
					Author {
						name
						articles {
							title
							rank: _bm25(query: "database", fields: ["body"])
						}
					}
				}`,
				ExpectedError: planner.ErrBM25NotOnCollectionScan.Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// A mutation's select decides which documents are written. Ranking it would decide that by
// relevance to a text query rather than by the mutation's own filter.
func TestBM25Query_OnAMutation_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: bm25ArticleActions(`@fullTextIndex`, &action.Request{
			Request: `mutation {
				delete_Article(filter: {title: {_eq: "common"}}) {
					title
					rank: _bm25(query: "database", fields: ["body"])
				}
			}`,
			ExpectedError: mapper.ErrBm25InMutation.Error(),
		}),
	}

	testUtils.ExecuteTestCase(t, test)
}

// One scan yields one ranking, so a second _bm25 field would come back unfilled.
func TestBM25Query_WithTwoBM25Fields_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: bm25ArticleActions(`@fullTextIndex`, &action.Request{
			Request: `query {
				Article {
					first: _bm25(query: "database", fields: ["body"])
					second: _bm25(query: "indexing", fields: ["body"])
				}
			}`,
			ExpectedError: planner.ErrMultipleBM25Fields.Error(),
		}),
	}

	testUtils.ExecuteTestCase(t, test)
}

// A query term no document holds matches nothing, the same as any other query whose terms are all
// absent.
func TestBM25Query_WithUnknownTerm_ReturnsNothing(t *testing.T) {
	test := testUtils.TestCase{
		Actions: bm25ArticleActions(`@fullTextIndex`, &action.Request{
			Request: `query {
				Article(order: {_alias: {rank: DESC}}) {
					title
					rank: _bm25(query: "sharding", fields: ["body"])
				}
			}`,
			Results: map[string]any{"Article": []map[string]any{}},
		}),
	}

	testUtils.ExecuteTestCase(t, test)
}

// Deleted-document scans combine active and tombstoned records, while a full-text index only
// contains active field values. Refuse the combination instead of silently returning an
// incomplete showDeleted result.
func TestBM25Query_WithShowDeleted_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: bm25ArticleActions(`@fullTextIndex`, &action.Request{
			Request: `query {
				Article(showDeleted: true) {
					title
					rank: _bm25(query: "database", fields: ["body"])
				}
			}`,
			ExpectedError: planner.ErrBM25WithShowDeleted.Error(),
		}),
	}

	testUtils.ExecuteTestCase(t, test)
}

// BM25 and vector similarity are both candidate-generating rank drivers. A scan cannot be driven
// by both without changing one ranking's semantics, so the planner rejects the request explicitly.
func TestBM25Query_WithSimilarity_ReturnsCompetingRankError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{SDL: `
				type Article {
					body: String @fullTextIndex
					vector: [Float32!] @vectorIndex(dimensions: 3, HNSW: {metric: COSINE})
				}
			`},
			&action.Request{
				Request: `query {
					Article(order: {_alias: {sim: DESC}}, limit: 2) {
						rank: _bm25(query: "database", fields: ["body"])
						sim: SIMILARITY(vector: {vector: [1, 0, 0]})
					}
				}`,
				ExpectedError: planner.ErrCompetingRankedIndexes.Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
