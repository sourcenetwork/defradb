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

func TestExplainQueryOneToOneJoinWithParallelNodeMultipleCounts(t *testing.T) {
	test := testUtils.QueryTestCase{

		Description: "Explain two counts for parallelNode and a 2 level deep join request.",

		Query: `query @explain {
				author {
					name
					numberOfBooks: _count(
						books: {
							filter: {
								name: {
									_eq: "Theif Lord"
								}
							}
						}
					)
					numberOfArticles: _count(articles: {})
					contact {
						email
						address {
							city
						}
					}
				}
			}`,

		Docs: map[int][]string{
			// articles
			0: {
				(`{
					"name": "After Guantánamo, Another Injustice",
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`),
				(`{
					"name": "To my dear readers",
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
					}`),
				(`{
					"name": "Twinklestar's Favourite Xmas Cookie",
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				}`),
			},
			// books
			1: {
				(`{
					"name": "Painted House",
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`),
				(`{
					"name": "A Time for Mercy",
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
					}`),
				(`{
					"name": "Theif Lord",
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				}`),
			},
			// authors
			2: {
				// _key: bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				(`{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"contact_id": "bae-1fe427b8-ab8d-56c3-9df2-826a6ce86fed"
				}`),
				// _key: bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				(`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false,
					"contact_id": "bae-c0960a29-b704-5c37-9c2e-59e1249e4559"
				}`),
			},
			// contact
			3: {
				// _key: bae-1fe427b8-ab8d-56c3-9df2-826a6ce86fed
				// "author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				(`{
					"cell": "5197212301",
					"email": "john_grisham@example.com",
					"address_id": "bae-c8448e47-6cd1-571f-90bd-364acb80da7b"
				}`),

				// _key: bae-c0960a29-b704-5c37-9c2e-59e1249e4559
				// "author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				(`{
					"cell": "5197212302",
					"email": "cornelia_funke@example.com",
					"address_id": "bae-c0960a29-b704-5c37-9c2e-59e1249e4559"
				}`),
			},

			// address
			4: {
				// _key: bae-c8448e47-6cd1-571f-90bd-364acb80da7b
				// "contact_id": "bae-1fe427b8-ab8d-56c3-9df2-826a6ce86fed"
				(`{
					"city": "Waterloo",
					"country": "Canada"
				}`),

				// _key: bae-f01bf83f-1507-5fb5-a6a3-09ecffa3c692
				// "contact_id": "bae-c0960a29-b704-5c37-9c2e-59e1249e4559"
				(`{
					"city": "Brampton",
					"country": "Canada"
				}`),
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"renderNode": dataMap{
							"countNode": dataMap{
								"filter": dataMap{
									"name": dataMap{
										"$eq": "Theif Lord",
									},
								},
								"sourceProperty": "books",
								"countNode": dataMap{
									"filter":         nil,
									"sourceProperty": "articles",
									"selectNode": dataMap{
										"filter": nil,
										"parallelNode": []dataMap{
											{
												"typeIndexJoin": dataMap{
													"joinType":  "typeJoinOne",
													"direction": "primary",
													"rootName":  "author",
													"root": dataMap{
														"scanNode": dataMap{
															"filter":         nil,
															"collectionID":   "3",
															"collectionName": "author",
															"spans": []dataMap{
																{
																	"start": "/3",
																	"end":   "/4",
																},
															},
														},
													},
													"subTypeName": "contact",
													"subType": dataMap{
														"selectTopNode": dataMap{
															"selectNode": dataMap{
																"filter": nil,
																"typeIndexJoin": dataMap{
																	"joinType":  "typeJoinOne",
																	"direction": "primary",
																	"rootName":  "contact",
																	"root": dataMap{
																		"scanNode": dataMap{
																			"filter":         nil,
																			"collectionID":   "4",
																			"collectionName": "authorContact",
																			"spans": []dataMap{
																				{
																					"start": "/4",
																					"end":   "/5",
																				},
																			},
																		},
																	},
																	"subTypeName": "address",
																	"subType": dataMap{
																		"selectTopNode": dataMap{
																			"selectNode": dataMap{
																				"filter": nil,
																				"scanNode": dataMap{
																					"filter":         nil,
																					"collectionID":   "5",
																					"collectionName": "contactAddress",
																					"spans": []dataMap{
																						{
																							"start": "/5",
																							"end":   "/6",
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
											{
												"typeIndexJoin": dataMap{
													"joinType": "typeJoinMany",
													"rootName": "author",
													"root": dataMap{
														"scanNode": dataMap{
															"filter":         nil,
															"collectionID":   "3",
															"collectionName": "author",
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
																	"filter":         nil,
																	"collectionID":   "1",
																	"collectionName": "article",
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
											{
												"typeIndexJoin": dataMap{
													"joinType": "typeJoinMany",
													"rootName": "author",
													"root": dataMap{
														"scanNode": dataMap{
															"filter":         nil,
															"collectionID":   "3",
															"collectionName": "author",
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
																	"filter":         nil,
																	"collectionID":   "2",
																	"collectionName": "book",
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
			},
		},
	}

	executeTestCase(t, test)
}
