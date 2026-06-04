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

func TestDefaultExplainRequestWithLimitAndOffsetOnInnerGroupSelection(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					Author(groupBy: [name]) {
						name
						GROUP(limit: 2, offset: 1) {
							age
						}
					}
				}`,

				ExpectedPatterns: groupPattern,

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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					Author(groupBy: [name]) {
						name
						innerFirstGroup: GROUP(limit: 1, offset: 2) {
							age
						}
						innerSecondGroup: GROUP(limit: 2) {
							age
						}
					}
				}`,

				ExpectedPatterns: groupPattern,

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
