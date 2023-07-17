// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package index

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestIndexWithExplain(t *testing.T) {
	test := testUtils.TestCase{
		Description: "",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String 
					age: Int
					verified: Boolean
				} 
			`),
			testUtils.Request{
				Request: `
					query @explain(type: execute) {
						users(filter: {name: {_eq: "Islam"}}) {
							name
						}
					}`,
				Asserter: newExplainAsserter(2, 2, 1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryWithIndex_WithOnlyIndexedField_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If there is only one indexed field in the query, it should be fetched",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String @index
				} 
			`),
			testUtils.Request{
				Request: `
					query {
						users(filter: {name: {_eq: "Islam"}}) {
							name
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Islam",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryWithIndex_WithNonIndexedFields_ShouldFetchAllOfThem(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If there are non-indexed fields in the query, they should be fetched",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String @index
					age: Int
				} 
			`),
			testUtils.Request{
				Request: `
					query {
						users(filter: {name: {_eq: "Islam"}}) {
							name
							age
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Islam",
						"age":  uint64(32),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryWithIndex_IfMoreThenOneDoc_ShouldFetchAll(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If there are more than one doc with the same indexed field, they should be fetched",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String @index
					age: Int
				} 
			`),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Islam",
					"age": 18
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						users(filter: {name: {_eq: "Islam"}}) {
							age
						}
					}`,
				Results: []map[string]any{
					{
						"age": uint64(32),
					},
					{
						"age": uint64(18),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryWithIndex_WithGreaterThanFilter_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _gt filter",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String 
					age: Int @index
				} 
			`),
			testUtils.Request{
				Request: `
					query {
						users(filter: {age: {_gt: 48}}) {
							name
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Chris",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryWithIndex_WithGreaterOrEqualFilter_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _ge filter",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String 
					age: Int @index
				} 
			`),
			testUtils.Request{
				Request: `
					query {
						users(filter: {age: {_ge: 48}}) {
							name
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Keenan",
					},
					{
						"name": "Chris",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryWithIndex_WithLessThanFilter_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _lt filter",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String 
					age: Int @index
				} 
			`),
			testUtils.Request{
				Request: `
					query {
						users(filter: {age: {_lt: 28}}) {
							name
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Shahzad",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryWithIndex_WithLessOrEqualFilter_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _le filter",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String 
					age: Int @index
				} 
			`),
			testUtils.Request{
				Request: `
					query {
						users(filter: {age: {_le: 28}}) {
							name
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Shahzad",
					},
					{
						"name": "Fred",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryWithIndex_WithNotEqualFilter_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _ne filter",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String @index
					age: Int 
				} 
			`),
			testUtils.Request{
				Request: `
					query {
						users(filter: {name: {_ne: "Islam"}}) {
							name
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Addo",
					},
					{
						"name": "Andy",
					},
					{
						"name": "Fred",
					},
					{
						"name": "John",
					},
					{
						"name": "Chris",
					},
					{
						"name": "Keenan",
					},
					{
						"name": "Shahzad",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryWithIndex_WithInFilter_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _in filter",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String 
					age: Int @index
				} 
			`),
			testUtils.Request{
				Request: `
					query {
						users(filter: {age: {_in: [20, 33]}}) {
							name
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Shahzad",
					},
					{
						"name": "Andy",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryWithIndex_WithNotInFilter_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _nin filter",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String 
					age: Int @index
				} 
			`),
			testUtils.Request{
				Request: `
					query {
						users(filter: {age: {_nin: [20, 28, 33, 42, 55]}}) {
							name
						}
					}`,
				Results: []map[string]any{
					{
						"name": "John",
					},
					{
						"name": "Islam",
					},
					{
						"name": "Keenan",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryWithIndex_WithLikeFilter_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _like filter",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String @index
					age: Int 
				} 
			`),
			testUtils.Request{
				Request: `
					query {
						users(filter: {name: {_like: "A%"}}) {
							name
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Addo",
					},
					{
						"name": "Andy",
					},
				},
			},
			testUtils.Request{
				Request: `
					query {
						users(filter: {name: {_like: "%d"}}) {
							name
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Fred",
					},
					{
						"name": "Shahzad",
					},
				},
			},
			testUtils.Request{
				Request: `
					query {
						users(filter: {name: {_like: "%e%"}}) {
							name
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Fred",
					},
					{
						"name": "Keenan",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryWithIndex_WithNotLikeFilter_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _nlike filter",
		Actions: []any{
			createSchemaWithDocs(`
				type users {
					name: String @index
					age: Int 
				} 
			`),
			testUtils.Request{
				Request: `
					query {
						users(filter: {name: {_nlike: "%h%"}}) {
							name
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Addo",
					},
					{
						"name": "Andy",
					},
					{
						"name": "Fred",
					},
					{
						"name": "Islam",
					},
					{
						"name": "Keenan",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}
