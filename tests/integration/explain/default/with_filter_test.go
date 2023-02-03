// Copyright 2023 Democratized Data Foundation
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

func TestExplainQuerySimpleWithStringFilterBlock(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain query with basic filter (name)",

		Request: `query @explain {
			author(filter: {name: {_eq: "Lone"}}) {
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
									"name": dataMap{
										"_eq": "Lone",
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

func TestExplainQuerySimpleWithStringFilterBlockAndSelect(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Explain query with basic filter(name), no results",

			Request: `query @explain {
				author(filter: {name: {_eq: "Bob"}}) {
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
									"filter": dataMap{
										"name": dataMap{
											"_eq": "Bob",
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
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestExplainQuerySimpleWithNumberEqualsFilterBlock(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain query with basic filter(age)",

		Request: `query @explain {
			author(filter: {age: {_eq: 26}}) {
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
									"age": dataMap{
										"_eq": int(26),
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

func TestExplainQuerySimpleWithNumberGreaterThanFilterBlock(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "Explain query with basic filter(age), greater than",

			Request: `query @explain {
				author(filter: {age: {_gt: 20}}) {
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
										"age": dataMap{
											"_gt": int(20),
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
		},

		{
			Description: "Explain query with basic filter(age), and aliased, multiple results",

			Request: `query @explain {
				author(filter: {age: {_gt: 20}}) {
					name
					Alias: age
					_key
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
										"age": dataMap{
											"_gt": int(20),
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
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestExplainQuerySimpleWithNumberGreaterThanAndNumberLessThanFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain query with logical compound filter (and)",

		Request: `query @explain {
			author(filter: {_and: [{age: {_gt: 20}}, {age: {_lt: 50}}]}) {
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
					"age": 32
				}`,
				`{
					"name": "Carlo",
					"age": 55
				}`,
				`{
					"name": "Alice",
					"age": 19
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
									"_and": []any{
										dataMap{
											"age": dataMap{
												"_gt": int(20),
											},
										},
										dataMap{
											"age": dataMap{
												"_lt": int(50),
											},
										},
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

func TestExplainQuerySimpleWithNumberEqualToXOrYFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain query with logical compound filter (or)",

		Request: `query @explain {
			author(filter: {_or: [{age: {_eq: 55}}, {age: {_eq: 19}}]}) {
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
					"age": 32
				}`,
				`{
					"name": "Carlo",
					"age": 55
				}`,
				`{
					"name": "Alice",
					"age": 19
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
									"_or": []any{
										dataMap{
											"age": dataMap{
												"_eq": int(55),
											},
										},
										dataMap{
											"age": dataMap{
												"_eq": int(19),
											},
										},
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

func TestExplainQuerySimpleWithNumberInFilter(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Explain query with special filter (or)",

		Request: `query @explain {
			author(filter: {age: {_in: [19, 40, 55]}}) {
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
					"age": 32
				}`,
				`{
					"name": "Carlo",
					"age": 55
				}`,
				`{
					"name": "Alice",
					"age": 19
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
									"age": dataMap{
										"_in": []any{
											int(19),
											int(40),
											int(55),
										},
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
