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

	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

var deletePattern = dataMap{
	"explain": dataMap{
		"deleteNode": dataMap{
			"selectTopNode": dataMap{
				"selectNode": dataMap{
					"scanNode": dataMap{},
				},
			},
		},
	},
}

func TestDefaultExplainMutationRequestWithDeleteUsingFilter(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) mutation request with delete using filter.",

		Request: `mutation @explain {
				delete_Author(filter: {name: {_eq: "Shahzad"}}) {
					_key
				}
			}`,

		Docs: map[int][]string{
			2: {
				`{
						"name": "Shahzad",
						"age":  26,
						"verified": true
					}`,
			},
		},

		ExpectedPatterns: []dataMap{deletePattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "deleteNode",
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"filter": dataMap{
						"name": dataMap{
							"_eq": "Shahzad",
						},
					},
					"ids": []string(nil),
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
							"_eq": "Shahzad",
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
	}

	runExplainTest(t, test)
}

func TestDefaultExplainMutationRequestWithDeleteUsingFilterToMatchEverything(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) mutation request with delete using filter to match everything.",

		Request: `mutation @explain {
				delete_Author(filter: {}) {
					DeletedKeyByFilter: _key
				}
			}`,

		Docs: map[int][]string{
			2: {
				`{
						"name": "Shahzad",
						"age":  26,
						"verified": true
					}`,
				`{
						"name": "Shahzad",
						"age":  25,
						"verified": true
					}`,
				`{
						"name": "Shahzad",
						"age":  6,
						"verified": true
					}`,
				`{
						"name": "Shahzad",
						"age":  1,
						"verified": true
					}`,
				`{
						"name": "Shahzad Lone",
						"age":  26,
						"verified": true
					}`,
			},
		},

		ExpectedPatterns: []dataMap{deletePattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "deleteNode",
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"filter": dataMap{},
					"ids":    []string(nil),
				},
			},

			{
				TargetNodeName:    "scanNode",
				IncludeChildNodes: true, // should be last node, so will have no child nodes.
				ExpectedAttributes: dataMap{
					"collectionID":   "3",
					"collectionName": "Author",
					"filter":         dataMap{},
					"spans": []dataMap{
						{
							"end":   "/4",
							"start": "/3",
						},
					},
				},
			},
		},
	}

	runExplainTest(t, test)
}

func TestDefaultExplainMutationRequestWithDeleteUsingId(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) mutation request with delete using id.",

		Request: `mutation @explain {
				delete_Author(id: "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d") {
					_key
				}
			}`,

		Docs: map[int][]string{
			2: {
				// "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
				`{
						"name":     "Shahzad Lone",
						"age":      27,
						"verified": true
					}`,
			},
		},

		ExpectedPatterns: []dataMap{deletePattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "deleteNode",
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"filter": nil,
					"ids": []string{
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
							"end":   "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9e",
							"start": "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
						},
					},
				},
			},
		},
	}

	runExplainTest(t, test)
}

func TestDefaultExplainMutationRequestWithDeleteUsingIds(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) mutation request with delete using ids.",

		Request: `mutation @explain {
				delete_Author(ids: [
					"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
					"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f"
				]) {
					AliasKey: _key
				}
			}`,

		Docs: map[int][]string{
			2: {
				// bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f
				`{
						"name": "Lone",
						"age":  26,
						"verified": false
					}`,
				// "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
				`{
						"name":     "Shahzad Lone",
						"age":      27,
						"verified": true
					}`,
			},
		},

		ExpectedPatterns: []dataMap{deletePattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "deleteNode",
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"filter": nil,
					"ids": []string{
						"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
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
							"end":   "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9e",
							"start": "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
						},
						{
							"end":   "/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67g",
							"start": "/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
						},
					},
				},
			},
		},
	}

	runExplainTest(t, test)
}

func TestDefaultExplainMutationRequestWithDeleteUsingNoIds(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) mutation request with delete using no ids.",

		Request: `mutation @explain {
				delete_Author(ids: []) {
					_key
				}
			}`,

		Docs: map[int][]string{
			2: {
				`{
						"name": "Shahzad",
						"age":  26,
						"verified": true
					}`,
			},
		},

		ExpectedPatterns: []dataMap{deletePattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "deleteNode",
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"filter": nil,
					"ids":    []string{},
				},
			},

			{
				TargetNodeName:    "scanNode",
				IncludeChildNodes: true, // should be last node, so will have no child nodes.
				ExpectedAttributes: dataMap{
					"collectionID":   "3",
					"collectionName": "Author",
					"filter":         nil,
					"spans":          []dataMap{},
				},
			},
		},
	}

	runExplainTest(t, test)
}

func TestDefaultExplainMutationRequestWithDeleteUsingFilterAndIds(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) mutation request with delete using filter and ids.",

		Request: `mutation @explain {
				delete_Author(
				    ids: ["bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d", "test"],
				    filter: {
					    _and: [
					    	{age: {_lt: 26}},
					    	{verified: {_eq: true}},
					    ]
					}
				) {
					_key
				}
			}`,

		Docs: map[int][]string{
			2: {
				// "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
				`{
						"name":     "Shahzad Lone",
						"age":      27,
						"verified": true
					}`,
			},
		},

		ExpectedPatterns: []dataMap{deletePattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "deleteNode",
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"filter": dataMap{
						"_and": []any{
							dataMap{
								"age": dataMap{
									"_lt": int(26),
								},
							},
							dataMap{
								"verified": dataMap{
									"_eq": true,
								},
							},
						},
					},
					"ids": []string{
						"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
						"test",
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
						"_and": []any{
							dataMap{
								"age": dataMap{
									"_lt": int(26),
								},
							},
							dataMap{
								"verified": dataMap{
									"_eq": true,
								},
							},
						},
					},
					"spans": []dataMap{
						{
							"end":   "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9e",
							"start": "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
						},
						{
							"end":   "/3/tesu",
							"start": "/3/test",
						},
					},
				},
			},
		},
	}

	runExplainTest(t, test)
}
