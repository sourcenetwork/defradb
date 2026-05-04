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

package add

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestMutationAdd_AddAfterDiscard(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some([]state.ClientType{
			state.GoClientType,
			state.HTTPClientType,
			state.CLIClientType,
			state.JSClientType,
			// The C client can return a different error on the second add doc call due to a race condition.
			// https://github.com/sourcenetwork/defradb/issues/4771
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				// Establishing the transaction under test must be done synchronously, otherwise we will
				// end up with multiple transactions due to a clash between the public API and the test
				// framework.
				TransactionID: immutable.Some(1),
				DocMap: map[string]any{
					"name": "John",
					"age":  27,
				},
			},
			&action.DiscardTransaction{
				TransactionID: 1,
			},
			&action.AddDoc{
				TransactionID: immutable.Some(1),
				DocMap: map[string]any{
					"name": "Fred",
					"age":  21,
				},
				// This error will occur if the commit txn action completes before the add document action.
				// It should not impact the test execution.
				ExpectedError: "this transaction has been discarded. Create a new one",
			},
			&action.Request{
				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
