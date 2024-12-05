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

func TestDefaultExplainRequestWithDocIDsOnInnerGroupSelection(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with docIDs on inner _group.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(
						groupBy: [age]
					) {
						age
						_group(docID: ["bae-6a4c5bc5-b044-5a03-a868-8260af6f2254"]) {
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
								{
									"collectionName": "Author",
									"docID":          []string{"bae-6a4c5bc5-b044-5a03-a868-8260af6f2254"},
									"filter":         nil,
									"groupBy":        nil,
									"limit":          nil,
									"orderBy":        nil,
								},
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
