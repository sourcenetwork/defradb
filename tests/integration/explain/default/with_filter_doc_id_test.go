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

func TestDefaultExplainRequestWithDocIDFilter(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with docID filter.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(docID: "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d") {
						name
						age
					}
				}`,

				ExpectedPatterns: basicPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName: "selectNode",
						ExpectedAttributes: dataMap{
							"docID": []string{
								"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
							},
							"filter": nil,
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

func TestDefaultExplainRequestWithDocIDsFilterUsingOneID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with docIDs filter using one ID.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(docID: ["bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"]) {
						name
						age
					}
				}`,

				ExpectedPatterns: basicPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName: "selectNode",
						ExpectedAttributes: dataMap{
							"docID": []string{
								"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
							},
							"filter": nil,
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

func TestDefaultExplainRequestWithDocIDsFilterUsingMultipleButDuplicateIDs(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with docIDs filter using multiple but duplicate IDs.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(
						docID: [
							"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
							"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
						]
					) {
						name
						age
					}
				}`,

				ExpectedPatterns: basicPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName: "selectNode",
						ExpectedAttributes: dataMap{
							"docID": []string{
								"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
								"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
							},
							"filter": nil,
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
								"/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
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

func TestDefaultExplainRequestWithDocIDsFilterUsingMultipleUniqueIDs(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with docIDs filter using multiple unique IDs.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(
						docID: [
							"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
							"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f"
						]
					) {
						name
						age
					}
				}`,

				ExpectedPatterns: basicPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName: "selectNode",
						ExpectedAttributes: dataMap{
							"docID": []string{
								"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
								"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
							},
							"filter": nil,
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

func TestDefaultExplainRequestWithMatchingIDFilter(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with a filter to match ID.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(
						filter: {
							_docID: {
								_eq: "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
							}
						}
					) {
						name
						age
					}
				}`,

				ExpectedPatterns: basicPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName: "selectNode",
						ExpectedAttributes: dataMap{
							"docID":  nil,
							"filter": nil,
						},
					},
					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be last node, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "3",
							"collectionName": "Author",
							"filter": dataMap{
								"_docID": dataMap{
									"_eq": "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
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
