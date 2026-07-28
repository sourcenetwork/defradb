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

	"github.com/onsi/gomega"

	"github.com/sourcenetwork/defradb/internal/planner"
	"github.com/sourcenetwork/defradb/internal/request/graphql/parser"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// bm25MultiFieldArticles is the corpus the multi-field tests run against. Each document holds the
// query term in one field only, so which field a boost favours decides the whole order.
var bm25MultiFieldArticles = []any{
	&action.AddDoc{Doc: `{"name": "in title", "title": "sharding", "body": "storage layout"}`},
	&action.AddDoc{Doc: `{"name": "in body", "title": "storage layout", "body": "sharding"}`},
	&action.AddDoc{Doc: `{"name": "in neither", "title": "replication", "body": "gossip"}`},
}

// bm25MultiFieldActions builds a collection with a BM25 index on both title and body.
func bm25MultiFieldActions(requests ...any) []any {
	actions := []any{
		&action.AddCollection{
			SDL: `
				type Article {
					name: String
					title: String @index(kind: BM25)
					body: String @index(kind: BM25)
				}
			`,
		},
	}
	actions = append(actions, bm25MultiFieldArticles...)
	return append(actions, requests...)
}

// names reads the name of every returned document in order, which is what the ranking tests below
// assert. The scores themselves are not pinned: they would only restate the formula, and the
// weighting is visible in the order.
func names(expected ...string) map[string]any {
	results := make([]map[string]any, 0, len(expected))
	for _, name := range expected {
		results = append(results, map[string]any{
			"name": name,
			"rank": gomega.BeNumerically(">", 0.0),
		})
	}
	return map[string]any{"Article": results}
}

// A document is scored over every field named, so one holding the term in either field is
// returned and the one holding it in neither is not.
//
// The corpus is symmetric: the two matching documents hold the term once, in fields of equal
// length, in indexes with equal statistics, so with equal weights they score equally and the order
// is the tie break on document short ID, which is insertion order. That is deliberate. It means
// any order the weighted tests below produce comes from the weights and from nothing else.
func TestBM25MultiFieldQuery_ScoresOverEveryFieldNamed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: bm25MultiFieldActions(&action.Request{
			Request: `query {
				Article(order: {_alias: {rank: DESC}}) {
					name
					rank: _bm25(query: "sharding", fields: ["title", "body"])
				}
			}`,
			Results: names("in title", "in body"),
		}),
	}

	testUtils.ExecuteTestCase(t, test)
}

// The weight after "^" decides how much a field's score counts. The same query over the same
// documents puts whichever field is weighted higher on top, which is the whole point of naming
// weights per field.
func TestBM25MultiFieldQuery_WithBoost_ReordersByWhichFieldMatched(t *testing.T) {
	assert := func(fields string, expected ...string) {
		testUtils.ExecuteTestCase(t, testUtils.TestCase{
			Actions: bm25MultiFieldActions(&action.Request{
				Request: `query {
					Article(order: {_alias: {rank: DESC}}) {
						name
						rank: _bm25(query: "sharding", fields: ` + fields + `)
					}
				}`,
				Results: names(expected...),
			}),
		})
	}

	assert(`["title^4", "body"]`, "in title", "in body")
	assert(`["title", "body^4"]`, "in body", "in title")
}

// A field named with no weight counts once, so naming every field with the same weight ranks the
// documents the same way as naming none of them with one.
func TestBM25MultiFieldQuery_WithoutBoost_DefaultsToOne(t *testing.T) {
	assert := func(fields string, expected ...string) {
		testUtils.ExecuteTestCase(t, testUtils.TestCase{
			Actions: bm25MultiFieldActions(&action.Request{
				Request: `query {
					Article(order: {_alias: {rank: DESC}}) {
						name
						rank: _bm25(query: "sharding", fields: ` + fields + `)
					}
				}`,
				Results: names(expected...),
			}),
		})
	}

	assert(`["title^1", "body^1"]`, "in title", "in body")
	assert(`["title", "body"]`, "in title", "in body")
}

// A weight of zero takes the field out of the query: it can neither raise a document's score nor
// bring a document into the results on its own, so the document matching only there is gone.
func TestBM25MultiFieldQuery_WithZeroBoost_ExcludesTheField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: bm25MultiFieldActions(&action.Request{
			Request: `query {
				Article(order: {_alias: {rank: DESC}}) {
					name
					rank: _bm25(query: "sharding", fields: ["title^0", "body"])
				}
			}`,
			Results: names("in body"),
		}),
	}

	testUtils.ExecuteTestCase(t, test)
}

// A weighted-zero field is not read at all, so it costs no index fetches. Only the results would
// change if it were merely scored and discarded, and the count is what shows it was skipped.
func TestBM25MultiFieldQuery_WithZeroBoost_DoesNotReadTheField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: bm25MultiFieldActions(&action.ExplainRequest{
			Request: `query @explain(type: execute) {
				Article {
					name
					rank: _bm25(query: "sharding", fields: ["title^0", "body"])
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
								"selectNode": map[string]any{
									"iterations":    uint64(2),
									"filterMatches": uint64(1),
									"scanNode": map[string]any{
										"iterations":   uint64(2),
										"docFetches":   uint64(1),
										"fieldFetches": uint64(3),
										"indexFetches": uint64(1),
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

// Each field's index is read separately, so a query over two of them fetches from both.
func TestBM25MultiFieldQuery_ReadsEveryNamedFieldsIndex(t *testing.T) {
	test := testUtils.TestCase{
		Actions: bm25MultiFieldActions(&action.ExplainRequest{
			Request: `query @explain(type: execute) {
				Article {
					name
					rank: _bm25(query: "sharding", fields: ["title", "body"])
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
								"selectNode": map[string]any{
									"iterations":    uint64(3),
									"filterMatches": uint64(2),
									"scanNode": map[string]any{
										"iterations":   uint64(3),
										"docFetches":   uint64(2),
										"fieldFetches": uint64(6),
										"indexFetches": uint64(2),
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

// Each field's index carries its own scoring parameters, so turning length normalisation off for
// one field does not turn it off for the other. Both documents hold the term twice, and only the
// field whose index still normalises penalises the longer one.
func TestBM25MultiFieldQuery_UsesEachIndexsOwnParameters(t *testing.T) {
	assert := func(fields string, expected ...string) {
		testUtils.ExecuteTestCase(t, testUtils.TestCase{
			Actions: []any{
				&action.AddCollection{
					SDL: `
						type Article {
							name: String
							title: String @index(kind: BM25, options: {b: 0})
							body: String @index(kind: BM25, options: {b: 1})
						}
					`,
				},
				&action.AddDoc{Doc: `{"name": "padded", ` +
					`"title": "alpha alpha beta gamma delta epsilon", ` +
					`"body": "alpha alpha beta gamma delta epsilon"}`},
				&action.AddDoc{Doc: `{"name": "terse", "title": "alpha alpha", "body": "alpha alpha"}`},
				&action.Request{
					Request: `query {
						Article(order: {_alias: {rank: DESC}}) {
							name
							rank: _bm25(query: "alpha", fields: ` + fields + `)
						}
					}`,
					Results: names(expected...),
				},
			},
		})
	}

	// Scored only on title, whose index does not normalise, the two documents hold the term
	// equally often and the tie is broken by document short ID, which is insertion order.
	assert(`["title"]`, "padded", "terse")
	// Scored only on body, whose index normalises hardest, the padded document is penalised.
	assert(`["body"]`, "terse", "padded")
}

// Every field named has to carry a BM25 index, and the one that does not is named in the error.
func TestBM25MultiFieldQuery_WithAFieldMissingItsIndex_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: bm25MultiFieldActions(&action.Request{
			Request: `query {
				Article {
					rank: _bm25(query: "sharding", fields: ["title", "name"])
				}
			}`,
			ExpectedError: planner.NewErrNoBM25Index("Article", "name").Error(),
		}),
	}

	testUtils.ExecuteTestCase(t, test)
}

// The fields argument carries its own syntax inside a string, so what a typed argument would have
// rejected when the request was validated is rejected when it is parsed instead.
func TestBM25MultiFieldQuery_WithMalformedFields_ReturnsError(t *testing.T) {
	assert := func(fields string, expectedError string) {
		testUtils.ExecuteTestCase(t, testUtils.TestCase{
			Actions: bm25MultiFieldActions(&action.Request{
				Request: `query {
					Article {
						rank: _bm25(query: "sharding", fields: ` + fields + `)
					}
				}`,
				ExpectedError: expectedError,
			}),
		})
	}

	assert(`[]`, parser.ErrBm25NoFields.Error())
	assert(`["^4"]`, parser.NewErrInvalidBm25Field("^4").Error())
	assert(`["title^"]`, parser.NewErrInvalidBm25Boost("title^", "").Error())
	assert(`["title^high"]`, parser.NewErrInvalidBm25Boost("title^high", "high").Error())
	assert(`["title^-1"]`, parser.NewErrInvalidBm25Boost("title^-1", "-1").Error())
	assert(`["title^NaN"]`, parser.NewErrInvalidBm25Boost("title^NaN", "NaN").Error())
	assert(`["title^Inf"]`, parser.NewErrInvalidBm25Boost("title^Inf", "Inf").Error())
	assert(`["title^2^3"]`, parser.NewErrInvalidBm25Boost("title^2^3", "2^3").Error())
	assert(`["title", "title^2"]`, parser.NewErrDuplicateBm25Field("title").Error())
}
