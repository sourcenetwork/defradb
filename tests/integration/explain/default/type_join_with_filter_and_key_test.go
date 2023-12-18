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

func TestDefaultExplainRequestWithRelatedAndRegularFilterAndKeys(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with related and regular filter + keys.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(
						filter: {
							name: {_eq: "John Grisham"},
							books: {name: {_eq: "Painted House"}}
						},
						docIDs: [
							"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
							"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f8e"
						]
					) {
						name
						age
					}
				}`,

				ExpectedPatterns: []dataMap{
					{
						"explain": dataMap{
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"typeIndexJoin": normalTypeJoinPattern,
								},
							},
						},
					},
				},

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName: "selectNode",
						ExpectedAttributes: dataMap{
							"docIDs": []string{
								"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
								"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f8e",
							},
							"filter": dataMap{
								"books": dataMap{
									"name": dataMap{
										"_eq": "Painted House",
									},
								},
							},
						},
					},
					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be last node, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "3",
							"collectionName": "Author",
							"filter": dataMap{
								"name": dataMap{
									"_eq": "John Grisham",
								},
							},
							"spans": []dataMap{
								{
									"start": "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
									"end":   "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9e",
								},
								{
									"start": "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f8e",
									"end":   "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f8f",
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

func TestDefaultExplainRequestWithManyRelatedFiltersAndKey(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with many related filters + key.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(
						filter: {
							name: {_eq: "Cornelia Funke"},
							articles: {name: {_eq: "To my dear readers"}},
							books: {name: {_eq: "Theif Lord"}}
						},
						docIDs: ["bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"]
					) {
						name
						age
					}
				}`,

				ExpectedPatterns: []dataMap{
					{
						"explain": dataMap{
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"parallelNode": []dataMap{
										{
											"typeIndexJoin": normalTypeJoinPattern,
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

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName: "selectNode",
						ExpectedAttributes: dataMap{
							"docIDs": []string{
								"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
							},
							"filter": dataMap{
								"articles": dataMap{
									"name": dataMap{
										"_eq": "To my dear readers",
									},
								},
								"books": dataMap{
									"name": dataMap{
										"_eq": "Theif Lord",
									},
								},
							},
						},
					},
					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be last node, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "3",
							"collectionName": "Author",
							"filter": dataMap{
								"name": dataMap{
									"_eq": "Cornelia Funke",
								},
							},
							"spans": []dataMap{
								{
									"start": "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
									"end":   "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9e",
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
