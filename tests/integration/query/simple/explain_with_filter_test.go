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

func TestExplainQuerySimpleWithDocKeyFilter(t *testing.T) {
	tests := []testUtils.QueryTestCase{
		{
			Description: "Explain query with basic filter (key by DocKey arg)",
			Query: `query @explain {
						users(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
							Name
							Age
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},

			Results: []map[string]interface{}{
				{
					"explain": map[string]interface{}{
						"selectTopNode": map[string]interface{}{
							"selectNode": map[string]interface{}{
								"filter": nil,
								"scanNode": map[string]interface{}{
									"collectionID":   "1",
									"collectionName": "users",
									"filter":         nil,
								},
							},
						},
					},
				},
			},
		},

		{
			Description: "Explain query with basic filter (key by DocKey arg), no results",
			Query: `query @explain {
								users(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009g") {
									Name
									Age
								}
							}`,
			Docs: map[int][]string{
				0: {
					(`{
							"Name": "John",
							"Age": 21
						}`)},
			},
			Results: []map[string]interface{}{
				{
					"explain": map[string]interface{}{
						"selectTopNode": map[string]interface{}{
							"selectNode": map[string]interface{}{
								"filter": nil,
								"scanNode": map[string]interface{}{
									"collectionID":   "1",
									"collectionName": "users",
									"filter":         nil,
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
								users(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
									Name
									Age
								}
							}`,
			Docs: map[int][]string{
				0: {
					(`{
							"Name": "John",
							"Age": 21
							}`),
					(`{
								"Name": "Bob",
								"Age": 32
							}`)},
			},
			Results: []map[string]interface{}{
				{
					"explain": map[string]interface{}{
						"selectTopNode": map[string]interface{}{
							"selectNode": map[string]interface{}{
								"filter": nil,
								"scanNode": map[string]interface{}{
									"collectionID":   "1",
									"collectionName": "users",
									"filter":         nil,
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
	tests := []testUtils.QueryTestCase{
		{
			Description: "Explain query with basic filter (single key by DocKeys arg)",
			Query: `query @explain {
						users(dockeys: ["bae-52b9170d-b77a-5887-b877-cbdbb99b009f"]) {
							Name
							Age
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			Results: []map[string]interface{}{
				{
					"explain": map[string]interface{}{
						"selectTopNode": map[string]interface{}{
							"selectNode": map[string]interface{}{
								"filter": nil,
								"scanNode": map[string]interface{}{
									"collectionID":   "1",
									"collectionName": "users",
									"filter":         nil,
								},
							},
						},
					},
				},
			},
		},

		{
			Description: "Explain query with basic filter (single key by DocKeys arg), no results",
			Query: `query @explain {
						users(dockeys: ["bae-52b9170d-b77a-5887-b877-cbdbb99b009g"]) {
							Name
							Age
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			Results: []map[string]interface{}{
				{
					"explain": map[string]interface{}{
						"selectTopNode": map[string]interface{}{
							"selectNode": map[string]interface{}{
								"filter": nil,
								"scanNode": map[string]interface{}{
									"collectionID":   "1",
									"collectionName": "users",
									"filter":         nil,
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
								users(dockeys: ["bae-52b9170d-b77a-5887-b877-cbdbb99b009f", "bae-52b9170d-b77a-5887-b877-cbdbb99b009f"]) {
									Name
									Age
								}
							}`,
			Docs: map[int][]string{
				0: {
					(`{
							"Name": "John",
							"Age": 21
							}`),
					(`{
								"Name": "Bob",
								"Age": 32
							}`)},
			},
			Results: []map[string]interface{}{
				{
					"explain": map[string]interface{}{
						"selectTopNode": map[string]interface{}{
							"selectNode": map[string]interface{}{
								"filter": nil,
								"scanNode": map[string]interface{}{
									"collectionID":   "1",
									"collectionName": "users",
									"filter":         nil,
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
								users(dockeys: ["bae-52b9170d-b77a-5887-b877-cbdbb99b009f", "bae-1378ab62-e064-5af4-9ea6-49941c8d8f94"]) {
									Name
									Age
								}
							}`,
			Docs: map[int][]string{
				0: {
					(`{
							"Name": "John",
							"Age": 21
							}`),
					(`{
								"Name": "Bob",
								"Age": 32
							}`),
					(`{
								"Name": "Jim",
								"Age": 27
							}`)},
			},
			Results: []map[string]interface{}{
				{
					"explain": map[string]interface{}{
						"selectTopNode": map[string]interface{}{
							"selectNode": map[string]interface{}{
								"filter": nil,
								"scanNode": map[string]interface{}{
									"collectionID":   "1",
									"collectionName": "users",
									"filter":         nil,
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

func TestExplainQuerySimpleWithKeyFilterBlock(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain query with basic filter (key by filter block)",
		Query: `query @explain {
					users(filter: {_key: {_eq: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f"}}) {
						Name
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 21
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"explain": map[string]interface{}{
					"selectTopNode": map[string]interface{}{
						"selectNode": map[string]interface{}{
							"filter": nil,
							"scanNode": map[string]interface{}{
								"collectionID":   "1",
								"collectionName": "users",
								"filter": map[string]interface{}{
									"_key": map[string]interface{}{
										"$eq": "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
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

func TestExplainQuerySimpleWithStringFilterBlock(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain query with basic filter (Name)",
		Query: `query @explain {
					users(filter: {Name: {_eq: "John"}}) {
						Name
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 21
			}`)},
		},

		Results: []map[string]interface{}{
			{
				"explain": map[string]interface{}{
					"selectTopNode": map[string]interface{}{
						"selectNode": map[string]interface{}{
							"filter": nil,
							"scanNode": map[string]interface{}{
								"collectionID":   "1",
								"collectionName": "users",
								"filter": map[string]interface{}{
									"Name": map[string]interface{}{
										"$eq": "John",
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
	tests := []testUtils.QueryTestCase{
		{
			Description: "Explain query with basic filter and selection",
			Query: `query @explain {
						users(filter: {Name: {_eq: "John"}}) {
							Name
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			Results: []map[string]interface{}{
				{
					"explain": map[string]interface{}{
						"selectTopNode": map[string]interface{}{
							"selectNode": map[string]interface{}{
								"filter": nil,
								"scanNode": map[string]interface{}{
									"collectionID":   "1",
									"collectionName": "users",
									"filter": map[string]interface{}{
										"Name": map[string]interface{}{
											"$eq": "John",
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
			Description: "Explain query with basic filter and selection (diff from filter)",
			Query: `query @explain {
								users(filter: {Name: {_eq: "John"}}) {
									Age
								}
							}`,
			Docs: map[int][]string{
				0: {
					(`{
							"Name": "John",
							"Age": 21
						}`)},
			},
			Results: []map[string]interface{}{
				{
					"explain": map[string]interface{}{
						"selectTopNode": map[string]interface{}{
							"selectNode": map[string]interface{}{
								"filter": nil,
								"scanNode": map[string]interface{}{
									"collectionID":   "1",
									"collectionName": "users",
									"filter": map[string]interface{}{
										"Name": map[string]interface{}{
											"$eq": "John",
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
			Description: "Explain query with basic filter(name), no results",
			Query: `query @explain {
								users(filter: {Name: {_eq: "Bob"}}) {
									Name
									Age
								}
							}`,
			Docs: map[int][]string{
				0: {
					(`{
							"Name": "John",
							"Age": 21
						}`)},
			},
			Results: []map[string]interface{}{
				{
					"explain": map[string]interface{}{
						"selectTopNode": map[string]interface{}{
							"selectNode": map[string]interface{}{
								"filter": nil,
								"scanNode": map[string]interface{}{
									"collectionID":   "1",
									"collectionName": "users",
									"filter": map[string]interface{}{
										"Name": map[string]interface{}{
											"$eq": "Bob",
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
	test := testUtils.QueryTestCase{
		Description: "Explain query with basic filter(age)",
		Query: `query @explain {
					users(filter: {Age: {_eq: 21}}) {
						Name
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 21
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"explain": map[string]interface{}{
					"selectTopNode": map[string]interface{}{
						"selectNode": map[string]interface{}{
							"filter": nil,
							"scanNode": map[string]interface{}{
								"collectionID":   "1",
								"collectionName": "users",
								"filter": map[string]interface{}{
									"Age": map[string]interface{}{
										"$eq": int64(21), // Should this be `uint64`?
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
	tests := []testUtils.QueryTestCase{
		{
			Description: "Explain query with basic filter(age), greater than",
			Query: `query @explain {
						users(filter: {Age: {_gt: 20}}) {
							Name
							Age
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"Age": 21
				}`)},
			},
			Results: []map[string]interface{}{
				{
					"explain": map[string]interface{}{
						"selectTopNode": map[string]interface{}{
							"selectNode": map[string]interface{}{
								"filter": nil,
								"scanNode": map[string]interface{}{
									"collectionID":   "1",
									"collectionName": "users",
									"filter": map[string]interface{}{
										"Age": map[string]interface{}{
											"$gt": int64(20), // Should this be `uint64`?
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
			Description: "Explain query with basic filter(age), no results",
			Query: `query @explain {
								users(filter: {Age: {_gt: 40}}) {
									Name
									Age
								}
							}`,
			Docs: map[int][]string{
				0: {
					(`{
							"Name": "John",
							"Age": 21
						}`),
					(`{
							"Name": "Bob",
							"Age": 32
						}`)},
			},
			Results: []map[string]interface{}{
				{
					"explain": map[string]interface{}{
						"selectTopNode": map[string]interface{}{
							"selectNode": map[string]interface{}{
								"filter": nil,
								"scanNode": map[string]interface{}{
									"collectionID":   "1",
									"collectionName": "users",
									"filter": map[string]interface{}{
										"Age": map[string]interface{}{
											"$gt": int64(40), // Should this be `uint64`?
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
			Query: `query @explain {
								users(filter: {Age: {_gt: 20}}) {
									Name
									Alias: Age
									_key
								}
							}`,
			Docs: map[int][]string{
				0: {
					(`{
							"Name": "John",
							"Age": 21
						}`),
					(`{
							"Name": "Bob",
							"Age": 32
						}`)},
			},
			Results: []map[string]interface{}{
				{
					"explain": map[string]interface{}{
						"selectTopNode": map[string]interface{}{
							"selectNode": map[string]interface{}{
								"filter": nil,
								"scanNode": map[string]interface{}{
									"collectionID":   "1",
									"collectionName": "users",
									"filter": map[string]interface{}{
										"Age": map[string]interface{}{
											"$gt": int64(20), // Should this be `uint64`?
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
	test := testUtils.QueryTestCase{
		Description: "Explain query with logical compound filter (and)",
		Query: `query @explain {
					users(filter: {_and: [{Age: {_gt: 20}}, {Age: {_lt: 50}}]}) {
						Name
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 21
			}`),
				(`{
				"Name": "Bob",
				"Age": 32
			}`),
				(`{
				"Name": "Carlo",
				"Age": 55
			}`),
				(`{
				"Name": "Alice",
				"Age": 19
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"explain": map[string]interface{}{
					"selectTopNode": map[string]interface{}{
						"selectNode": map[string]interface{}{
							"filter": nil,
							"scanNode": map[string]interface{}{
								"collectionID":   "1",
								"collectionName": "users",
								"filter": map[string]interface{}{
									"$and": []interface{}{
										map[string]interface{}{
											"Age": map[string]interface{}{
												"$gt": int64(20), // Should this be `uint64`?
											},
										},
										map[string]interface{}{
											"Age": map[string]interface{}{
												"$lt": int64(50), // Should this be `uint64`?
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

func TestExplainQuerySimpleWithNumberEqualToXOrYFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain query with logical compound filter (or)",
		Query: `query @explain {
					users(filter: {_or: [{Age: {_eq: 55}}, {Age: {_eq: 19}}]}) {
						Name
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 21
			}`),
				(`{
				"Name": "Bob",
				"Age": 32
			}`),
				(`{
				"Name": "Carlo",
				"Age": 55
			}`),
				(`{
				"Name": "Alice",
				"Age": 19
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"explain": map[string]interface{}{
					"selectTopNode": map[string]interface{}{
						"selectNode": map[string]interface{}{
							"filter": nil,
							"scanNode": map[string]interface{}{
								"collectionID":   "1",
								"collectionName": "users",
								"filter": map[string]interface{}{
									"$or": []interface{}{
										map[string]interface{}{
											"Age": map[string]interface{}{
												"$eq": int64(55), // Should this be `uint64`?
											},
										},
										map[string]interface{}{
											"Age": map[string]interface{}{
												"$eq": int64(19), // Should this be uint64?
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

func TestExplainQuerySimpleWithNumberInFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Explain query with special filter (or)",
		Query: `query @explain {
					users(filter: {Age: {_in: [19, 40, 55]}}) {
						Name
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 21
			}`),
				(`{
				"Name": "Bob",
				"Age": 32
			}`),
				(`{
				"Name": "Carlo",
				"Age": 55
			}`),
				(`{
				"Name": "Alice",
				"Age": 19
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"explain": map[string]interface{}{
					"selectTopNode": map[string]interface{}{
						"selectNode": map[string]interface{}{
							"filter": nil,
							"scanNode": map[string]interface{}{
								"collectionID":   "1",
								"collectionName": "users",
								"filter": map[string]interface{}{
									"Age": map[string]interface{}{
										"$in": []interface{}{
											int64(19), // Should this be uint64?
											int64(40), // Should this be uint64?
											int64(55), // Should this be uint64?
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
