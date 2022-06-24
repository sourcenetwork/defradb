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

func TestExplainQuerySumOfRelatedOneToManyField(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain a simple sum query of a One-to-Many realted sub-type.",
		Query: `query @explain {
			author {
				name
				_key
				TotalPages: _sum(books: {field: pages})
			}
		}`,

		Docs: map[int][]string{
			// articles
			0: {
				(`{
					"name": "After Guantánamo, Another Injustice",
					"author_id": "bae-25fafcc7-f251-58c1-9495-ead73e676fb8"
				}`),
				(`{
					"name": "To my dear readers",
					"author_id": "bae-3dddb519-3612-5e43-86e5-49d6295d4f84"
					}`),
				(`{
					"name": "Twinklestar's Favourite Xmas Cookie",
					"author_id": "bae-3dddb519-3612-5e43-86e5-49d6295d4f84"
				}`),
			},
			// books
			1: {
				(`{
					"name": "Painted House",
					"author_id": "bae-25fafcc7-f251-58c1-9495-ead73e676fb8",
					"pages": 22
				}`),
				(`{
					"name": "A Time for Mercy",
					"author_id": "bae-25fafcc7-f251-58c1-9495-ead73e676fb8",
					"pages": 101
					}`),
				(`{
					"name": "Theif Lord",
					"author_id": "bae-3dddb519-3612-5e43-86e5-49d6295d4f84",
					"pages": 321
				}`),
			},
			2: {
				// _key: "bae-25fafcc7-f251-58c1-9495-ead73e676fb8"
				(`{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"contact_id": "bae-1fe427b8-ab8d-56c3-9df2-826a6ce86fed"
				}`),
				// _key: "bae-3dddb519-3612-5e43-86e5-49d6295d4f84"
				(`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false,
					"contact_id": "bae-c0960a29-b704-5c37-9c2e-59e1249e4559"
				}`),
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"renderNode": dataMap{
							"sumNode": dataMap{
								"sourceCollection": "books",
								"sourceProperty":   "pages",
								"filter":           nil,
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
				},
			},
		},
	}

	executeTestCase(t, test)
}
