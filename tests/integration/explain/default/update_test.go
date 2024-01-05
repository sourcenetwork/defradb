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
		"updateNode": dataMap{
			"selectTopNode": dataMap{
				"selectNode": dataMap{
					"scanNode": dataMap{},
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

				ExpectedPatterns: []dataMap{updatePattern},

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "updateNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"data": dataMap{
								"age": int32(59),
							},
							"filter": dataMap{
								"verified": dataMap{
									"_eq": true,
								},
							},
							"docIDs": []string(nil),
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
							"spans": []dataMap{
								{
									"end":   "/4",
									"start": "/3",
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

func TestDefaultExplainMutationRequestWithUpdateUsingIds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) mutation request with update using ids.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain {
					update_Author(
						docIDs: [
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

				ExpectedPatterns: []dataMap{updatePattern},

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "updateNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"data": dataMap{
								"age": int32(59),
							},
							"filter": nil,
							"docIDs": []string{
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
							"spans": []dataMap{
								{
									"end":   "/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67g",
									"start": "/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
								},
								{
									"end":   "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9e",
									"start": "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
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

				ExpectedPatterns: []dataMap{updatePattern},

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "updateNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"data": dataMap{
								"age": int32(59),
							},
							"filter": nil,
							"docIDs": []string{
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
							"spans": []dataMap{
								{
									"start": "/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
									"end":   "/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67g",
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
						docIDs: [
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

				ExpectedPatterns: []dataMap{updatePattern},

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "updateNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"data": dataMap{
								"age": int32(59),
							},
							"filter": dataMap{
								"verified": dataMap{
									"_eq": true,
								},
							},
							"docIDs": []string{
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
							"spans": []dataMap{
								{
									"start": "/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
									"end":   "/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67g",
								},
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
