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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

var createPattern = dataMap{
	"explain": dataMap{
		"createNode": dataMap{
			"selectTopNode": dataMap{
				"selectNode": dataMap{
					"scanNode": dataMap{},
				},
			},
		},
	},
}

func TestDefaultExplainMutationRequestWithCreate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Explain (default) mutation request with create.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain {
					create_Author(input: {name: "Shahzad Lone", age: 27, verified: true}) {
						name
						age
					}
				}`,

				ExpectedPatterns: []dataMap{createPattern},

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "createNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"data": dataMap{
								"age":      int32(27),
								"name":     "Shahzad Lone",
								"verified": true,
							},
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainMutationRequestDoesNotCreateDocGivenDuplicate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Explain (default) mutation request with create, document exists.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain {
					create_Author(input: {name: "Shahzad Lone", age: 27}) {
						name
						age
					}
				}`,

				ExpectedPatterns: []dataMap{createPattern},

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "createNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"data": dataMap{
								"age":  int32(27),
								"name": "Shahzad Lone",
							},
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}
