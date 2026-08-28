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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationTruncate(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{SDL: `type User { name: String }`},
			&action.AddDoc{CollectionID: 0, Doc: `{"name":"Alice"}`},
			&action.Request{
				Request: `mutation { truncate_User }`,
				Results: map[string]any{"truncate_User": true},
			},
			&action.Request{
				Request: `query { User { name } }`,
				Results: map[string]any{"User": []map[string]any{}},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationTruncateMustBeStandalone(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{SDL: `type User { name: String }`},
			&action.Request{
				Request: `mutation {
					truncate_User
					add_User(input: {name: "Alice"}) { name }
				}`,
				ExpectedError: "truncate mutation must be the only field in an operation",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationTruncateCannotRunInTransaction(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{SDL: `type User { name: String }`},
			&action.Request{
				TransactionID: immutable.Some(0),
				Request:       `mutation { truncate_User }`,
				ExpectedError: "truncate mutation cannot run in a transaction",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
