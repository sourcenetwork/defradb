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

func TestDefaultExplainRequestWithRelatedAndRegularFilter(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with related and regular filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(
						filter: {
							name: {_eq: "John Grisham"},
							books: {name: {_eq: "Painted House"}}
						}
					) {
						name
						age
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"typeIndexJoin": normalTypeJoinPattern,
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
							"docIDs": nil,
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
									"start": "/3",
									"end":   "/4",
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

func TestDefaultExplainRequestWithManyRelatedFilters(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with many related filters.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(
						filter: {
							name: {_eq: "Cornelia Funke"},
							articles: {name: {_eq: "To my dear readers"}},
							books: {name: {_eq: "Theif Lord"}}
						}
					) {
						name
						age
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
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
				},

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName: "selectNode",
						ExpectedAttributes: dataMap{
							"docIDs": nil,
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
									"start": "/3",
									"end":   "/4",
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
