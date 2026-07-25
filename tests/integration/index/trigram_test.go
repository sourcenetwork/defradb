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

// A trigram index declared in SDL is created, backfilled over the documents that already exist,
// and reported ready.
func TestTrigramIndex_DeclaredInSDL_IsCreatedAndBackfilled(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String @index(kind: TRIGRAM)
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Islam"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Andy"}`,
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "User_name_ASC",
						ID:     1,
						Kind:   client.IndexKindTrigram,
						Fields: []client.IndexedFieldDescription{{Name: "name"}},
					},
				},
				ExpectedStatuses: map[string]client.ActionExecution{
					"User_name_ASC": {
						Status: client.CompletedActionStatus,
					},
				},
			},
			&action.Request{
				Request: `query {
					User(filter: {name: {_like: "%sla%"}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Islam",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Writing documents through a trigram index maintains its entries. An update and a delete both
// recompute the entries of the value they are removing and fail if any of them is missing, so a
// write path that did not store the right keys cannot get through this test.
func TestTrigramIndex_WritesUpdatesAndDeletes_MaintainEntries(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String @index(kind: TRIGRAM)
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "banana"}`,
			},
			// A value shorter than three bytes has no trigrams and so no index entries. Writing
			// and removing it must still work.
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "hi"}`,
			},
			&action.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc:          `{"name": "peach"}`,
			},
			&action.UpdateDoc{
				CollectionID: 0,
				DocID:        1,
				Doc:          `{"name": "hello"}`,
			},
			&action.Request{
				Request: `mutation {
					delete_User(docID: "{{.DocID0_0}}") {
						_docID
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{
						{
							"_docID": "{{.DocID0_0}}",
						},
					},
				},
			},
			&action.Request{
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "hello",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestTrigramIndex_OnNonStringField_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						age: Int @index(kind: TRIGRAM)
					}
				`,
				ExpectedError: db.NewErrTrigramIndexFieldNotString("age", "Int").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestTrigramIndex_OnMoreThanOneField_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User @index(kind: TRIGRAM, includes: [{field: "name"}, {field: "alias"}]) {
						name: String
						alias: String
					}
				`,
				ExpectedError: db.ErrTrigramIndexNotSingleField.Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestTrigramIndex_WithUnique_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String @index(kind: TRIGRAM, unique: true)
					}
				`,
				ExpectedError: db.ErrTrigramIndexUnique.Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
