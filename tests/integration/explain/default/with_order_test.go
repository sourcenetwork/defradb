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

		Docs: map[int][]string{
			// authors
			2: {
				// _key: bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// _key: bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},

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

		Docs: map[int][]string{
			// authors
			2: {
				// _key: bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// _key: bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},

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
