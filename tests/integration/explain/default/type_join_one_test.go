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

	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

func TestDefaultExplainRequestWithAOneToOneJoin(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with a 1-to-1 join.",

		Request: `query @explain {
			Author {
				OnlyEmail: contact {
					email
				}
			}
		}`,

		Docs: map[int][]string{
			// articles
			0: {
				`{
					"name": "After Guantánamo, Another Injustice",
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "To my dear readers",
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
					}`,
				`{
					"name": "Twinklestar's Favourite Xmas Cookie",
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				}`,
			},
			// books
			1: {
				`{
					"name": "Painted House",
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "A Time for Mercy",
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
					}`,
				`{
					"name": "Theif Lord",
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				}`,
			},
			// authors
			2: {
				// _key: bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"contact_id": "bae-1fe427b8-ab8d-56c3-9df2-826a6ce86fed"
				}`,
				// _key: bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false,
					"contact_id": "bae-c0960a29-b704-5c37-9c2e-59e1249e4559"
				}`,
			},
			// contact
			3: {
				// _key: bae-1fe427b8-ab8d-56c3-9df2-826a6ce86fed
				// "author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				`{
					"cell": "5197212301",
					"email": "john_grisham@example.com",
					"address_id": "bae-c8448e47-6cd1-571f-90bd-364acb80da7b"
				}`,

				// _key: bae-c0960a29-b704-5c37-9c2e-59e1249e4559
				// "author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				`{
					"cell": "5197212302",
					"email": "cornelia_funke@example.com",
					"address_id": "bae-c0960a29-b704-5c37-9c2e-59e1249e4559"
				}`,
			},

			// address
			4: {
				// _key: bae-c8448e47-6cd1-571f-90bd-364acb80da7b
				// "contact_id": "bae-1fe427b8-ab8d-56c3-9df2-826a6ce86fed"
				`{
					"city": "Waterloo",
					"country": "Canada"
				}`,

				// _key: bae-f01bf83f-1507-5fb5-a6a3-09ecffa3c692
				// "contact_id": "bae-c0960a29-b704-5c37-9c2e-59e1249e4559"
				`{
					"city": "Brampton",
					"country": "Canada"
				}`,
			},
		},

		ExpectedPatterns: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"typeIndexJoin": normalTypeJoinPattern,
						},
					},
				},
			},
		},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "typeIndexJoin",
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"direction":   "primary",
					"joinType":    "typeJoinOne",
					"rootName":    "author",
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
						"collectionID":   "3",
						"collectionName": "Author",
						"spans": []dataMap{
							{
								"start": "/3",
								"end":   "/4",
							},
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
							"filter": nil,
							"scanNode": dataMap{
								"filter":         nil,
								"collectionID":   "4",
								"collectionName": "AuthorContact",
								"spans": []dataMap{
									{
										"start": "/4",
										"end":   "/5",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	runExplainTest(t, test)
}

func TestDefaultExplainRequestWithTwoLevelDeepNestedJoins(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "Explain (default) request with two level deep nested joins.",

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

		Docs: map[int][]string{
			// articles
			0: {
				`{
					"name": "After Guantánamo, Another Injustice",
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "To my dear readers",
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
					}`,
				`{
					"name": "Twinklestar's Favourite Xmas Cookie",
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				}`,
			},
			// books
			1: {
				`{
					"name": "Painted House",
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "A Time for Mercy",
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
					}`,
				`{
					"name": "Theif Lord",
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				}`,
			},
			// authors
			2: {
				// _key: bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"contact_id": "bae-1fe427b8-ab8d-56c3-9df2-826a6ce86fed"
				}`,
				// _key: bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false,
					"contact_id": "bae-c0960a29-b704-5c37-9c2e-59e1249e4559"
				}`,
			},
			// contact
			3: {
				// _key: bae-1fe427b8-ab8d-56c3-9df2-826a6ce86fed
				// "author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				`{
					"cell": "5197212301",
					"email": "john_grisham@example.com",
					"address_id": "bae-c8448e47-6cd1-571f-90bd-364acb80da7b"
				}`,

				// _key: bae-c0960a29-b704-5c37-9c2e-59e1249e4559
				// "author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				`{
					"cell": "5197212302",
					"email": "cornelia_funke@example.com",
					"address_id": "bae-c0960a29-b704-5c37-9c2e-59e1249e4559"
				}`,
			},

			// address
			4: {
				// _key: bae-c8448e47-6cd1-571f-90bd-364acb80da7b
				// "contact_id": "bae-1fe427b8-ab8d-56c3-9df2-826a6ce86fed"
				`{
					"city": "Waterloo",
					"country": "Canada"
				}`,

				// _key: bae-f01bf83f-1507-5fb5-a6a3-09ecffa3c692
				// "contact_id": "bae-c0960a29-b704-5c37-9c2e-59e1249e4559"
				`{
					"city": "Brampton",
					"country": "Canada"
				}`,
			},
		},

		ExpectedPatterns: []dataMap{
			{
				"explain": dataMap{
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

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName:    "typeIndexJoin",
				OccurancesToSkip:  0,
				IncludeChildNodes: false,
				ExpectedAttributes: dataMap{
					"direction":   "primary",
					"joinType":    "typeJoinOne",
					"rootName":    "author",
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
						"collectionID":   "3",
						"collectionName": "Author",
						"spans": []dataMap{
							{
								"start": "/3",
								"end":   "/4",
							},
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
					"rootName":    "contact",
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
						"collectionID":   "4",
						"collectionName": "AuthorContact",
						"spans": []dataMap{
							{
								"start": "/4",
								"end":   "/5",
							},
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
							"filter": nil,
							"scanNode": dataMap{
								"filter":         nil,
								"collectionID":   "5",
								"collectionName": "ContactAddress",
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
	}

	runExplainTest(t, test)
}
