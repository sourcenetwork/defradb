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
