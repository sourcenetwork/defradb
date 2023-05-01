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
	test := explainUtils.ExplainRequestTestCase{
		Description: "Explain (default) mutation request with create.",

		Request: `mutation @explain {
			create_author(data: "{\"name\": \"Shahzad Lone\",\"age\": 27,\"verified\": true}") {
				name
				age
			}
		}`,

		ExpectedPatterns: []dataMap{createPattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "createNode",
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"data": dataMap{
						"age":      float64(27),
						"name":     "Shahzad Lone",
						"verified": true,
					},
				},
			},
		},
	}

	runExplainTest(t, test)
}

func TestDefaultExplainMutationRequestDoesNotCreateDocGivenDuplicate(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{
		Description: "Explain (default) mutation request with create, document exists.",

		Request: `mutation @explain {
			create_author(data: "{\"name\": \"Shahzad Lone\",\"age\": 27}") {
				name
				age
			}
		}`,

		Docs: map[int][]string{
			2: {
				`{
					"name": "Shahzad Lone",
					"age": 27
				}`,
			},
		},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "createNode",
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"data": dataMap{
						"age":  float64(27),
						"name": "Shahzad Lone",
					},
				},
			},
		},
	}

	runExplainTest(t, test)
}
