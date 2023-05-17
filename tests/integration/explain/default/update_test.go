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
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) mutation request with update using boolean filter.",

		Request: `mutation @explain {
			update_Author(
				filter: {
					verified: {
						_eq: true
					}
				},
				data: "{\"age\": 59}"
			) {
				_key
				name
				age
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

		ExpectedPatterns: []dataMap{updatePattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "updateNode",
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"data": dataMap{
						"age": float64(59),
					},
					"filter": dataMap{
						"verified": dataMap{
							"_eq": true,
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
	}

	runExplainTest(t, test)
}

func TestDefaultExplainMutationRequestWithUpdateUsingIds(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) mutation request with update using ids.",

		Request: `mutation @explain {
			update_Author(
				ids: [
					"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
					"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
				],
				data: "{\"age\": 59}"
			) {
				_key
				name
				age
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

		ExpectedPatterns: []dataMap{updatePattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "updateNode",
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"data": dataMap{
						"age": float64(59),
					},
					"filter": nil,
					"ids": []string{
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
	}

	runExplainTest(t, test)
}

func TestDefaultExplainMutationRequestWithUpdateUsingId(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) mutation request with update using id.",

		Request: `mutation @explain {
			update_Author(
				id: "bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
				data: "{\"age\": 59}"
			) {
				_key
				name
				age
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

		ExpectedPatterns: []dataMap{updatePattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "updateNode",
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"data": dataMap{
						"age": float64(59),
					},
					"filter": nil,
					"ids": []string{
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
	}

	runExplainTest(t, test)
}

func TestDefaultExplainMutationRequestWithUpdateUsingIdsAndFilter(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) mutation request with update using both ids and filter.",

		Request: `mutation @explain {
			update_Author(
				filter: {
					verified: {
						_eq: true
					}
				},
				ids: [
					"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
					"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
				],
				data: "{\"age\": 59}"
			) {
				_key
				name
				age
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

		ExpectedPatterns: []dataMap{updatePattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "updateNode",
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"data": dataMap{
						"age": float64(59),
					},
					"filter": dataMap{
						"verified": dataMap{
							"_eq": true,
						},
					},
					"ids": []string{
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
	}

	runExplainTest(t, test)
}
