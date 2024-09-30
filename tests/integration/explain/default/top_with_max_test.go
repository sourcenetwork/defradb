// Copyright 2024 Democratized Data Foundation
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

var topLevelMaxPattern = dataMap{
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
						"maxNode": dataMap{},
					},
				},
			},
		},
	},
}

func TestDefaultExplain_WithTopLevelMaxRequest_Succeeds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) top-level max request.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					_max(
						Author: {
							field: age
						}
					)
				}`,

				ExpectedPatterns: topLevelMaxPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "3",
							"collectionName": "Author",
							"filter":         nil,
							"spans": []dataMap{
								{
									"start": "/3",
									"end":   "/4",
								},
							},
						},
					},
					{
						TargetNodeName:    "maxNode",
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"sources": []dataMap{
								{
									"fieldName":      "Author",
									"childFieldName": "age",
									"filter":         nil,
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

func TestDefaultExplain_WithTopLevelMaxRequestWithFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) top-level max request with filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					_max(
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

				ExpectedPatterns: topLevelMaxPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "3",
							"collectionName": "Author",
							"filter": dataMap{
								"age": dataMap{
									"_gt": int32(26),
								},
							},
							"spans": []dataMap{
								{
									"start": "/3",
									"end":   "/4",
								},
							},
						},
					},
					{
						TargetNodeName:    "maxNode",
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"sources": []dataMap{
								{
									"fieldName":      "Author",
									"childFieldName": "age",
									"filter": dataMap{
										"age": dataMap{
											"_gt": int32(26),
										},
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
