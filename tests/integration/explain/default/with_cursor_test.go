// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain_default

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

var userSchemaWithIndex = &action.AddSchema{
	Schema: `
		type User {
			name: String
			age: Int @index
		}
	`,
}

func TestDefaultExplainCursorQueryWithFirstOnly(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			userSchemaWithIndex,

			&action.ExplainRequest{
				Request: `query @explain {
					_cursor {
						User(first: 5, order: {age: ASC}) {
							name
						}
					}
				}`,

				ExpectedPatterns: cursorPattern,

				ExpectedTargets: []action.PlanNodeTargetCase{
					{
						TargetNodeName:    "cursorNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"first":       uint64(5),
							"after":       nil,
							"cursorValue": nil,
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainCursorQueryWithFirstAndAfterNull(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			userSchemaWithIndex,

			&action.ExplainRequest{
				Request: `query @explain {
					_cursor {
						User(first: 3, after: null, order: {age: ASC}) {
							name
						}
					}
				}`,

				ExpectedPatterns: cursorPattern,

				ExpectedTargets: []action.PlanNodeTargetCase{
					{
						TargetNodeName:    "cursorNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"first":       uint64(3),
							"after":       nil,
							"cursorValue": nil,
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}
