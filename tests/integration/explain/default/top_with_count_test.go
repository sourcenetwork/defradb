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

var topLevelCountPattern = dataMap{
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
						"countNode": dataMap{},
					},
				},
			},
		},
	},
}

func TestDefaultExplainTopLevelCountRequest(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) top-level count request.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					_count(Author: {})
				}`,

				ExpectedPatterns: topLevelCountPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "bafkreig3ohatunyfbhmfgkvs5u7tn36dhaqfufajt5h47s6hi56cw2xm4a",
							"collectionName": "Author",
							"filter":         nil,
							"prefixes": []string{
								"/3",
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
									"filter":    nil,
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

func TestDefaultExplainTopLevelCountRequestWithFilter(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) top-level count request with filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					_count(
						Author: {
							filter: {
								age: {
									_gt: 26
								}
							}
						}
					)
				}`,

				ExpectedPatterns: topLevelCountPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "bafkreig3ohatunyfbhmfgkvs5u7tn36dhaqfufajt5h47s6hi56cw2xm4a",
							"collectionName": "Author",
							"filter": dataMap{
								"age": dataMap{
									"_gt": int32(26),
								},
							},
							"prefixes": []string{
								"/3",
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
