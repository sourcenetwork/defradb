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

func TestExplainQuerySimpleOnFieldDirective_BadUsage(t *testing.T) {
	test := testUtils.RequestTestCase{

		Description: "Explain a query by providing the directive on wrong location (field).",

		Request: `query {
			author @explain {
				_key
				name
				age
			}
		}`,

		Docs: map[int][]string{
			2: {
				`{
					"name": "John",
					"age": 21
				}`,
			},
		},

		ExpectedError: "Directive \"explain\" may not be used on FIELD.",
	}
	executeTestCase(t, test)
}

func TestExplainQuerySimple(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain a query with no filter",

		Request: `query @explain {
			author {
				_key
				name
				age
			}
		}`,

		Docs: map[int][]string{
			2: {
				`{
					"name": "John",
					"age": 21
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"filter": nil,
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
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestExplainQuerySimpleWithAlias(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain a query with alias, no filter",

		Request: `query @explain {
			author {
				username: name
				age: age
			}
		}`,

		Docs: map[int][]string{
			2: {
				`{
					"name": "John",
					"age": 21
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"filter": nil,
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
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestExplainQuerySimpleWithMultipleRows(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain a query with no filter, mutiple rows",

		Request: `query @explain {
			author {
				name
				age
			}
		}`,

		Docs: map[int][]string{
			2: {
				`{
					"name": "John",
					"age": 21
				}`,
				`{
					"name": "Bob",
					"age": 27
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"filter": nil,
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
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
