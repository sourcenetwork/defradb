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

func TestDefaultExplainRequestWithAOneToOneJoin(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with a 1-to-1 join.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author {
						OnlyEmail: contact {
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
										"typeIndexJoin": normalTypeJoinPattern,
									},
								},
							},
						},
					},
				},

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
					{
						TargetNodeName:    "typeIndexJoin",
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
						IncludeChildNodes: true, // We care about checking children nodes.
						ExpectedAttributes: dataMap{
							"scanNode": dataMap{
								"filter":         nil,
								"collectionID":   "bafkreig3ohatunyfbhmfgkvs5u7tn36dhaqfufajt5h47s6hi56cw2xm4a",
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
										"collectionID":   "bafkreid5ciigzzwckygf4jskcvr2mkam45p6xpaszii3nycqwsz2fxmqn4",
										"collectionName": "AuthorContact",
										"prefixes": []string{
											"/4",
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

func TestDefaultExplainRequestWithTwoLevelDeepNestedJoins(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Explain (default) request with two level deep nested joins.",

		Actions: []any{
			explainUtils.SchemaForExplainTests,

			testUtils.ExplainRequest{

				Request: `query @explain {
					Author {
						name
						contact {
							email
							address {
								city
							}
						}
					}
				}`,

				ExpectedPatterns: dataMap{
					"explain": dataMap{
						"operationNode": []dataMap{
							{
								"selectTopNode": dataMap{
									"selectNode": dataMap{
										"typeIndexJoin": dataMap{
											"root": dataMap{
												"scanNode": dataMap{},
											},
											"subType": dataMap{
												"selectTopNode": dataMap{
													"selectNode": dataMap{
														"typeIndexJoin": normalTypeJoinPattern,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},

				ExpectedTargets: []testUtils.PlanNodeTargetCase{
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
						TargetNodeName:    "root",
						OccurancesToSkip:  0,
						IncludeChildNodes: true, // We care about checking children nodes.
						ExpectedAttributes: dataMap{
							"scanNode": dataMap{
								"filter":         nil,
								"collectionID":   "bafkreig3ohatunyfbhmfgkvs5u7tn36dhaqfufajt5h47s6hi56cw2xm4a",
								"collectionName": "Author",
								"prefixes": []string{
									"/3",
								},
							},
						},
					},

					// Note: the 1st `subType` will contain the entire rest of the graph so we target
					//       and select only the nodes we care about inside it and not `subType` itself.

					{
						TargetNodeName:    "typeIndexJoin",
						OccurancesToSkip:  1,
						IncludeChildNodes: false,
						ExpectedAttributes: dataMap{
							"direction":   "primary",
							"joinType":    "typeJoinOne",
							"rootName":    immutable.Some("contact"),
							"subTypeName": "address",
						},
					},
					{
						TargetNodeName:    "root",
						OccurancesToSkip:  1,
						IncludeChildNodes: true,
						ExpectedAttributes: dataMap{
							"scanNode": dataMap{
								"filter":         nil,
								"collectionID":   "bafkreid5ciigzzwckygf4jskcvr2mkam45p6xpaszii3nycqwsz2fxmqn4",
								"collectionName": "AuthorContact",
								"prefixes": []string{
									"/4",
								},
							},
						},
					},
					{
						TargetNodeName:    "subType", // The last subType (assert everything under it).
						OccurancesToSkip:  1,
						IncludeChildNodes: true,
						ExpectedAttributes: dataMap{
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"docID":  nil,
									"filter": nil,
									"scanNode": dataMap{
										"filter":         nil,
										"collectionID":   "bafkreienr7idffhd72pfj2oxdob7bpdyzgxwfjalbkpallc42nszjyctqi",
										"collectionName": "ContactAddress",
										"prefixes": []string{
											"/5",
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
