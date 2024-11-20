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

	"github.com/sourcenetwork/immutable"

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

		Description: "Explain (default) request with 2 single joins and 1 many join.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

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

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
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
								"collectionID":   "3",
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
										"collectionID":   "4",
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
								"collectionID":   "3",
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
										"collectionID":   "1",
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
							"collectionID":   "3",
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
							"collectionID":   "4",
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
