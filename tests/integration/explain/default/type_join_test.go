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

var normalTypeJoinPattern = dataMap{
	"root": dataMap{
		"scanNode": dataMap{},
	},
	"subType": dataMap{
		"selectTopNode": dataMap{
			"selectNode": dataMap{
				"scanNode": dataMap{},
			},
		},
	},
}

func TestDefaultExplainRequestWith2SingleJoinsAnd1ManyJoin(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			&action.ExplainRequest{

				Request: `query @explain {
					Author {
						OnlyEmail: contact {
							email
						}
						articles {
							name
						}
						contact {
							cell
							email
						}
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"parallelNode": []dataMap{
											{
												"typeIndexJoin": normalTypeJoinPattern,
											},
											{
												"typeIndexJoin": normalTypeJoinPattern,
											},
											{
												"typeIndexJoin": normalTypeJoinPattern,
											},
										},
									},
								},
							},
						},
					},
				},

				ExpectedTargets: []action.PlanNodeTargetCase{
					// 1st join's assertions.
					{
						TargetNodeName:    "typeIndexJoin",
						OccurancesToSkip:  0,
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"direction":   "primary",
							"joinType":    "typeJoinOne",
							"rootName":    immutable.Some("author"),
							"subTypeName": "contact",
						},
					},
					{
						// Note: `root` is not a node but is a special case because for typeIndexJoin we
						//       restructure to show both `root` and `subType` at the same level.
						TargetNodeName:    "root",
						OccurancesToSkip:  0,
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
						OccurancesToSkip:  0,
						IncludeChildNodes: true, // We care about checking children nodes.
						ExpectedAttributes: dataMap{
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"docID":  nil,
									"filter": nil,
									"scanNode": dataMap{
										"filter":         nil,
										"collectionID":   "bafyreie2qrsugrpukipgyuxhdtneyjf4tstssauisjvjfqhps4trc4c2py",
										"collectionName": "AuthorContact",
										"prefixes": []string{
											"/4",
										},
									},
								},
							},
						},
					},

					// 2nd join's assertions (the one to many join).
					{
						TargetNodeName:    "typeIndexJoin",
						OccurancesToSkip:  1,
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
						OccurancesToSkip:  1,
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
						OccurancesToSkip:  1,
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

					// 3rd join's assertions (should be same as 1st one, so after `typeIndexJoin` lets just
					// assert that the `scanNode`s are valid only.
					{
						TargetNodeName:    "typeIndexJoin",
						OccurancesToSkip:  2,
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"direction":   "primary",
							"joinType":    "typeJoinOne",
							"rootName":    immutable.Some("author"),
							"subTypeName": "contact",
						},
					},
					{
						// Note: `root` is not a node but is a special case because for typeIndexJoin we
						//       restructure to show both `root` and `subType` at the same level.
						TargetNodeName:    "scanNode",
						OccurancesToSkip:  4,    // As we encountered 2 `scanNode`s per join.
						IncludeChildNodes: true, // Shouldn't have any.
						ExpectedAttributes: dataMap{
							"filter":         nil,
							"collectionID":   "bafyreid73sgzodav5hxhrsypjapj6r2uzo7mhm3vqykjhfehj7i5hhksuu",
							"collectionName": "Author",
							"prefixes": []string{
								"/3",
							},
						},
					},
					{
						// Note: `subType` is not a node but is a special case because for typeIndexJoin we
						//       restructure to show both `root` and `subType` at the same level.
						TargetNodeName:    "scanNode",
						OccurancesToSkip:  5,    // As we encountered 2 `scanNode`s per join + 1 in the `root` above.
						IncludeChildNodes: true, // Shouldn't have any.
						ExpectedAttributes: dataMap{
							"filter":         nil,
							"collectionID":   "bafyreie2qrsugrpukipgyuxhdtneyjf4tstssauisjvjfqhps4trc4c2py",
							"collectionName": "AuthorContact",
							"prefixes": []string{
								"/4",
							},
						},
					},
				},
			},
		},
	}

	explainUtils.ExecuteTestCase(t, test)
}
