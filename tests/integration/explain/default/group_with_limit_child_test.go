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

func TestDefaultExplainRequestWithLimitAndOffsetOnInnerGroupSelection(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with limit and offset on inner _group selection.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(groupBy: [name]) {
						name
						_group(limit: 2, offset: 1) {
							age
						}
					}
				}`,

				ExpectedPatterns: groupPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "groupNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"groupByFields": []string{"name"},
							"childSelects": []dataMap{
								{
									"collectionName": "Author",
									"limit": dataMap{
										"limit":  uint64(2),
										"offset": uint64(1),
									},
									"docID":   nil,
									"filter":  nil,
									"groupBy": nil,
									"orderBy": nil,
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

func TestDefaultExplainRequestWithLimitAndOffsetOnMultipleInnerGroupSelections(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with limit and offset on multiple inner _group selections.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(groupBy: [name]) {
						name
						innerFirstGroup: _group(limit: 1, offset: 2) {
							age
						}
						innerSecondGroup: _group(limit: 2) {
							age
						}
					}
				}`,

				ExpectedPatterns: groupPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "groupNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"groupByFields": []string{"name"},
							"childSelects": []dataMap{
								{
									"collectionName": "Author",
									"limit": dataMap{
										"limit":  uint64(1),
										"offset": uint64(2),
									},
									"docID":   nil,
									"filter":  nil,
									"groupBy": nil,
									"orderBy": nil,
								},
								{
									"collectionName": "Author",
									"limit": dataMap{
										"limit":  uint64(2),
										"offset": uint64(0),
									},
									"docID":   nil,
									"filter":  nil,
									"groupBy": nil,
									"orderBy": nil,
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
