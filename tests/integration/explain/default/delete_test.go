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

var deletePattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"deleteNode": dataMap{
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

func TestDefaultExplainMutationRequestWithDeleteUsingFilter(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) mutation request with delete using filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain {
					delete_Author(filter: {name: {_eq: "Shahzad"}}) {
						_docID
					}
				}`,

				ExpectedPatterns: deletePattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "deleteNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"filter": dataMap{
								"name": dataMap{
									"_eq": "Shahzad",
								},
							},
							"docID": []string(nil),
						},
					},

					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be last node, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "bafkreig3ohatunyfbhmfgkvs5u7tn36dhaqfufajt5h47s6hi56cw2xm4a",
							"collectionName": "Author",
							"filter": dataMap{
								"name": dataMap{
									"_eq": "Shahzad",
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

func TestDefaultExplainMutationRequestWithDeleteUsingFilterToMatchEverything(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) mutation request with delete using filter to match everything.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain {
					delete_Author(filter: {}) {
						DeletedKeyByFilter: _docID
					}
				}`,

				ExpectedPatterns: deletePattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "deleteNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"filter": nil,
							"docID":  []string(nil),
						},
					},

					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be last node, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "bafkreig3ohatunyfbhmfgkvs5u7tn36dhaqfufajt5h47s6hi56cw2xm4a",
							"collectionName": "Author",
							"filter":         nil,
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

func TestDefaultExplainMutationRequestWithDeleteUsingId(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) mutation request with delete using document id.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain {
					delete_Author(docID: "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d") {
						_docID
					}
				}`,

				ExpectedPatterns: deletePattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "deleteNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"filter": nil,
							"docID": []string{
								"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
							},
						},
					},

					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be last node, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "bafkreig3ohatunyfbhmfgkvs5u7tn36dhaqfufajt5h47s6hi56cw2xm4a",
							"collectionName": "Author",
							"filter":         nil,
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

func TestDefaultExplainMutationRequestWithDeleteUsingIds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) mutation request with delete using ids.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain {
					delete_Author(docID: [
						"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
						"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f"
					]) {
						AliasKey: _docID
					}
				}`,

				ExpectedPatterns: deletePattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "deleteNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"filter": nil,
							"docID": []string{
								"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
								"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
							},
						},
					},

					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be last node, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "bafkreig3ohatunyfbhmfgkvs5u7tn36dhaqfufajt5h47s6hi56cw2xm4a",
							"collectionName": "Author",
							"filter":         nil,
							"prefixes": []string{
								"/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
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

func TestDefaultExplainMutationRequestWithDeleteUsingNoIds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) mutation request with delete using no ids.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain {
					delete_Author(docID: []) {
						_docID
					}
				}`,

				ExpectedPatterns: deletePattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "deleteNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"filter": nil,
							"docID":  []string{},
						},
					},

					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be last node, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "bafkreig3ohatunyfbhmfgkvs5u7tn36dhaqfufajt5h47s6hi56cw2xm4a",
							"collectionName": "Author",
							"filter":         nil,
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

func TestDefaultExplainMutationRequestWithDeleteUsingFilterAndIds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) mutation request with delete using filter and ids.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain {
					delete_Author(
						docID: ["bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d", "test"],
						filter: {
							_and: [
								{age: {_lt: 26}},
								{verified: {_eq: true}},
							]
						}
					) {
						_docID
					}
				}`,

				ExpectedPatterns: deletePattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "deleteNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"filter": dataMap{
								"_and": []any{
									dataMap{
										"age": dataMap{
											"_lt": int32(26),
										},
									},
									dataMap{
										"verified": dataMap{
											"_eq": true,
										},
									},
								},
							},
							"docID": []string{
								"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
								"test",
							},
						},
					},

					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be last node, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "bafkreig3ohatunyfbhmfgkvs5u7tn36dhaqfufajt5h47s6hi56cw2xm4a",
							"collectionName": "Author",
							"filter": dataMap{
								"_and": []any{
									dataMap{
										"age": dataMap{
											"_lt": int32(26),
										},
									},
									dataMap{
										"verified": dataMap{
											"_eq": true,
										},
									},
								},
							},
							"prefixes": []string{
								"/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
								"/3/test",
							},
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}
