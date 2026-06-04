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

var groupOrderPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
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
		},
	},
}

func TestDefaultExplainRequestWithDescendingOrderOnParentGroupBy(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					Author(
						groupBy: [name],
						order: {name: DESC}
					) {
						name
						GROUP {
							age
						}
					}
				}`,

				ExpectedPatterns: groupOrderPattern,

				ExpectedTargets: []action.PlanNodeTargetCase{
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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					Author(
						groupBy: [name],
						order: {name: ASC}
					) {
						name
						GROUP {
							age
						}
					}
				}`,

				ExpectedPatterns: groupOrderPattern,

				ExpectedTargets: []action.PlanNodeTargetCase{
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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					Author(
						groupBy: [name],
						order: {name: DESC}
					) {
						name
						GROUP (order: {age: ASC}){
							age
						}
					}
				}`,

				ExpectedPatterns: groupOrderPattern,

				ExpectedTargets: []action.PlanNodeTargetCase{
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
									"docID":   nil,
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
