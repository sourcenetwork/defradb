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

var groupOrderPattern = dataMap{
	"explain": dataMap{
		"selectTopNode": dataMap{
			"orderNode": dataMap{
				"groupNode": dataMap{
					"selectNode": dataMap{
						"scanNode": dataMap{},
					},
				},
			},
		},
	},
}

func TestDefaultExplainRequestWithDescendingOrderOnParentGroupBy(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with order (descending) on parent groupBy.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(
						groupBy: [name],
						order: {name: DESC}
					) {
						name
						_group {
							age
						}
					}
				}`,

				ExpectedPatterns: []dataMap{groupOrderPattern},

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "groupNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"groupByFields": []string{"name"},
							"childSelects": []dataMap{
								emptyChildSelectsAttributeForAuthor,
							},
						},
					},
					{
						TargetNodeName:    "orderNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"orderings": []dataMap{
								{
									"direction": "DESC",
									"fields":    []string{"name"},
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

func TestDefaultExplainRequestWithAscendingOrderOnParentGroupBy(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with order (ascending) on parent groupBy.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(
						groupBy: [name],
						order: {name: ASC}
					) {
						name
						_group {
							age
						}
					}
				}`,

				ExpectedPatterns: []dataMap{groupOrderPattern},

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "groupNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"groupByFields": []string{"name"},
							"childSelects": []dataMap{
								emptyChildSelectsAttributeForAuthor,
							},
						},
					},
					{
						TargetNodeName:    "orderNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"orderings": []dataMap{
								{
									"direction": "ASC",
									"fields":    []string{"name"},
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

func TestDefaultExplainRequestWithOrderOnParentGroupByAndOnInnerGroupSelection(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with order on parent groupBy and inner _group selection.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(
						groupBy: [name],
						order: {name: DESC}
					) {
						name
						_group (order: {age: ASC}){
							age
						}
					}
				}`,

				ExpectedPatterns: []dataMap{groupOrderPattern},

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "groupNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"groupByFields": []string{"name"},
							"childSelects": []dataMap{
								{
									"collectionName": "Author",
									"orderBy": []dataMap{
										{
											"direction": "ASC",
											"fields":    []string{"age"},
										},
									},
									"docIDs":  nil,
									"groupBy": nil,
									"limit":   nil,
									"filter":  nil,
								},
							},
						},
					},
					{
						TargetNodeName:    "orderNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"orderings": []dataMap{
								{
									"direction": "DESC",
									"fields":    []string{"name"},
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
