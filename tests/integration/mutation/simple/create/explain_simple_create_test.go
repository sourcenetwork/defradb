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
	test := testUtils.RequestTestCase{
		Description: "Explain simple create mutation.",

		Request: `mutation @explain {
					create_user(data: "{\"name\": \"John\",\"age\": 27,\"points\": 42.1,\"verified\": true}") {
						_key
						name
						age
					}
				}`,

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"createNode": dataMap{
								"data": dataMap{
									"age":      float64(27),
									"name":     "John",
									"points":   float64(42.1),
									"verified": true,
								},
							},
							"filter": nil,
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
	test := testUtils.RequestTestCase{
		Description: "Explain simple create mutation, where document already exists.",

		Request: `mutation @explain {
					create_user(data: "{\"name\": \"John\",\"age\": 27}") {
						_key
						name
						age
					}
				}`,

		Docs: map[int][]string{
			0: {
				(`{
				"name": "John",
				"age": 27
			}`)},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"createNode": dataMap{
								"data": dataMap{
									"age":  float64(27),
									"name": "John",
								},
							},
							"filter": nil,
						},
					},
				},
			},
		},

		ExpectedError: "",
	}

	simpleTests.ExecuteTestCase(t, test)
}
