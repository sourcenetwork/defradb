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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

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

				ExpectedTargets: []action.PlanNodeTargetCase{
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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

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

				ExpectedTargets: []action.PlanNodeTargetCase{
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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

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
