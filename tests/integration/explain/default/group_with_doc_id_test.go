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

func TestDefaultExplainRequestWithDocIDOnParentGroupBy(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with a docID on parent groupBy.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(
						groupBy: [age],
						docID: "bae-6a4c5bc5-b044-5a03-a868-8260af6f2254"
					) {
						age
						_group {
							name
						}
					}
				}`,

				ExpectedPatterns: groupPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "groupNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"groupByFields": []string{"age"},
							"childSelects": []dataMap{
								emptyChildSelectsAttributeForAuthor,
							},
						},
					},
					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "3",
							"collectionName": "Author",
							"filter":         nil,
							"spans": []dataMap{
								{
									"start": "/3/bae-6a4c5bc5-b044-5a03-a868-8260af6f2254",
									"end":   "/3/bae-6a4c5bc5-b044-5a03-a868-8260af6f2255",
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

func TestDefaultExplainRequestWithDocIDsAndFilterOnParentGroupBy(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with docIDs and filter on parent groupBy.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(
						groupBy: [age],
						filter: {age: {_eq: 20}},
						docID: [
							"bae-6a4c5bc5-b044-5a03-a868-8260af6f2254",
							"bae-4ea9d148-13f3-5a48-a0ef-9ffd344caeed"
						]
					) {
						age
						_group {
							name
						}
					}
				}`,

				ExpectedPatterns: groupPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "groupNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"groupByFields": []string{"age"},
							"childSelects": []dataMap{
								emptyChildSelectsAttributeForAuthor,
							},
						},
					},
					{
						TargetNodeName:    "scanNode",
						IncludeChildNodes: true, // should be leaf of it's branch, so will have no child nodes.
						ExpectedAttributes: dataMap{
							"collectionID":   "3",
							"collectionName": "Author",
							"filter": dataMap{
								"age": dataMap{
									"_eq": int32(20),
								},
							},
							"spans": []dataMap{
								{
									"start": "/3/bae-6a4c5bc5-b044-5a03-a868-8260af6f2254",
									"end":   "/3/bae-6a4c5bc5-b044-5a03-a868-8260af6f2255",
								},
								{
									"start": "/3/bae-4ea9d148-13f3-5a48-a0ef-9ffd344caeed",
									"end":   "/3/bae-4ea9d148-13f3-5a48-a0ef-9ffd344caeee",
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
