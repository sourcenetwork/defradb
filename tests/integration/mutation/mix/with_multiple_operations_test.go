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

package mix

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// TestMutationMultipleOperationsExecuteInRequestOrder ensures that when a single
// request contains multiple mutation operations, they execute in the order they
// appear in the request (per the GraphQL spec's serial execution requirement).
//
// Documents are keyed in storage by a monotonic insertion sequence, so a default
// (unordered) query returns them in insertion order. This test therefore asserts
// the documents come back in the same order they were declared in the request.
//
// Regression test for https://github.com/sourcenetwork/defradb/issues/5012, where
// the operations were collected into a Go map and executed in randomized order.
func TestMutationMultipleOperationsExecuteInRequestOrder(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.Request{
				Request: `mutation {
					a1: add_User(input: {name: "Alice"}) { name }
					a2: add_User(input: {name: "Bob"}) { name }
					a3: add_User(input: {name: "Carol"}) { name }
					a4: add_User(input: {name: "Dave"}) { name }
					a5: add_User(input: {name: "Eve"}) { name }
				}`,
				Results: map[string]any{
					"a1": []map[string]any{{"name": "Alice"}},
					"a2": []map[string]any{{"name": "Bob"}},
					"a3": []map[string]any{{"name": "Carol"}},
					"a4": []map[string]any{{"name": "Dave"}},
					"a5": []map[string]any{{"name": "Eve"}},
				},
			},
			&action.Request{
				// No order argument: results follow insertion order, which must
				// match the order the mutations were declared above.
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Alice"},
						{"name": "Bob"},
						{"name": "Carol"},
						{"name": "Dave"},
						{"name": "Eve"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestMutationMultipleOperationsViaFragmentSpreadExecuteInRequestOrder ensures the
// execution order follows where a fragment spread appears in the operation, not
// where the fragment is defined in the document. Here `...AddFirst` is spread
// before the inline `second` mutation, even though the fragment definition (and
// thus the `first` field's source location) comes later in the document, so
// `first` must execute before `second`.
//
// Regression test for ordering derived from field source position rather than
// first-encounter traversal order (https://github.com/sourcenetwork/defradb/pull/5013).
func TestMutationMultipleOperationsViaFragmentSpreadExecuteInRequestOrder(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.Request{
				Request: `mutation {
					...AddFirst
					second: add_User(input: {name: "Second"}) { name }
				}
				fragment AddFirst on Mutation {
					first: add_User(input: {name: "First"}) { name }
				}`,
				Results: map[string]any{
					"first":  []map[string]any{{"name": "First"}},
					"second": []map[string]any{{"name": "Second"}},
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
						{"name": "First"},
						{"name": "Second"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
