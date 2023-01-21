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

func TestExplainSimpleMutationUpdateWithBooleanFilter(t *testing.T) {
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

func TestExplainSimpleMutationUpdateWithIdInFilter(t *testing.T) {
	test := testUtils.QueryTestCase{

		Description: "Explain simple update mutation with id in filter, multiple rows",

		Query: `mutation @explain {
			update_author(
				ids: [
					"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
					"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
				],
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
						"filter": nil,
						"ids": []string{
							"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
							"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
						},
						"selectTopNode": dataMap{
							"selectNode": dataMap{
								"filter": nil,
								"scanNode": dataMap{
									"collectionID":   "3",
									"collectionName": "author",
									"filter":         nil,
									"spans": []dataMap{
										{
											"end":   "/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67g",
											"start": "/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
										},
										{
											"end":   "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9e",
											"start": "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
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

func TestExplainSimpleMutationUpdateWithIdEqualsFilter(t *testing.T) {
	test := testUtils.QueryTestCase{

		Description: "Explain simple update mutation with id equals filter, multiple rows but single match",

		Query: `mutation @explain {
			update_author(
				id: "bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
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
						"filter": nil,
						"ids": []string{
							"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
						},
						"selectTopNode": dataMap{
							"selectNode": dataMap{
								"filter": nil,
								"scanNode": dataMap{
									"collectionID":   "3",
									"collectionName": "author",
									"filter":         nil,
									"spans": []dataMap{
										{
											"end":   "/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67g",
											"start": "/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
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

func TestExplainSimpleMutationUpdateWithIdAndFilter(t *testing.T) {
	test := testUtils.QueryTestCase{

		Description: "Explain simple update mutation with ids and filter, multiple rows",

		Query: `mutation @explain {
			update_author(
				filter: {
					verified: {
						_eq: true
					}
				},
				ids: [
					"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
					"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
				],
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
						"ids": []string{
							"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
							"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
						},
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
											"end":   "/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67g",
											"start": "/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
										},
										{
											"end":   "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9e",
											"start": "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
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
