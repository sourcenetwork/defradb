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

func TestExplainQueryOneToManyWithACount(t *testing.T) {
	test := testUtils.QueryTestCase{

		Description: "Explain one one-to-many relation query with count.",

		Query: `query @explain {
				author {
					name
					numberOfBooks: _count(books: {})
				}
			}`,

		Docs: map[int][]string{
			//articles
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
			//books
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
			//authors
			2: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				(`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`),
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				(`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`),
			},
			//authorContact
			3: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				(`{
					"cell": "5197212301",
					"email": "john_grisham@example.com",
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`),
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				(`{
					"cell": "5197212302",
					"email": "cornelia_funke@example.com",
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				}`),
			},
		},

		// ----> selectTopNode                    (explainable but no-attributes)
		//     ----> renderNode                   (explainable)
		//         ----> countNode                (explainable)
		//             ----> selectNode           (explainable)
		//                 ----> typeIndexJoin    (explainable)
		//                     ----> typeJoinMany (non-explainable)
		//                         ----> scanNode (explainable)
		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"renderNode": dataMap{
							"countNode": dataMap{
								"filter":         nil,
								"sourceProperty": "books",
								"selectNode": dataMap{
									"filter": nil,
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
	}

	executeTestCase(t, test)
}

func TestExplainQueryOneToManyMultipleWithCounts(t *testing.T) {
	test := testUtils.QueryTestCase{

		Description: "Explain two typeJoinMany query with both count.",

		Query: `query @explain {
				author {
					name
					numberOfBooks: _count(books: {})
					numberOfArticles: _count(
						articles: {
							filter: {
								name: {
									_eq: "After Guantánamo, Another Injustice"
								}
							}
						}
					)
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

		// ----> selectTopNode                                 (explainable but no attributes)
		//     ----> renderNode                                (explainable)
		//         ----> countNode                             (explainable)
		//             ----> countNode                         (explainable)
		//                 ----> selectNode                    (explainable)
		//                     ----> parallelNode              (non-explainable but wraps children)
		//                         ----> typeIndexJoin         (explainable)
		//                             ----> typeJoinMany      (non-explainable)
		//                                 ----> multiscanNode (non-explainable)
		//                                     ----> scanNode  (explainable)
		//                         ----> typeIndexJoin         (explainable)
		//                             ----> typeJoinMany      (non-explainable)
		//                                 ----> multiscanNode (non-explainable)
		//                                     ----> scanNode  (explainable)
		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"renderNode": dataMap{
							"countNode": dataMap{
								"filter":         nil,
								"sourceProperty": "books",
								"countNode": dataMap{
									"filter": dataMap{
										"name": dataMap{
											"$eq": "After Guantánamo, Another Injustice",
										},
									},
									"sourceProperty": "articles",
									"selectNode": dataMap{
										"filter": nil,
										"parallelNode": []dataMap{
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
