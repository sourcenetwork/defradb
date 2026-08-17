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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/internal/planner"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func defaultFullTextDescription() *client.FullTextIndexDescription {
	return &client.FullTextIndexDescription{
		Algorithm: client.FullTextAlgorithmBM25,
		BM25: &client.BM25Params{
			K1: client.DefaultBM25K1,
			B:  client.DefaultBM25B,
		},
	}
}

// The SDL directive produces the typed full-text description, backfills existing documents, and
// exposes a ready index that can answer ranked queries.
func TestBM25Index_DeclaredInSDL_IsCreatedAndBackfilled(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{SDL: `type Article { body: String @fullTextIndex }`},
			&action.AddDoc{Doc: `{"body": "database indexing"}`},
			&action.AddDoc{Doc: `{"body": "database systems"}`},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{{
					Name:            "Article_body_ASC",
					ID:              1,
					Kind:            client.IndexKindFullText,
					KindDescription: defaultFullTextDescription(),
					Fields:          []client.IndexedFieldDescription{{Name: "body"}},
				}},
				ExpectedStatuses: map[string]client.ActionExecution{
					"Article_body_ASC": {Status: client.CompletedActionStatus},
				},
			},
			&action.Request{
				Request: `query {
					Article(order: {_alias: {rank: DESC}}) {
						body
						rank: _bm25(query: "indexing", fields: ["body"])
					}
				}`,
				Results: map[string]any{"Article": []map[string]any{{
					"body": "database indexing", "rank": testUtils.BeNumerically(">", 0),
				}}},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// The collection API uses the same typed description as SDL. Backfill, reopened handles, live
// mutation, and drop all preserve the full-text lifecycle contract.
func TestBM25Index_APIBackfillRestartMutationAndDrop(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{SDL: `type Article { title: String  body: String }`},
			&action.AddDoc{Doc: `{"title": "first", "body": "database indexing"}`},
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "body",
				FullText: &client.FullTextIndexDescription{
					Algorithm: client.FullTextAlgorithmBM25,
					BM25:      &client.BM25Params{K1: 2, B: 0.5},
				},
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{{
					Name: "Article_body_ASC", ID: 1, Kind: client.IndexKindFullText,
					KindDescription: &client.FullTextIndexDescription{
						Algorithm: client.FullTextAlgorithmBM25,
						BM25:      &client.BM25Params{K1: 2, B: 0.5},
					},
					Fields: []client.IndexedFieldDescription{{Name: "body"}},
				}},
			},
			testUtils.Restart{},
			&action.UpdateDoc{DocID: 0, Doc: `{"body": "replication gossip"}`},
			&action.Request{
				Request: `query {
					Article {
						title
						rank: _bm25(query: "replication", fields: ["body"])
					}
				}`,
				Results: map[string]any{"Article": []map[string]any{{
					"title": "first", "rank": testUtils.BeNumerically(">", 0),
				}}},
			},
			&action.DeleteIndex{CollectionID: 0, IndexName: "Article_body_ASC"},
			&action.ListIndexes{CollectionID: 0, ExpectedIndexes: []client.IndexDescription{}},
			&action.Request{
				Request:       `query { Article { rank: _bm25(query: "replication", fields: ["body"]) } }`,
				ExpectedError: planner.NewErrNoBM25Index("Article", "body").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// The shared integration client matrix runs this request through the Go, HTTP, CLI, and C
// collection APIs. In particular, an empty typed description must survive each transport and be
// normalized to the canonical BM25 defaults returned by ListIndexes.
func TestBM25Index_APIDefaultsRoundTrip(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{SDL: `type Article { body: String }`},
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "body",
				FullText:     &client.FullTextIndexDescription{},
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{{
					Name:            "Article_body_ASC",
					ID:              1,
					Kind:            client.IndexKindFullText,
					KindDescription: defaultFullTextDescription(),
					Fields:          []client.IndexedFieldDescription{{Name: "body"}},
				}},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestBM25Index_InvalidTypedRequestsReturnErrors(t *testing.T) {
	tests := []struct {
		name     string
		action   *action.NewIndex
		expected string
	}{
		{
			name: "multiple fields",
			action: &action.NewIndex{
				CollectionID: 0,
				Fields:       []client.IndexedFieldDescription{{Name: "body"}, {Name: "title"}},
				FullText:     defaultFullTextDescription(),
			},
			expected: db.NewErrStringIndexRequiresSingleField("full-text", 2).Error(),
		},
		{
			name: "unique",
			action: &action.NewIndex{
				CollectionID: 0, FieldName: "body", Unique: true, FullText: defaultFullTextDescription(),
			},
			expected: db.NewErrStringIndexCannotBeUnique("full-text", "body").Error(),
		},
		{
			name: "non-string field",
			action: &action.NewIndex{
				CollectionID: 0, FieldName: "age", FullText: defaultFullTextDescription(),
			},
			expected: db.NewErrUnsupportedStringIndexFieldType(
				"full-text",
				client.FieldKind_NILLABLE_INT,
			).Error(),
		},
		{
			name: "invalid b",
			action: &action.NewIndex{
				CollectionID: 0,
				FieldName:    "body",
				FullText: &client.FullTextIndexDescription{
					Algorithm: client.FullTextAlgorithmBM25,
					BM25:      &client.BM25Params{K1: 1.2, B: 2},
				},
			},
			expected: db.NewErrInvalidBM25Parameter("b", 2).Error(),
		},
		{
			name: "mutually exclusive kind descriptions",
			action: &action.NewIndex{
				CollectionID: 0,
				FieldName:    "body",
				FullText:     defaultFullTextDescription(),
				Trigram:      &client.TrigramIndexDescription{},
			},
			expected: db.ErrMultipleIndexKindDescriptions.Error(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.action.ExpectedError = test.expected
			testUtils.ExecuteTestCase(t, testUtils.TestCase{Actions: []any{
				&action.AddCollection{SDL: `type Article { title: String  body: String  age: Int }`},
				test.action,
			}})
		})
	}
}
