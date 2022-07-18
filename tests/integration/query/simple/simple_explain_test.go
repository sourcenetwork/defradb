// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestExplainQuerySimpleOnFieldDirective_BadUsage(t *testing.T) {
	test := testUtils.QueryTestCase{

		Description: "Explain a query by providing the directive on wrong location (field).",

		Query: `query {
					users @explain {
						_key
						Name
						Age
					}
				}`,

		Docs: map[int][]string{
			0: {
				`{
				"Name": "John",
				"Age": 21
			}`},
		},

		Results: []dataMap{},

		ExpectedError: "[Directive \"explain\" may not be used on FIELD.]",
	}
	executeTestCase(t, test)
}

func TestExplainQuerySimple(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain a query with no filter",
		Query: `query @explain {
					users {
						_key
						Name
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
				"Name": "John",
				"Age": 21
			}`},
		},
		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"filter": nil,
							"scanNode": dataMap{
								"filter":         nil,
								"collectionID":   "1",
								"collectionName": "users",
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
	}

	executeTestCase(t, test)
}

func TestExplainQuerySimpleWithAlias(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain a query with alias, no filter",
		Query: `query @explain {
					users {
						username: Name
						age: Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
				"Name": "John",
				"Age": 21
			}`},
		},
		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"filter": nil,
							"scanNode": dataMap{
								"filter":         nil,
								"collectionID":   "1",
								"collectionName": "users",
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
	}

	executeTestCase(t, test)
}

func TestExplainQuerySimpleWithMultipleRows(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain a query with no filter, mutiple rows",
		Query: `query @explain {
					users {
						Name
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
				"Name": "John",
				"Age": 21
			}`,
				`{
				"Name": "Bob",
				"Age": 27
			}`},
		},
		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"filter": nil,
							"scanNode": dataMap{
								"filter":         nil,
								"collectionID":   "1",
								"collectionName": "users",
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
	}

	executeTestCase(t, test)
}
