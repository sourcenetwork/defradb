// Copyright 2022 Democratized Data Foundation
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
)

func TestExplainQueryOneToManyWithACount(t *testing.T) {
	test := testUtils.RequestTestCase{

		Description: "Explain one one-to-many relation query with count.",

		Request: `query @explain {
			Author {
				name
				numberOfBooks: _count(books: {})
			}
		}`,

		Docs: map[int][]string{
			//articles
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
			//books
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
			//authors
			2: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},

		// ----> selectTopNode                (explainable but no-attributes)
		//    ----> countNode                 (explainable)
		//        ----> selectNode            (explainable)
		//             ----> typeIndexJoin    (explainable)
		//                 ----> typeJoinMany (non-explainable)
		//                     ----> scanNode (explainable)
		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"countNode": dataMap{
							"sources": []dataMap{
								{
									"filter":    nil,
									"fieldName": "books",
								},
							},
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
	}

	executeTestCase(t, test)
}
