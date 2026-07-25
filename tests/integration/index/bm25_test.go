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
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// A BM25 index declared in SDL is created, backfilled over the documents that already exist, and
// reported ready.
func TestBM25Index_DeclaredInSDL_IsCreatedAndBackfilled(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Article {
						body: String @index(kind: BM25)
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"body": "the cat sat on the mat"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"body": "the dog barked"}`,
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "Article_body_ASC",
						ID:     1,
						Kind:   client.IndexKindBM25,
						Fields: []client.IndexedFieldDescription{{Name: "body"}},
					},
				},
				ExpectedStatuses: map[string]client.ActionExecution{
					"Article_body_ASC": {
						Status: client.CompletedActionStatus,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// The scoring parameters given in the directive are stored with the index and come back out of
// the client unchanged.
func TestBM25Index_DeclaredWithOptions_RoundTripsThemThroughTheClient(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Article {
						body: String @index(kind: BM25, options: {k1: 1.2, b: 0.75})
					}
				`,
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:    "Article_body_ASC",
						ID:      1,
						Kind:    client.IndexKindBM25,
						Fields:  []client.IndexedFieldDescription{{Name: "body"}},
						Options: map[string]any{"k1": 1.2, "b": 0.75},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Writing documents through a BM25 index maintains its entries. An update and a delete both
// recompute the terms of the value they are removing and fail if the document is not in the
// index, so a write path that did not store the right keys cannot get through this test.
func TestBM25Index_WritesUpdatesAndDeletes_MaintainEntries(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Article {
						body: String @index(kind: BM25)
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"body": "the cat sat on the mat"}`,
			},
			// A value with no terms of its own gets no entries. Writing and removing it must
			// still work.
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"body": "a ---"}`,
			},
			&action.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc:          `{"body": "the dog barked"}`,
			},
			&action.UpdateDoc{
				CollectionID: 0,
				DocID:        1,
				Doc:          `{"body": "quiet afternoon"}`,
			},
			&action.Request{
				Request: `mutation {
					delete_Article(docID: "{{.DocID0_0}}") {
						_docID
					}
				}`,
				Results: map[string]any{
					"delete_Article": []map[string]any{
						{
							"_docID": "{{.DocID0_0}}",
						},
					},
				},
			},
			&action.Request{
				Request: `query {
					Article {
						body
					}
				}`,
				Results: map[string]any{
					"Article": []map[string]any{
						{
							"body": "quiet afternoon",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestBM25Index_OnNonStringField_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Article {
						views: Int @index(kind: BM25)
					}
				`,
				ExpectedError: db.NewErrBM25IndexFieldNotString("views", "Int").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestBM25Index_OnMoreThanOneField_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Article @index(kind: BM25, includes: [{field: "title"}, {field: "body"}]) {
						title: String
						body: String
					}
				`,
				ExpectedError: db.ErrBM25IndexNotSingleField.Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestBM25Index_WithUnique_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Article {
						body: String @index(kind: BM25, unique: true)
					}
				`,
				ExpectedError: db.ErrBM25IndexUnique.Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestBM25Index_WithUnknownOption_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Article {
						body: String @index(kind: BM25, options: {k2: 1.2})
					}
				`,
				ExpectedError: db.NewErrBM25IndexUnknownOption("k2").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
