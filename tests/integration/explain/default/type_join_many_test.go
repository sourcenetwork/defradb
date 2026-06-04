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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

func TestDefaultExplainRequestWithAOneToManyJoin(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					Author {
						articles {
							name
						}
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"typeIndexJoin": normalTypeJoinPattern,
									},
								},
							},
						},
					},
				},

				ExpectedTargets: []action.PlanNodeTargetCase{
					{
						TargetNodeName:    "typeIndexJoin",
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"joinType":    "typeJoinMany",
							"rootName":    immutable.Some("author"),
							"subTypeName": "articles",
						},
					},
					{
						// Note: `root` is not a node but is a special case because for typeIndexJoin we
						//       restructure to show both `root` and `subType` at the same level.
						TargetNodeName:    "root",
						IncludeChildNodes: true, // We care about checking children nodes.
						ExpectedAttributes: dataMap{
							"scanNode": dataMap{
								"filter":         nil,
								"collectionID":   "bafyreid73sgzodav5hxhrsypjapj6r2uzo7mhm3vqykjhfehj7i5hhksuu",
								"collectionName": "Author",
								"prefixes": []string{
									"/3",
								},
							},
						},
					},
					{
						// Note: `subType` is not a node but is a special case because for typeIndexJoin we
						//       restructure to show both `root` and `subType` at the same level.
						TargetNodeName:    "subType",
						IncludeChildNodes: true, // We care about checking children nodes.
						ExpectedAttributes: dataMap{
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"docID":  nil,
									"filter": nil,
									"scanNode": dataMap{
										"filter":         nil,
										"collectionID":   "bafyreidpjzf5kytlap2nilnnemywjmwt74hr56e72llri2z7c6w7un7sje",
										"collectionName": "Article",
										"prefixes": []string{
											"/1",
										},
									},
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
