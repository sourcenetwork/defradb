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

var updatePattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"updateNode": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"scanNode": dataMap{},
						},
					},
				},
			},
		},
	},
}

func TestDefaultExplainMutationRequestWithUpdateUsingBooleanFilter(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) mutation request with update using boolean filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain {
					update_Author(
						filter: {
							verified: {
								_eq: true
							}
						},
						input: {age: 59}
					) {
						_docID
						name
						age
					}
				}`,

				ExpectedPatterns: updatePattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "updateNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"input": dataMap{
								"age": int32(59),
							},
							"filter": dataMap{
								"verified": dataMap{
									"_eq": true,
								},
							},
							"docID": []string(nil),
						},
					},
					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be last node, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "3",
							"collectionName": "Author",
							"filter": dataMap{
								"verified": dataMap{
									"_eq": true,
								},
							},
							"prefixes": []string{
								"/3",
							},
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainMutationRequestWithUpdateUsingIds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) mutation request with update using ids.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain {
					update_Author(
						docID: [
							"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
							"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
						],
						input: {age: 59}
					) {
						_docID
						name
						age
					}
				}`,

				ExpectedPatterns: updatePattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "updateNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"input": dataMap{
								"age": int32(59),
							},
							"filter": nil,
							"docID": []string{
								"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
								"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
							},
						},
					},
					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be last node, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "3",
							"collectionName": "Author",
							"filter":         nil,
							"prefixes": []string{
								"/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
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

func TestDefaultExplainMutationRequestWithUpdateUsingId(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) mutation request with update using document id.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain {
					update_Author(
						docID: "bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
						input: {age: 59}
					) {
						_docID
						name
						age
					}
				}`,

				ExpectedPatterns: updatePattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "updateNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"input": dataMap{
								"age": int32(59),
							},
							"filter": nil,
							"docID": []string{
								"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
							},
						},
					},
					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be last node, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "3",
							"collectionName": "Author",
							"filter":         nil,
							"prefixes": []string{
								"/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
							},
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainMutationRequestWithUpdateUsingIdsAndFilter(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) mutation request with update using both ids and filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain {
					update_Author(
						filter: {
							verified: {
								_eq: true
							}
						},
						docID: [
							"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
							"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
						],
						input: {age: 59}
					) {
						_docID
						name
						age
					}
				}`,

				ExpectedPatterns: updatePattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "updateNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"input": dataMap{
								"age": int32(59),
							},
							"filter": dataMap{
								"verified": dataMap{
									"_eq": true,
								},
							},
							"docID": []string{
								"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
								"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
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
								"verified": dataMap{
									"_eq": true,
								},
							},
							"prefixes": []string{
								"/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
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
