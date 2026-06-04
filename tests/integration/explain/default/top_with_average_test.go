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

var topLevelAveragePattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"topLevelNode": []dataMap{
					{
						"selectTopNode": dataMap{
							"selectNode": dataMap{
								"scanNode": dataMap{},
							},
						},
					},
					{
						"sumNode": dataMap{},
					},
					{
						"countNode": dataMap{},
					},
					{
						"averageNode": dataMap{},
					},
				},
			},
		},
	},
}

func TestDefaultExplainTopLevelAverageRequest(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					AVG(
						Author: {
							field: age
						}
					)
				}`,

				ExpectedPatterns: topLevelAveragePattern,

				ExpectedTargets: []action.PlanNodeTargetCase{
					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "bafyreid73sgzodav5hxhrsypjapj6r2uzo7mhm3vqykjhfehj7i5hhksuu",
							"collectionName": "Author",
							"filter": dataMap{
								"age": dataMap{
									"_neq": nil,
								},
							},
							"prefixes": []string{
								"/3",
							},
						},
					},
					{
						TargetNodeName:    "sumNode",
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"sources": []dataMap{
								{
									"childFieldName": "age",
									"fieldName":      "Author",
									"filter": dataMap{
										"age": dataMap{
											"_neq": nil,
										},
									},
								},
							},
						},
					},
					{
						TargetNodeName:    "countNode",
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"sources": []dataMap{
								{
									"fieldName": "Author",
									"filter": dataMap{
										"age": dataMap{
											"_neq": nil,
										},
									},
								},
							},
						},
					},
					{
						TargetNodeName:     "averageNode",
						IncludeChildNodes:  true,      // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{}, // no attributes
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainTopLevelAverageRequestWithFilter(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					AVG(
						Author: {
							field: age,
							filter: {
								age: {
									_gt: 26
								}
							}
						}
					)
				}`,

				ExpectedPatterns: topLevelAveragePattern,

				ExpectedTargets: []action.PlanNodeTargetCase{
					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "bafyreid73sgzodav5hxhrsypjapj6r2uzo7mhm3vqykjhfehj7i5hhksuu",
							"collectionName": "Author",
							"filter": dataMap{
								"age": dataMap{
									"_gt":  int32(26),
									"_neq": nil,
								},
							},
							"prefixes": []string{
								"/3",
							},
						},
					},
					{
						TargetNodeName:    "sumNode",
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"sources": []dataMap{
								{
									"childFieldName": "age",
									"fieldName":      "Author",
									"filter": dataMap{
										"age": dataMap{
											"_gt":  int32(26),
											"_neq": nil,
										},
									},
								},
							},
						},
					},
					{
						TargetNodeName:    "countNode",
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"sources": []dataMap{
								{
									"fieldName": "Author",
									"filter": dataMap{
										"age": dataMap{
											"_gt":  int32(26),
											"_neq": nil,
										},
									},
								},
							},
						},
					},
					{
						TargetNodeName:     "averageNode",
						IncludeChildNodes:  true,      // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{}, // no attributes
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}
