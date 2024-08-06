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

func TestDefaultExplainRequestWithOnlyLimitOnRelatedChildWithCount(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with limit on related child with count.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author {
						numberOfArts: _count(articles: {})
						articles(limit: 2) {
							name
						}
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"countNode": dataMap{
										"selectNode": dataMap{
											"parallelNode": []dataMap{
												{
													"typeIndexJoin": limitTypeJoinPattern,
												},
												{
													"typeIndexJoin": normalTypeJoinPattern,
												},
											},
										},
									},
								},
							},
						},
					},
				},

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "countNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"sources": []dataMap{
								{
									"fieldName": "articles",
									"filter":    nil,
								},
							},
						},
					},
					{
						TargetNodeName:    "limitNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"limit":  uint64(2),
							"offset": uint64(0),
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainRequestWithLimitArgsOnParentAndRelatedChildWithCount(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with limit args on parent and related child with count.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(limit: 3, offset: 1) {
						numberOfArts: _count(articles: {})
						articles(limit: 2) {
							name
						}
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"limitNode": dataMap{
										"countNode": dataMap{
											"selectNode": dataMap{
												"parallelNode": []dataMap{
													{
														"typeIndexJoin": limitTypeJoinPattern,
													},
													{
														"typeIndexJoin": normalTypeJoinPattern,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "limitNode",
						OccurancesToSkip:  0,
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"limit":  uint64(3),
							"offset": uint64(1),
						},
					},
					{
						TargetNodeName:    "countNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"sources": []dataMap{
								{
									"fieldName": "articles",
									"filter":    nil,
								},
							},
						},
					},
					{
						TargetNodeName:    "limitNode",
						OccurancesToSkip:  1,
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"limit":  uint64(2),
							"offset": uint64(0),
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}
