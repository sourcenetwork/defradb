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

func TestExplainTopLevelSumQuery(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain top-level sum query.",

		Query: `query @explain {
			_sum(
				author: {
					field: age
				}
			)
		}`,

		Docs: map[int][]string{
			//authors
			2: {
				`{
					"name": "John",
					"verified": true,
					"age": 21
				}`,
				`{
					"name": "Bob",
					"verified": true,
					"age": 30
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"topLevelNode": []dataMap{
						{
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"filter": nil,
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
							},
						},
						{
							"sumNode": dataMap{
								"sources": []dataMap{
									{
										"fieldName":      "author",
										"childFieldName": "age",
										"filter":         nil,
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

func TestExplainTopLevelSumQueryWithFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain top-level sum query with filter.",

		Query: `query @explain {
			_sum(
				author: {
					field: age,
					filter: {
						age: {
							_gt: 26
						}
					}
				}
			)
		}`,

		Docs: map[int][]string{
			//authors
			2: {
				`{
					"name": "John",
					"verified": false,
					"age": 21
				}`,
				`{
					"name": "Bob",
					"verified": false,
					"age": 30
				}`,
				`{
					"name": "Alice",
					"verified": true,
					"age": 32
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"topLevelNode": []dataMap{
						{
							"selectTopNode": dataMap{
								"selectNode": dataMap{
									"filter": nil,
									"scanNode": dataMap{
										"collectionID":   "3",
										"collectionName": "author",
										"filter": dataMap{
											"age": dataMap{
												"_gt": int(26),
											},
										},
										"spans": []dataMap{
											{
												"start": "/3",
												"end":   "/4",
											},
										},
									},
								},
							},
						},
						{
							"sumNode": dataMap{
								"sources": []dataMap{
									{
										"fieldName":      "author",
										"childFieldName": "age",
										"filter": dataMap{
											"age": dataMap{
												"_gt": int(26),
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
