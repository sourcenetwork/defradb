// Copyright 2023 Democratized Data Foundation
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

var orderTypeJoinPattern = dataMap{
	"root": dataMap{
		"scanNode": dataMap{},
	},
	"subType": dataMap{
		"selectTopNode": dataMap{
			"orderNode": dataMap{
				"selectNode": dataMap{
					"scanNode": dataMap{},
				},
			},
		},
	},
}

func TestDefaultExplainRequestWithOrderFieldOnRelatedChild(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with order field on a related child.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author {
						name
						articles(order: {name: DESC}) {
							name
						}
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"typeIndexJoin": orderTypeJoinPattern,
									},
								},
							},
						},
					},
				},

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "orderNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"orderings": []dataMap{
								{
									"direction": "DESC",
									"fields": []string{
										"name",
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

func TestDefaultExplainRequestWithOrderFieldOnParentAndRelatedChild(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with order field on parent and related child.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(order: {name: ASC}) {
						name
						articles(order: {name: DESC}) {
							name
						}
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"orderNode": dataMap{
										"selectNode": dataMap{
											"typeIndexJoin": orderTypeJoinPattern,
										},
									},
								},
							},
						},
					},
				},

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "orderNode",
						OccurancesToSkip:  0,
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"orderings": []dataMap{
								{
									"direction": "ASC",
									"fields": []string{
										"name",
									},
								},
							},
						},
					},
					{
						TargetNodeName:    "orderNode",
						OccurancesToSkip:  1,
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"orderings": []dataMap{
								{
									"direction": "DESC",
									"fields": []string{
										"name",
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

func TestDefaultExplainRequestWhereParentIsOrderedByItsRelatedChild(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request where parent is ordered by it's related child.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(
						order: {
							articles: {name: ASC}
						}
					) {
						articles {
							name
						}
					}
				}`,

				ExpectedError: "Argument \"order\" has invalid value {articles: {name: ASC}}.\nIn field \"articles\": Unknown field.",
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}
