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

func TestQueryOneToManyWithCountWithFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many relation query from many side with count with filter",
		Request: `query {
			Author {
				name
				_count(published: {filter: {rating: {_gt: 4.8}}})
			}
		}`,
		Docs: map[int][]string{
			//books
			0: { // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				`{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				}`,
			},
			//authors
			1: {
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
		Results: []map[string]any{
			{
				"name":   "John Grisham",
				"_count": 1,
			},
			{
				"name":   "Cornelia Funke",
				"_count": 0,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToManyWithCountWithFilterAndChildFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many relation query from many side with count with filter",
		Request: `query {
			Author {
				name
				_count(published: {filter: {rating: {_ne: null}}})
				published(filter: {rating: {_ne: null}}){
					name
				}
			}
		}`,
		Docs: map[int][]string{
			//books
			0: { // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				`{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "The Associate",
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				}`,
			},
			//authors
			1: {
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
		Results: []map[string]any{
			{
				"name":   "John Grisham",
				"_count": 2,
				"published": []map[string]any{
					{
						"name": "Painted House",
					},
					{
						"name": "A Time for Mercy",
					},
				},
			},
			{
				"name":   "Cornelia Funke",
				"_count": 1,
				"published": []map[string]any{
					{
						"name": "Theif Lord",
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

// This test asserts that only a single join is used - the _count reuses the rendered join as they
// have matching filters.
func TestQueryOneToManyWithCountWithFilterAndChildFilterSharesJoinField(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many relation query from many side with count with filter",
		Request: `query @explain {
			Author {
				name
				_count(published: {filter: {rating: {_ne: null}}})
				published(filter: {rating: {_ne: null}}){
					name
				}
			}
		}`,
		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
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
	}

	executeTestCase(t, test)
}

// This test asserts that two joins are used - the _count cannot reuse the rendered join as they
// dont have matching filters.
func TestQueryOneToManyWithCountAndChildFilterDoesNotShareJoinField(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many relation query from many side with count",
		Request: `query @explain {
			Author {
				name
				_count(published: {})
				published(filter: {rating: {_ne: null}}){
					name
				}
			}
		}`,
		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"countNode": dataMap{
							"selectNode": dataMap{
								"_keys":  nil,
								"filter": nil,
								"parallelNode": []dataMap{
									{
										"typeIndexJoin": dataMap{
											"joinType": "typeJoinMany",
											"root": dataMap{
												"scanNode": dataMap{
													"collectionID":   "2",
													"collectionName": "Author",
													"filter":         nil,
													"spans": []dataMap{
														{
															"end":   "/3",
															"start": "/2",
														},
													},
												},
											},
											"rootName": "author",
											"subType": dataMap{
												"selectTopNode": dataMap{
													"selectNode": dataMap{
														"_keys":  nil,
														"filter": nil,
														"scanNode": dataMap{
															"collectionID":   "1",
															"collectionName": "Book",
															"filter": dataMap{
																"rating": dataMap{
																	"_ne": nil,
																},
															},
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
											"subTypeName": "published",
										},
									},
									{
										"typeIndexJoin": dataMap{
											"joinType": "typeJoinMany",
											"root": dataMap{
												"scanNode": dataMap{
													"collectionID":   "2",
													"collectionName": "Author",
													"filter":         nil,
													"spans": []dataMap{
														{
															"end":   "/3",
															"start": "/2",
														},
													},
												},
											},
											"rootName": "author",
											"subType": dataMap{
												"selectTopNode": dataMap{
													"selectNode": dataMap{
														"_keys":  nil,
														"filter": nil,
														"scanNode": dataMap{
															"collectionID":   "1",
															"collectionName": "Book",
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
											"subTypeName": "published",
										},
									},
								},
							},
							"sources": []dataMap{
								{
									"fieldName": "published",
									"filter":    nil,
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
