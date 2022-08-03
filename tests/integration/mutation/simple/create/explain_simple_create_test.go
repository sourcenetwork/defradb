// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package create

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	simpleTests "github.com/sourcenetwork/defradb/tests/integration/mutation/simple"
)

type dataMap = map[string]interface{}

func TestExplainMutationCreateSimple(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain simple create mutation.",

		Query: `mutation @explain {
					create_user(data: "{\"name\": \"John\",\"age\": 27,\"points\": 42.1,\"verified\": true}") {
						_key
						name
						age
					}
				}`,

		Results: []dataMap{
			{
				"explain": dataMap{
					"createNode": dataMap{
						"data": dataMap{
							"age":      float64(27),
							"name":     "John",
							"points":   42.1,
							"verified": true,
						},
						"selectTopNode": dataMap{
							"selectNode": dataMap{
								"filter": nil,
								"scanNode": dataMap{
									"collectionID":   "1",
									"collectionName": "user",
									"filter":         nil,
									"spans":          []dataMap{},
								},
							},
						},
					},
				},
			},
		},

		ExpectedError: "",
	}

	simpleTests.ExecuteTestCase(t, test)
}

func TestExplainMutationCreateSimpleDoesNotCreateDocGivenDuplicate(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain simple create mutation, where document already exists.",

		Query: `mutation @explain {
					create_user(data: "{\"name\": \"John\",\"age\": 27}") {
						_key
						name
						age
					}
				}`,

		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"age": 27
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"createNode": dataMap{
						"data": dataMap{
							"age":  float64(27),
							"name": "John",
						},
						"selectTopNode": dataMap{
							"selectNode": dataMap{
								"filter": nil,
								"scanNode": dataMap{
									"collectionID":   "1",
									"collectionName": "user",
									"filter":         nil,
									"spans":          []dataMap{},
								},
							},
						},
					},
				},
			},
		},

		ExpectedError: "",
	}

	simpleTests.ExecuteTestCase(t, test)
}
