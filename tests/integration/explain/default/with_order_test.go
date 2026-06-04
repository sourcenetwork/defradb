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

var orderPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"selectTopNode": dataMap{
					"orderNode": dataMap{
						"selectNode": dataMap{
							"scanNode": dataMap{},
						},
					},
				},
			},
		},
	},
}

func TestDefaultExplainRequestWithAscendingOrderOnParent(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					Author(order: {age: ASC}) {
						name
						age
					}
				}`,

				ExpectedPatterns: orderPattern,

				ExpectedTargets: []action.PlanNodeTargetCase{
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
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainRequestWithMultiOrderFieldsOnParent(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					Author(order: [{name: ASC}, {age: DESC}]) {
						name
						age
					}
				}`,

				ExpectedPatterns: orderPattern,

				ExpectedTargets: []action.PlanNodeTargetCase{
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
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}
