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

func TestDefaultExplainRequestWithDescendingOrderOnInnerGroupSelection(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Default explain of groupBy with descending order on inner GROUP shows groupNode plan tree.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					Author(groupBy: [name]) {
						name
						GROUP (order: {age: DESC}){
							age
						}
					}
				}`,

				ExpectedPatterns: groupPattern,

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
											"direction": "DESC",
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
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainRequestWithAscendingOrderOnInnerGroupSelection(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Default explain of groupBy with ascending order on inner GROUP shows groupNode plan tree.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					Author(groupBy: [name]) {
						name
						GROUP (order: {age: ASC}){
							age
						}
					}
				}`,

				ExpectedPatterns: groupPattern,

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
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainRequestWithOrderOnNestedParentGroupByAndOnNestedParentsInnerGroupSelection(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Default explain of nested groupBy with order on parent GROUP and inner GROUP shows groupNode plan.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					Author(groupBy: [name]) {
						name
						GROUP (
							groupBy: [verified],
							order: {verified: ASC}
						){
							verified
							GROUP (order: {age: DESC}) {
								age
							}
						}
					}
				}`,

				ExpectedPatterns: groupPattern,

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
											"fields":    []string{"verified"},
										},
									},
									"groupBy": []string{"verified", "name"},
									"docID":   nil,
									"limit":   nil,
									"filter":  nil,
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
