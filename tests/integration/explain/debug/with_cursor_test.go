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

package test_explain_debug

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

var cursorPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"selectTopNode": dataMap{
					"cursorNode": dataMap{
						"selectNode": dataMap{
							"scanNode": dataMap{},
						},
					},
				},
			},
		},
	},
}

var userSchemaWithIndex = &action.AddCollection{
	SDL: `
		type User {
			name: String
			age: Int @index
		}
	`,
}

func TestDebugExplainCursorQueryWithFirst(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			userSchemaWithIndex,

			&action.ExplainRequest{
				Request: `query @explain(type: debug) {
					_cursor {
						User(first: 3, order: {age: ASC}) {
							name
						}
					}
				}`,

				ExpectedPatterns: cursorPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDebugExplainCursorQueryWithFirstAndAfterNull(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			userSchemaWithIndex,

			&action.ExplainRequest{
				Request: `query @explain(type: debug) {
					_cursor {
						User(first: 5, after: null, order: {age: ASC}) {
							name
						}
					}
				}`,

				ExpectedPatterns: cursorPattern,
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}
