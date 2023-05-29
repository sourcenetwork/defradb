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

func TestDefaultExplainRequestWithDocKeyFilter(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with dockey filter.",

		Request: `query @explain {
			Author(dockey: "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d") {
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

		ExpectedPatterns: []dataMap{basicPattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "scanNode",
				IncludeChildNodes: true, // should be last node, so will have no child nodes.
				ExpectedAttributes: dataMap{
					"collectionID":   "3",
					"collectionName": "Author",
					"filter":         nil,
					"spans": []dataMap{
						{
							"start": "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
							"end":   "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9e",
						},
					},
				},
			},
		},
	}

	explainUtils.RunExplainTest(t, test)
}

func TestDefaultExplainRequestWithDocKeysFilterUsingOneKey(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with dockeys filter using one key.",

		Request: `query @explain {
			Author(dockeys: ["bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"]) {
				name
				age
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

		ExpectedPatterns: []dataMap{basicPattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "scanNode",
				IncludeChildNodes: true, // should be last node, so will have no child nodes.
				ExpectedAttributes: dataMap{
					"collectionID":   "3",
					"collectionName": "Author",
					"filter":         nil,
					"spans": []dataMap{
						{
							"start": "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
							"end":   "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9e",
						},
					},
				},
			},
		},
	}

	explainUtils.RunExplainTest(t, test)
}

func TestDefaultExplainRequestWithDocKeysFilterUsingMultipleButDuplicateKeys(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with dockeys filter using multiple but duplicate keys.",

		Request: `query @explain {
			Author(
				dockeys: [
					"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
					"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
				]
			) {
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

		ExpectedPatterns: []dataMap{basicPattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "scanNode",
				IncludeChildNodes: true, // should be last node, so will have no child nodes.
				ExpectedAttributes: dataMap{
					"collectionID":   "3",
					"collectionName": "Author",
					"filter":         nil,
					"spans": []dataMap{
						{
							"start": "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
							"end":   "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9e",
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

	explainUtils.RunExplainTest(t, test)
}

func TestDefaultExplainRequestWithDocKeysFilterUsingMultipleUniqueKeys(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with dockeys filter using multiple unique keys.",

		Request: `query @explain {
			Author(
				dockeys: [
					"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
					"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f"
				]
			) {
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

		ExpectedPatterns: []dataMap{basicPattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "scanNode",
				IncludeChildNodes: true, // should be last node, so will have no child nodes.
				ExpectedAttributes: dataMap{
					"collectionID":   "3",
					"collectionName": "Author",
					"filter":         nil,
					"spans": []dataMap{
						{
							"start": "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
							"end":   "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9e",
						},
						{
							"start": "/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
							"end":   "/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67g",
						},
					},
				},
			},
		},
	}

	explainUtils.RunExplainTest(t, test)
}

func TestDefaultExplainRequestWithMatchingKeyFilter(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with a filter to match key.",

		Request: `query @explain {
			Author(filter: {_key: {_eq: "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"}}) {
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

		ExpectedPatterns: []dataMap{basicPattern},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "scanNode",
				IncludeChildNodes: true, // should be last node, so will have no child nodes.
				ExpectedAttributes: dataMap{
					"collectionID":   "3",
					"collectionName": "Author",
					"filter": dataMap{
						"_key": dataMap{
							"_eq": "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
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

	explainUtils.RunExplainTest(t, test)
}
