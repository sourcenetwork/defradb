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

func TestDefaultExplainRequestWithDescendingOrderOnInnerGroupSelection(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with order (descending) on inner _group selection.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(groupBy: [name]) {
						name
						_group (order: {age: DESC}){
							age
						}
					}
				}`,

				ExpectedPatterns: groupPattern,

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

		Description: "Explain (default) request with order (ascending) on inner _group selection.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(groupBy: [name]) {
						name
						_group (order: {age: ASC}){
							age
						}
					}
				}`,

				ExpectedPatterns: groupPattern,

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

		Description: "Explain (default) request with order on nested parent groupBy and on nested parent's inner _group.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(groupBy: [name]) {
						name
						_group (
							groupBy: [verified],
							order: {verified: ASC}
						){
							verified
							_group (order: {age: DESC}) {
								age
							}
						}
					}
				}`,

				ExpectedPatterns: groupPattern,

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
