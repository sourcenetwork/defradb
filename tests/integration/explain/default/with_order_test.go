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

var orderPattern = dataMap{
	"explain": dataMap{
		"selectTopNode": dataMap{
			"orderNode": dataMap{
				"selectNode": dataMap{
					"scanNode": dataMap{},
				},
			},
		},
	},
}

func TestDefaultExplainRequestWithAscendingOrderOnParent(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with ascending order on parent.",

		Request: `query @explain {
			Author(order: {age: ASC}) {
				name
				age
			}
		}`,

		ExpectedPatterns: []dataMap{orderPattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "orderNode",
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"orderings": []dataMap{
						{
							"direction": "ASC",
							"fields": []string{
								"age",
							},
						},
					},
				},
			},
		},
	}

	explainUtils.RunExplainTest(t, test)
}

func TestDefaultExplainRequestWithMultiOrderFieldsOnParent(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with multiple order fields on parent.",

		Request: `query @explain {
			Author(order: {name: ASC, age: DESC}) {
				name
				age
			}
		}`,

		ExpectedPatterns: []dataMap{orderPattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "orderNode",
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"orderings": []dataMap{
						{
							"direction": "ASC",
							"fields": []string{
								"name",
							},
						},
						{
							"direction": "DESC",
							"fields": []string{
								"age",
							},
						},
					},
				},
			},
		},
	}

	explainUtils.RunExplainTest(t, test)
}
