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
						name: String @trigramIndex
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
						Name:            "User_name_ASC",
						ID:              1,
						Kind:            client.IndexKindTrigram,
						KindDescription: &client.TrigramIndexDescription{},
						Fields:          []client.IndexedFieldDescription{{Name: "name"}},
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

// The typed collection API is exercised by the shared Go, HTTP, CLI, and C client matrix. It
// backfills and reopens the Trigram kind, maintains entries across live writes, then drops and
// recreates the index through the canonical asynchronous lifecycle.
func TestTrigramIndex_APIBackfillRestartMutationDropAndRecreate(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User { name: String }`,
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
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "name",
				Trigram:      &client.TrigramIndexDescription{},
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{{
					Name:            "User_name_ASC",
					ID:              1,
					Kind:            client.IndexKindTrigram,
					KindDescription: &client.TrigramIndexDescription{},
					Fields:          []client.IndexedFieldDescription{{Name: "name"}},
				}},
			},
			&action.Request{
				Request: `query { User(filter: {name: {_like: "%ana%"}}) { name } }`,
				Results: map[string]any{"User": []map[string]any{{"name": "banana"}}},
			},
			testUtils.Restart{},
			&action.Request{
				Request: `query { User(filter: {name: {_like: "%ana%"}}) { name } }`,
				Results: map[string]any{"User": []map[string]any{{"name": "banana"}}},
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
				Request: `query { User(filter: {name: {_like: "%ell%"}}) { name } }`,
				Results: map[string]any{"User": []map[string]any{{"name": "hello"}}},
			},
			&action.DeleteIndex{CollectionID: 0, IndexName: "User_name_ASC"},
			&action.ListIndexes{CollectionID: 0, ExpectedIndexes: []client.IndexDescription{}},
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "name",
				Trigram:      &client.TrigramIndexDescription{},
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{{
					Name:            "User_name_ASC",
					ID:              2,
					Kind:            client.IndexKindTrigram,
					KindDescription: &client.TrigramIndexDescription{},
					Fields:          []client.IndexedFieldDescription{{Name: "name"}},
				}},
			},
			&action.Request{
				Request: `query { User(filter: {name: {_like: "%ell%"}}) { name } }`,
				Results: map[string]any{"User": []map[string]any{{"name": "hello"}}},
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
						age: Int @trigramIndex
					}
				`,
				ExpectedError: db.NewErrUnsupportedStringIndexFieldType("trigram", client.FieldKind_NILLABLE_INT).Error(),
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
					type User {
						name: String
						alias: String
					}
				`,
			},
			&action.NewIndex{
				CollectionID: 0,
				Fields: []client.IndexedFieldDescription{
					{Name: "name"},
					{Name: "alias"},
				},
				Trigram:       &client.TrigramIndexDescription{},
				ExpectedError: db.NewErrStringIndexRequiresSingleField("trigram", 2).Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestTrigramIndex_WithUnique_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type User { name: String }`,
			},
			&action.NewIndex{
				CollectionID:  0,
				FieldName:     "name",
				Unique:        true,
				Trigram:       &client.TrigramIndexDescription{},
				ExpectedError: db.NewErrStringIndexCannotBeUnique("trigram", "name").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
