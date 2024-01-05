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

var groupLimitPattern = dataMap{
	"explain": dataMap{
		"selectTopNode": dataMap{
			"limitNode": dataMap{
				"groupNode": dataMap{
					"selectNode": dataMap{
						"scanNode": dataMap{},
					},
				},
			},
		},
	},
}

func TestDefaultExplainRequestWithLimitAndOffsetOnParentGroupBy(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with limit and offset on parent groupBy.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(
						groupBy: [name],
						limit: 1,
						offset: 1
					) {
						name
						_group {
							age
						}
					}
				}`,

				ExpectedPatterns: []dataMap{groupLimitPattern},

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "groupNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"groupByFields": []string{"name"},
							"childSelects": []dataMap{
								emptyChildSelectsAttributeForAuthor,
							},
						},
					},
					{
						TargetNodeName:    "limitNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"limit":  uint64(1),
							"offset": uint64(1),
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}

func TestDefaultExplainRequestWithLimitOnParentGroupByAndInnerGroupSelection(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with limit and offset on parent groupBy and inner _group selection.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author(
						groupBy: [name],
						limit: 1
					) {
						name
						_group(limit: 2) {
							age
						}
					}
				}`,

				ExpectedPatterns: []dataMap{groupLimitPattern},

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
										"offset": uint64(0),
									},
									"orderBy": nil,
									"docIDs":  nil,
									"groupBy": nil,
									"filter":  nil,
								},
							},
						},
					},
					{
						TargetNodeName:    "limitNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"limit":  uint64(1),
							"offset": uint64(0),
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}
