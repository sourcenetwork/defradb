// Copyright 2022 Democratized Data Foundation
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

var addPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"addNode": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"scanNode": dataMap{},
						},
					},
				},
			},
		},
	},
}

func TestDefaultExplainMutationRequestWithAdd(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `mutation @explain {
					add_Author(input: {name: "Shahzad Lone", age: 27, verified: true}) {
						name
						age
					}
				}`,

				ExpectedPatterns: addPattern,

				ExpectedTargets: []action.PlanNodeTargetCase{
					{
						TargetNodeName:    "addNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"input": []dataMap{{
								"age":      int32(27),
								"name":     "Shahzad Lone",
								"verified": true,
							}},
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainMutationRequestDoesNotAddDocGivenDuplicate(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `mutation @explain {
					add_Author(input: {name: "Shahzad Lone", age: 27}) {
						name
						age
					}
				}`,

				ExpectedPatterns: addPattern,

				ExpectedTargets: []action.PlanNodeTargetCase{
					{
						TargetNodeName:    "addNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"input": []dataMap{{
								"age":  int32(27),
								"name": "Shahzad Lone",
							}},
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}
