// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package test_explain_default

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

var groupLimitPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
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
		},
	},
}

func TestDefaultExplainRequestWithLimitAndOffsetOnParentGroupBy(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					Author(
						groupBy: [name],
						limit: 1,
						offset: 1
					) {
						name
						GROUP {
							age
						}
					}
				}`,

				ExpectedPatterns: groupLimitPattern,

				ExpectedTargets: []action.PlanNodeTargetCase{
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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					Author(
						groupBy: [name],
						limit: 1
					) {
						name
						GROUP(limit: 2) {
							age
						}
					}
				}`,

				ExpectedPatterns: groupLimitPattern,

				ExpectedTargets: []action.PlanNodeTargetCase{
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
									"docID":   nil,
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
