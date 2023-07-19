// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_many

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// This test asserts that only a single join is used - the _avg reuses the rendered join as they
// have matching filters (average adds a ne nil filter).
func TestQueryOneToManyWithAverageAndChildNeNilFilterSharesJoinField(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many relation query from many side with average",
		Request: `query @explain {
			Author {
				name
				_avg(published: {field: rating})
				published(filter: {rating: {_ne: null}}){
					name
				}
			}
		}`,
		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"averageNode": dataMap{
							"countNode": dataMap{
								"sources": []dataMap{
									{
										"filter": dataMap{
											"rating": dataMap{
												"_ne": nil,
											},
										},
										"fieldName": "published",
									},
								},
								"sumNode": dataMap{
									"sources": []dataMap{
										{
											"filter": dataMap{
												"rating": dataMap{
													"_ne": nil,
												},
											},
											"fieldName":      "published",
											"childFieldName": "rating",
										},
									},
									"selectNode": dataMap{
										"_keys":  nil,
										"filter": nil,
										"typeIndexJoin": dataMap{
											"joinType": "typeJoinMany",
											"rootName": "author",
											"root": dataMap{
												"scanNode": dataMap{
													"filter":         nil,
													"collectionID":   "2",
													"collectionName": "Author",
													"spans": []dataMap{
														{
															"start": "/2",
															"end":   "/3",
														},
													},
												},
											},
											"subTypeName": "published",
											"subType": dataMap{
												"selectTopNode": dataMap{
													"selectNode": dataMap{
														"_keys":  nil,
														"filter": nil,
														"scanNode": dataMap{
															"filter": dataMap{
																"rating": dataMap{
																	"_ne": nil,
																},
															},
															"collectionID":   "1",
															"collectionName": "Book",
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
