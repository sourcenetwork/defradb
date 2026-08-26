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

func TestMutationTruncateWithFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{SDL: `type User { name: String }`},
			&action.AddDoc{CollectionID: 0, Doc: `{"name":"Alice"}`},
			&action.AddDoc{CollectionID: 0, Doc: `{"name":"Bob"}`},
			&action.Request{
				Request: `mutation {
					truncate_User(filter: {name: {_eq: "Alice"}}, pruneHistory: true)
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

func TestMutationTruncatePruneHistoryRequiresFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{SDL: `type User { name: String }`},
			&action.Request{
				Request:       `mutation { truncate_User(pruneHistory: true) }`,
				ExpectedError: "prune history requires a filter",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationTruncateRejectsNullFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{SDL: `type User { name: String }`},
			&action.AddDoc{CollectionID: 0, Doc: `{"name":"Alice"}`},
			&action.Request{
				Request:       `mutation { truncate_User(filter: null) }`,
				ExpectedError: "truncate filter cannot be null",
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
