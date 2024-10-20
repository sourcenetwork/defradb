// Copyright 2024 Democratized Data Foundation
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

var upsertPattern = dataMap{
	"explain": dataMap{
		"operationNode": []dataMap{
			{
				"upsertNode": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"scanNode": dataMap{},
						},
					},
				},
			},
		},
	},
}

func TestDefaultExplainMutationRequest_WithUpsert_Succeeds(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) mutation request with upsert.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `mutation @explain {
					upsert_Author(
						filter: {name: {_eq: "Bob"}},
						create: {name: "Bob", age: 59},
						update: {age: 59}
					) {
						_docID
						name
						age
					}
				}`,

				ExpectedPatterns: upsertPattern,

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "upsertNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"create": dataMap{
								"name": "Bob",
								"age":  int32(59),
							},
							"update": dataMap{
								"age": int32(59),
							},
							"filter": dataMap{
								"name": dataMap{
									"_eq": "Bob",
								},
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
								"name": dataMap{
									"_eq": "Bob",
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
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}
