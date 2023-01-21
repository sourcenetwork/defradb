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
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
)

func TestExplainSimpleNew(t *testing.T) {
	test := explainUtils.ExplainRequestTestCase{

		Description: "NEW Explain simple update mutation with boolean equals filter, multiple rows",

		Request: `mutation @explain {
			update_author(
				filter: {
					verified: {
						_eq: true
					}
				},
				data: "{\"age\": 59}"
			) {
				_key
				name
				age
			}
		}`,

		Docs: map[int][]string{
			2: {
				// bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f
				`{
					"name": "Lone",
					"age":  26,
					"verified": false
				}`,
				// "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
				`{
					"name":     "Shahzad Lone",
					"age":      27,
					"verified": true
				}`,
			},
		},

		// This obviously is optional and shouldn't be provided in most tests, but leaving here for demonstration purposes.
		ExpectedFullGraph: []dataMap{
			{
				"explain": dataMap{
					"updateNode": dataMap{
						"data": dataMap{
							"age": float64(59),
						},
						"filter": dataMap{
							"verified": dataMap{
								"_eq": true,
							},
						},
						"ids": []string(nil),
						"selectTopNode": dataMap{
							"selectNode": dataMap{
								"filter": nil,
								"scanNode": dataMap{
									"collectionID":   "3",
									"collectionName": "author",
									"filter": dataMap{
										"verified": dataMap{
											"_eq": true,
										},
									},
									"spans": []dataMap{
										{
											"end":   "/4",
											"start": "/3",
										},
									},
								},
							},
						},
					},
				},
			},
		},

		ExpectedPatterns: []dataMap{
			{
				"explain": dataMap{
					"updateNode": dataMap{
						"selectTopNode": dataMap{
							"selectNode": dataMap{
								"scanNode": dataMap{},
							},
						},
					},
				},
			},
		},

		ExpectedTargets: []explainUtils.PlanNodeTargetCase{
			{
				TargetNodeName: "updateNode",
				ExpectedAttributes: dataMap{
					"data": dataMap{
						"age": float64(59),
					},
					"filter": dataMap{
						"verified": dataMap{
							"_eq": true,
						},
					},
					"ids": []string(nil),
				},
			},

			{
				TargetNodeName: "selectNode",
				ExpectedAttributes: dataMap{
					"filter": nil,
				},
			},

			{
				TargetNodeName:    "selectNode",
				IncludeChildNodes: true,
				ExpectedAttributes: dataMap{
					"filter": nil,
					"scanNode": dataMap{
						"collectionID":   "3",
						"collectionName": "author",
						"filter": dataMap{
							"verified": dataMap{
								"_eq": true,
							},
						},
						"spans": []dataMap{
							{
								"end":   "/4",
								"start": "/3",
							},
						},
					},
				},
			},
		},
	}

	executeExplainTestCase(t, test)
}

func TestExplainSimpleOld(t *testing.T) {
	test := testUtils.QueryTestCase{

		Description: "Explain simple update mutation with boolean equals filter, multiple rows",

		Query: `mutation @explain {
			update_author(
				filter: {
					verified: {
						_eq: true
					}
				},
				data: "{\"age\": 59}"
			) {
				_key
				name
				age
			}
		}`,

		Docs: map[int][]string{
			2: {
				// bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f
				`{
					"name": "Lone",
					"age":  26,
					"verified": false
				}`,
				// "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
				`{
					"name":     "Shahzad Lone",
					"age":      27,
					"verified": true
				}`,
			},
		},

		Results: []dataMap{
			{
				"explain": dataMap{
					"updateNode": dataMap{
						"data": dataMap{
							"age": float64(59),
						},
						"filter": dataMap{
							"verified": dataMap{
								"_eq": true,
							},
						},
						"ids": []string(nil),
						"selectTopNode": dataMap{
							"selectNode": dataMap{
								"filter": nil,
								"scanNode": dataMap{
									"collectionID":   "3",
									"collectionName": "author",
									"filter": dataMap{
										"verified": dataMap{
											"_eq": true,
										},
									},
									"spans": []dataMap{
										{
											"end":   "/4",
											"start": "/3",
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
