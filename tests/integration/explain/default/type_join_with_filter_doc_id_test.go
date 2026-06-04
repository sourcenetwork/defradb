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

func TestDefaultExplainRequestWithRelatedAndRegularFilterAndDocIDs(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					Author(
						filter: {
							name: {_eq: "John Grisham"},
							books: {name: {_eq: "Painted House"}}
						},
						docID: [
							"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
							"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f8e"
						]
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

				ExpectedTargets: []action.PlanNodeTargetCase{
					{
						TargetNodeName: "selectNode",
						ExpectedAttributes: dataMap{
							"docID": []string{
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
							"collectionID":   "bafyreid73sgzodav5hxhrsypjapj6r2uzo7mhm3vqykjhfehj7i5hhksuu",
							"collectionName": "Author",
							"filter": dataMap{
								"name": dataMap{
									"_eq": "John Grisham",
								},
							},
							"prefixes": []string{
								"/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
								"/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f8e",
							},
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainRequestWithManyRelatedFiltersAndDocID(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					Author(
						filter: {
							name: {_eq: "Cornelia Funke"},
							articles: {name: {_eq: "To my dear readers"}},
							books: {name: {_eq: "Theif Lord"}}
						},
						docID: ["bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"]
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

				ExpectedTargets: []action.PlanNodeTargetCase{
					{
						TargetNodeName: "selectNode",
						ExpectedAttributes: dataMap{
							"docID": []string{
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
							"collectionID":   "bafyreid73sgzodav5hxhrsypjapj6r2uzo7mhm3vqykjhfehj7i5hhksuu",
							"collectionName": "Author",
							"filter": dataMap{
								"name": dataMap{
									"_eq": "Cornelia Funke",
								},
							},
							"prefixes": []string{
								"/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
							},
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}
