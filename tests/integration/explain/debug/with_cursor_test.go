// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

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
