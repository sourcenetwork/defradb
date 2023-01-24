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

func TestExplainQuerySimpleWithDocKeyFilter(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Explain query with basic filter (key by DocKey arg)",

			Query: `query @explain {
				author(dockey: "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d") {
					name
					age
				}
			}`,

			Docs: map[int][]string{
				2: {
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
						"selectTopNode": dataMap{
							"selectNode": dataMap{
								"filter": nil,
								"scanNode": dataMap{
									"collectionID":   "3",
									"collectionName": "author",
									"filter":         nil,
									"spans": []dataMap{
										{
											"start": "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
											"end":   "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9e",
										},
									},
								},
							},
						},
					},
				},
			},
		},

		{
			Description: "Explain query with basic filter (key by DocKey arg), partial results",

			Query: `query @explain {
				author(dockey: "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d") {
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
						"selectTopNode": dataMap{
							"selectNode": dataMap{
								"filter": nil,
								"scanNode": dataMap{
									"collectionID":   "3",
									"collectionName": "author",
									"filter":         nil,
									"spans": []dataMap{
										{
											"start": "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
											"end":   "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9e",
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

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestExplainQuerySimpleWithDocKeysFilter(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Explain query with basic filter (single key by DocKeys arg)",

			Query: `query @explain {
				author(dockeys: ["bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"]) {
					name
					age
				}
			}`,

			Docs: map[int][]string{
				2: {
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
						"selectTopNode": dataMap{
							"selectNode": dataMap{
								"filter": nil,
								"scanNode": dataMap{
									"collectionID":   "3",
									"collectionName": "author",
									"filter":         nil,
									"spans": []dataMap{
										{
											"start": "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
											"end":   "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9e",
										},
									},
								},
							},
						},
					},
				},
			},
		},

		{
			Description: "Explain query with basic filter (duplicate key by DocKeys arg), partial results",

			Query: `query @explain {
				author(dockeys: [
					"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
					"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"
				]) {
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
						"selectTopNode": dataMap{
							"selectNode": dataMap{
								"filter": nil,
								"scanNode": dataMap{
									"collectionID":   "3",
									"collectionName": "author",
									"filter":         nil,
									"spans": []dataMap{
										{
											"start": "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
											"end":   "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9e",
										},
										{
											"start": "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
											"end":   "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9e",
										},
									},
								},
							},
						},
					},
				},
			},
		},

		{
			Description: "Explain query with basic filter (multiple key by DocKeys arg), partial results",

			Query: `query @explain {
				author(dockeys: [
					"bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
					"bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f"
				]) {
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
						"selectTopNode": dataMap{
							"selectNode": dataMap{
								"filter": nil,
								"scanNode": dataMap{
									"collectionID":   "3",
									"collectionName": "author",
									"filter":         nil,
									"spans": []dataMap{
										{
											"start": "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
											"end":   "/3/bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9e",
										},
										{
											"start": "/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67f",
											"end":   "/3/bae-bfbfc89c-0d63-5ea4-81a3-3ebd295be67g",
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

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestExplainSimpleFilterWithMatchingKey(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain with basic filter with matching key.",

		Query: `query @explain {
			author(filter: {_key: {_eq: "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d"}}) {
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
					"selectTopNode": dataMap{
						"selectNode": dataMap{
							"filter": nil,
							"scanNode": dataMap{
								"collectionID":   "3",
								"collectionName": "author",
								"filter": dataMap{
									"_key": dataMap{
										"_eq": "bae-079d0bd8-4b1b-5f5f-bd95-4d915c277f9d",
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
			},
		},
	}

	executeTestCase(t, test)
}
