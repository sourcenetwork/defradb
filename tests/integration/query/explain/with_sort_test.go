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

func TestExplainQuerySimpleSort(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain a simple sort query.",
		Query: `query @explain {
			author(order: {age: ASC}) {
				name
				age
				verified
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
							"sortNode": dataMap{
								"selectNode": dataMap{
									"filter": nil,
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
								"orderings": []dataMap{
									{
										"direction": "ASC",
										"field":     "age",
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

func TestExplainQuerySortAscendingOnParentAndDescendingOnChild(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain Ascending Sort Order on Parent and Descending Sort Order on Child.",
		Query: `query @explain {
			author(order: {name: ASC, age: ASC}) {
				name
				age
				verified
				articles(order: {name: DESC}) {
					name
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
							"sortNode": dataMap{
								"orderings": []dataMap{
									{
										"direction": "ASC",
										"field":     "name",
									},
									{
										"direction": "ASC",
										"field":     "age",
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
										"subTypeName": "articles",
										"subType": dataMap{
											"selectTopNode": dataMap{
												"sortNode": dataMap{
													"orderings": []dataMap{
														{
															"direction": "DESC",
															"field":     "name",
														},
													},
													"selectNode": dataMap{
														"filter": nil,
														"scanNode": dataMap{
															"collectionID":   "1",
															"collectionName": "article",
															"filter":         nil,
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
	}

	executeTestCase(t, test)
}
