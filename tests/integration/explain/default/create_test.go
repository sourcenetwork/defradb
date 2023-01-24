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

func TestExplainMutationCreateSimple(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain simple create mutation.",

		Query: `mutation @explain {
			create_author(data: "{\"name\": \"Shahzad Lone\",\"age\": 27,\"verified\": true}") {
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
							"name":     "Shahzad Lone",
							"verified": true,
						},
						"selectTopNode": dataMap{
							"selectNode": dataMap{
								"filter": nil,
								"scanNode": dataMap{
									"collectionID":   "3",
									"collectionName": "author",
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

	executeTestCase(t, test)
}

func TestExplainMutationCreateSimpleDoesNotCreateDocGivenDuplicate(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain simple create mutation, where document already exists.",

		Query: `mutation @explain {
			create_author(data: "{\"name\": \"Shahzad Lone\",\"age\": 27}") {
				_key
				name
				age
			}
		}`,

		Docs: map[int][]string{
			2: {
				`{
					"name": "Shahzad Lone",
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
							"name": "Shahzad Lone",
						},
						"selectTopNode": dataMap{
							"selectNode": dataMap{
								"filter": nil,
								"scanNode": dataMap{
									"collectionID":   "3",
									"collectionName": "author",
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

	executeTestCase(t, test)
}
