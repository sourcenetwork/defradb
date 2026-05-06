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

package update

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

func TestMutationUpdate_ConcurrentWrite(t *testing.T) {
	test := testUtils.TestCase{
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
			&action.Async{
				Child: &action.UpdateDoc{
					TransactionID: immutable.Some(1),
					Doc: `{
						"name": "Fred",
						"age": 21
					}`,
				},
			},
			&action.Async{
				Child: &action.UpdateDoc{
					TransactionID: immutable.Some(1),
					Doc: `{
						"name": "Shahzad",
						"age": 31
					}`,
				},
			},
			&action.Await{},
			&action.CommitTransaction{
				TransactionID: 1,
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
					"Users": testUtils.AnyOf(
						// The results must never be a mix of Fred and Shahzad,
						// it must always be either Fred,21 *or* Shahzad,31 depending on
						// execution order.
						[]map[string]any{
							{
								"name": "Fred",
								"age":  int64(21),
							},
						},
						[]map[string]any{
							{
								"name": "Shahzad",
								"age":  int64(31),
							},
						},
					),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_ConcurrentCommit(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some([]state.ClientType{
			state.GoClientType,
			state.HTTPClientType,
			state.CLIClientType,
			state.JSClientType,
			// The C client can return a different error on the update doc call due to a race condition.
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
			&action.Async{
				Child: &action.UpdateDoc{
					TransactionID: immutable.Some(1),
					Doc: `{
						"name": "Fred",
						"age": 21
					}`,
					// This error will occur if the commit txn action completes before the update document action.
					// It should not impact the test execution.
					IgnoreError: "this transaction has been discarded. Create a new one",
				},
			},
			&action.Async{
				Child: &action.CommitTransaction{
					TransactionID: 1,
				},
			},
			&action.Await{},
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
					"Users": testUtils.AnyOf(
						// The results must never be a mix of John and Fred,
						// it must always be either John,27 *or* Fred,21 depending on
						// execution order.
						[]map[string]any{
							{
								"name": "John",
								"age":  int64(27),
							},
						},
						[]map[string]any{
							{
								"name": "Fred",
								"age":  int64(21),
							},
						},
					),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationUpdate_ConcurrentDiscard(t *testing.T) {
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
			&action.Async{
				Child: &action.AddDoc{
					TransactionID: immutable.Some(1),
					DocMap: map[string]any{
						"name": "Fred",
						"age":  21,
					},
					// This error will occur if the commit txn action completes before the add document action.
					// It should not impact the test execution.
					IgnoreError: "this transaction has been discarded. Create a new one",
				},
			},
			&action.Async{
				Child: &action.DiscardTransaction{
					TransactionID: 1,
				},
			},
			&action.Await{},
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
