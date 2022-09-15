// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestExplainSimpleAverageQueryOnArrayField(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain a simple query using average on array field.",
		Query: `query @explain {
					book {
						name
						_avg(chapterPages: {})
					}
				}`,

		Docs: map[int][]string{
			// books
			1: {
				`{
					"name": "Painted House",
					"chapterPages": [1, 22, 33, 44, 55, 66]
				}`,
				`{
					"name": "A Time for Mercy",
					"chapterPages": [0, 22, 101, 321]
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"averageNode": dataMap{
							"countNode": dataMap{
								"sources": []dataMap{
									{
										"filter":    dataMap{"_ne": nil},
										"fieldName": "chapterPages",
									},
								},
								"sumNode": dataMap{
									"sources": []dataMap{
										{
											"filter":         dataMap{"_ne": nil},
											"fieldName":      "chapterPages",
											"childFieldName": nil,
										},
									},
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
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestExplainAverageQueryOnJoinedField(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain a average query on joined field.",
		Query: `query @explain {
					author {
						name
						_avg(books: {field: pages})
					}
				}`,

		Docs: map[int][]string{
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
					"pages": 178
				}`,
				`{
					"name": "Theif Lord",
					"author_id": "bae-3dddb519-3612-5e43-86e5-49d6295d4f84",
					"pages": 321
				 }`,
				`{
					"name": "Incomplete book",
					"author_id": "bae-3dddb519-3612-5e43-86e5-49d6295d4f84",
					"pages": 79
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
						"averageNode": dataMap{
							"countNode": dataMap{
								"sources": []dataMap{
									{
										"fieldName": "books",
										"filter": dataMap{
											"pages": dataMap{
												"_ne": nil,
											},
										},
									},
								},
								"sumNode": dataMap{
									"sources": []dataMap{
										{
											"childFieldName": "pages",
											"fieldName":      "books",
											"filter": dataMap{
												"pages": dataMap{
													"_ne": nil,
												},
											},
										},
									},
									"selectNode": dataMap{
										"filter": nil,
										"typeIndexJoin": dataMap{
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
															"filter": dataMap{
																"pages": dataMap{
																	"_ne": nil,
																},
															},
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

func TestExplainAverageQueryOnMultipleJoinedFieldWithFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain a average query on multiple joined fields with filter.",
		Query: `query @explain {
					author {
						name
						_avg(
							books: {field: pages},
							articles: {field: pages, filter: {pages: {_gt: 3}}}
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
					"pages": 22,
					"chapterPages": [1, 20]
				}`,
				`{
					"name": "A Time for Mercy",
					"author_id": "bae-25fafcc7-f251-58c1-9495-ead73e676fb8",
					"pages": 178,
					"chapterPages": [1, 11, 30, 50, 80, 120, 150]
				}`,
				`{
					"name": "Theif Lord",
					"author_id": "bae-3dddb519-3612-5e43-86e5-49d6295d4f84",
					"pages": 321,
					"chapterPages": [22, 211, 310]
				}`,
				`{
					"name": "Incomplete book",
					"author_id": "bae-3dddb519-3612-5e43-86e5-49d6295d4f84",
					"pages": 79,
					"chapterPages": [1, 22, 33, 44, 55, 66]
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
						"averageNode": dataMap{
							"countNode": dataMap{
								"sources": []dataMap{
									{
										"fieldName": "books",
										"filter": dataMap{
											"pages": dataMap{
												"_ne": nil,
											},
										},
									},
									{
										"fieldName": "articles",
										"filter": dataMap{
											"pages": dataMap{
												"_gt": int64(3),
												"_ne": nil,
											},
										},
									},
								},
								"sumNode": dataMap{
									"sources": []dataMap{
										{
											"childFieldName": "pages",
											"fieldName":      "books",
											"filter": dataMap{
												"pages": dataMap{
													"_ne": nil,
												},
											},
										},
										{
											"childFieldName": "pages",
											"fieldName":      "articles",
											"filter": dataMap{
												"pages": dataMap{
													"_gt": int64(3),
													"_ne": nil,
												},
											},
										},
									},
									"selectNode": dataMap{
										"filter": nil,
										"parallelNode": []dataMap{
											{
												"typeIndexJoin": dataMap{
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
																	"filter": dataMap{
																		"pages": dataMap{
																			"_ne": nil,
																		},
																	},
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
												"typeIndexJoin": dataMap{
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
													"subTypeName": "articles",
													"subType": dataMap{
														"selectTopNode": dataMap{
															"selectNode": dataMap{
																"filter": nil,
																"scanNode": dataMap{
																	"collectionID":   "1",
																	"collectionName": "article",
																	"filter": dataMap{
																		"pages": dataMap{
																			"_gt": int64(3),
																			"_ne": nil,
																		},
																	},
																	"spans": []dataMap{
																		{
																			"start": "/1",
																			"end":   "/2",
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
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
