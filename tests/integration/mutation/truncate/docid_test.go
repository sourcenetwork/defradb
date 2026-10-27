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

package truncate

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationTruncateWithDocID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{SDL: `type User { name: String }`},
			&action.AddDoc{CollectionID: 0, Doc: `{"name":"Alice"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"name":"Bob"}`},
			&action.Request{
				Request: `mutation {
					truncate_User(docID: ["{{.DocID0_0}}"])
				}`,
				Results: map[string]any{"truncate_User": true},
			},
			&action.Request{
				Request: `query { User { name } }`,
				Results: map[string]any{
					"User": []map[string]any{{"name": "Bob"}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationTruncateWithScalarDocID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{SDL: `type User { name: String }`},
			&action.AddDoc{CollectionID: 0, Doc: `{"name":"Alice"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"name":"Bob"}`},
			&action.Request{
				Request: `mutation {
					truncate_User(docID: "{{.DocID0_0}}")
				}`,
				Results: map[string]any{"truncate_User": true},
			},
			&action.Request{
				Request: `query { User { name } }`,
				Results: map[string]any{
					"User": []map[string]any{{"name": "Bob"}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationTruncateWithMultipleDocIDs(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{SDL: `type User { name: String }`},
			&action.AddDoc{CollectionID: 0, Doc: `{"name":"Alice"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"name":"Bob"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"name":"Carol"}`},
			&action.Request{
				Request: `mutation {
					truncate_User(docID: ["{{.DocID0_0}}", "{{.DocID0_1}}"])
				}`,
				Results: map[string]any{"truncate_User": true},
			},
			&action.Request{
				Request: `query { User { name } }`,
				Results: map[string]any{
					"User": []map[string]any{{"name": "Carol"}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationTruncateWithUnknownDocID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{SDL: `type User { name: String }`},
			&action.AddDoc{CollectionID: 0, Doc: `{"name":"Alice"}`},
			&action.Request{
				Request: `mutation {
					truncate_User(docID: ["bae-390b4419-fe1c-506b-98bd-20847cdab2d9"])
				}`,
				Results: map[string]any{"truncate_User": true},
			},
			&action.Request{
				Request: `query { User { name } }`,
				Results: map[string]any{
					"User": []map[string]any{{"name": "Alice"}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationTruncateWithDocIDAndFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{SDL: `type User { name: String }`},
			&action.AddDoc{CollectionID: 0, Doc: `{"name":"Alice"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"name":"Bob"}`},
			&action.Request{
				// The docID matches Alice, but the filter only matches Bob, so nothing
				// should be removed.
				Request: `mutation {
					truncate_User(docID: ["{{.DocID0_0}}"], filter: {name: {_eq: "Bob"}})
				}`,
				Results: map[string]any{"truncate_User": true},
			},
			&action.Request{
				Request: `query { User { name } }`,
				Results: map[string]any{
					"User": []map[string]any{{"name": "Alice"}, {"name": "Bob"}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationTruncateRejectsNullDocID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{SDL: `type User { name: String }`},
			&action.AddDoc{CollectionID: 0, Doc: `{"name":"Alice"}`},
			&action.Request{
				Request:       `mutation { truncate_User(docID: null) }`,
				ExpectedError: "truncate docID cannot be null",
			},
			&action.Request{
				Request: `query { User { name } }`,
				Results: map[string]any{
					"User": []map[string]any{{"name": "Alice"}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
