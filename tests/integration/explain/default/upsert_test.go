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

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `mutation @explain {
					upsert_Author(
						filter: {name: {_eq: "Bob"}},
						add: {name: "Bob", age: 59},
						update: {age: 59}
					) {
						_docID
						name
						age
					}
				}`,

				ExpectedPatterns: upsertPattern,

				ExpectedTargets: []action.PlanNodeTargetCase{
					{
						TargetNodeName:    "upsertNode",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"add": dataMap{
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
							"collectionID":   "bafyreid73sgzodav5hxhrsypjapj6r2uzo7mhm3vqykjhfehj7i5hhksuu",
							"collectionName": "Author",
							"filter": dataMap{
								"name": dataMap{
									"_eq": "Bob",
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
