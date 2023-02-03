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

func TestExplainJoinsNew(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain a simple sum query of a One-to-Many realted sub-type with many sources.",

		Request: `query @explain {
			author {
				name
				TotalPages: _sum(
					books: {field: pages},
					articles: {field: pages}
				)
			}
		}`,

		ExpectedPatterns: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"sumNode": dataMap{
							"selectNode": dataMap{
								"parallelNode": []dataMap{
									{
										"typeIndexJoin": dataMap{
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
										},
									},
									{
										"typeIndexJoin": dataMap{
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
										},
									},
								},
							},
						},
					},
				},
			},
		},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{

			{
				TargetNodeName:     "selectTopNode",
				ExpectedAttributes: dataMap{},
			},

			{
				TargetNodeName: "typeIndexJoin",
				ExpectedAttributes: dataMap{
					"joinType":    "typeJoinMany",
					"rootName":    "author",
					"subTypeName": "books",
				},
			},

			{
				TargetNodeName:   "typeIndexJoin",
				OccurancesToSkip: 1,
				ExpectedAttributes: dataMap{
					"joinType":    "typeJoinMany",
					"rootName":    "author",
					"subTypeName": "articles",
				},
			},

			{
				TargetNodeName:    "typeIndexJoin",
				IncludeChildNodes: true,
				ExpectedAttributes: dataMap{
					"joinType": "typeJoinMany",
					"rootName": "author",
					"root": dataMap{
						"scanNode": dataMap{
							"collectionID":   "3",
							"collectionName": "author",
							"filter":         nil,
							"spans": []dataMap{
								{
									"start": "/3",
									"end":   "/4",
								},
							},
						},
					},
					"subTypeName": "books",
					"subType": dataMap{
						"selectTopNode": dataMap{
							"selectNode": dataMap{
								"filter": nil,
								"scanNode": dataMap{
									"collectionID":   "2",
									"collectionName": "book",
									"filter":         nil,
									"spans": []dataMap{
										{
											"start": "/2",
											"end":   "/3",
										},
									},
								},
							},
						},
					},
				},
			},

			{
				TargetNodeName:   "scanNode",
				OccurancesToSkip: 0,
				ExpectedAttributes: dataMap{
					"collectionID":   "3",
					"collectionName": "author",
					"filter":         nil,
					"spans": []dataMap{
						{
							"end":   "/4",
							"start": "/3",
						},
					},
				},
			},

			{
				TargetNodeName:   "scanNode",
				OccurancesToSkip: 1,
				ExpectedAttributes: dataMap{
					"collectionID":   "2",
					"collectionName": "book",
					"filter":         nil,
					"spans": []dataMap{
						{
							"end":   "/3",
							"start": "/2",
						},
					},
				},
			},

			{
				TargetNodeName:   "scanNode",
				OccurancesToSkip: 2,
				ExpectedAttributes: dataMap{
					"collectionID":   "3",
					"collectionName": "author",
					"filter":         nil,
					"spans": []dataMap{
						{
							"end":   "/4",
							"start": "/3",
						},
					},
				},
			},

			{
				// I want to target the last scanNode in the second typeIndexJoin, inorder to navigate there
				// I need to skip 3 occurances / offset target by 2 in a way (look the test below, I marked the target).
				TargetNodeName:   "scanNode",
				OccurancesToSkip: 3,
				ExpectedAttributes: dataMap{
					"collectionID":   "1",
					"collectionName": "article",
					"filter":         nil,
					"spans": []dataMap{
						{
							"end":   "/2",
							"start": "/1",
						},
					},
				},
			},
		},
	}

	executeExplainTestCase(t, test)
}

func TestExplainJoinsOld(t *testing.T) {
	test := testUtils.RequestTestCase{

		Description: "Explain a simple sum query of a One-to-Many realted sub-type with many sources.",

		Request: `query @explain {
			author {
				name
				TotalPages: _sum(
					books: {field: pages},
					articles: {field: pages}
				)
			}
		}`,

		Docs: map[int][]string{
			// articles
			0: {
				`{
					"name": "After Guantánamo, Another Injustice",
					"author_id": "bae-25fafcc7-f251-58c1-9495-ead73e676fb8",
					"pages": 2
				}`,
				`{
					"name": "To my dear readers",
					"author_id": "bae-3dddb519-3612-5e43-86e5-49d6295d4f84",
					"pages": 11
				}`,
				`{
					"name": "Twinklestar's Favourite Xmas Cookie",
					"author_id": "bae-3dddb519-3612-5e43-86e5-49d6295d4f84",
					"pages": 31
				}`,
			},

			// books
			1: {
				`{
					"name": "Painted House",
					"author_id": "bae-25fafcc7-f251-58c1-9495-ead73e676fb8",
					"pages": 22
				}`,
				`{
					"name": "A Time for Mercy",
					"author_id": "bae-25fafcc7-f251-58c1-9495-ead73e676fb8",
					"pages": 101
				}`,
				`{
					"name": "Theif Lord",
					"author_id": "bae-3dddb519-3612-5e43-86e5-49d6295d4f84",
					"pages": 321
				}`,
			},

			// authors
			2: {
				// _key: "bae-25fafcc7-f251-58c1-9495-ead73e676fb8"
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"contact_id": "bae-1fe427b8-ab8d-56c3-9df2-826a6ce86fed"
				}`,
				// _key: "bae-3dddb519-3612-5e43-86e5-49d6295d4f84"
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false,
					"contact_id": "bae-c0960a29-b704-5c37-9c2e-59e1249e4559"
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"sumNode": dataMap{
							"selectNode": dataMap{
								"filter": nil,
								"parallelNode": []dataMap{
									{
										"typeIndexJoin": dataMap{
											"joinType": "typeJoinMany",
											"root": dataMap{
												"scanNode": dataMap{ //------------ scan nodes to skip to get here = 0
													"collectionID":   "3",
													"collectionName": "author",
													"filter":         nil,
													"spans": []dataMap{
														{
															"end":   "/4",
															"start": "/3",
														},
													},
												},
											},
											"rootName": "author",
											"subType": dataMap{
												"selectTopNode": dataMap{
													"selectNode": dataMap{
														"filter": nil,
														"scanNode": dataMap{ //---- scan nodes to skip to get here = 1
															"collectionID":   "2",
															"collectionName": "book",
															"filter":         nil,
															"spans": []dataMap{
																{
																	"end":   "/3",
																	"start": "/2",
																},
															},
														},
													},
												},
											},
											"subTypeName": "books",
										},
									},
									{
										"typeIndexJoin": dataMap{
											"joinType": "typeJoinMany",
											"root": dataMap{
												"scanNode": dataMap{ //------------ scan nodes to skip to get here = 2
													"collectionID":   "3",
													"collectionName": "author",
													"filter":         nil,
													"spans": []dataMap{
														{
															"end":   "/4",
															"start": "/3",
														},
													},
												},
											},
											"rootName": "author",
											"subType": dataMap{
												"selectTopNode": dataMap{
													"selectNode": dataMap{
														"filter": nil,
														"scanNode": dataMap{ //---- scan nodes to skip to get here = 3
															// In the last case of the new example above we are trying to target this scanNode.
															"collectionID":   "1",
															"collectionName": "article",
															"filter":         nil,
															"spans": []dataMap{
																{
																	"end":   "/2",
																	"start": "/1",
																},
															},
														},
													},
												},
											},
											"subTypeName": "articles",
										},
									},
								},
							},
							"sources": []dataMap{
								{
									"childFieldName": "pages",
									"fieldName":      "books",
									"filter":         nil,
								},

								{
									"childFieldName": "pages",
									"fieldName":      "articles",
									"filter":         nil,
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
