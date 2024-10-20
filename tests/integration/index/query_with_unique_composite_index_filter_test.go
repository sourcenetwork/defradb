// Copyright 2024 Democratized Data Foundation
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

func TestQueryWithUniqueCompositeIndex_WithEqualFilter_ShouldFetch(t *testing.T) {
	req1 := `query {
		User(filter: {name: {_eq: "Islam"}}) {
			name
			age
		}
	}`
	req2 := `query {
		User(filter: {name: {_eq: "Islam"}, age: {_eq: 32}}) {
			name
			age
		}
	}`
	req3 := `query {
		User(filter: {name: {_eq: "Islam"}, age: {_eq: 66}}) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
		Description: "Test filtering on composite index with _eq filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Islam",
						"age":	40
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Islam",
						"age":	50
					}`,
			},
			testUtils.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Islam", "age": 32},
						{"name": "Islam", "age": 40},
						{"name": "Islam", "age": 50},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(3),
			},
			testUtils.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Islam", "age": 32},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(1),
			},
			testUtils.Request{
				Request: req3,
				Results: map[string]any{
					"User": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithUniqueCompositeIndex_WithGreaterThanFilterOnFirstField_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {name: {_ne: "Keenan"}, age: {_gt: 44}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _gt filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "age"}, {field: "name"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Chris"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithUniqueCompositeIndex_WithGreaterThanFilterOnSecondField_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {name: {_ne: "Keenan"}, age: {_gt: 44}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _gt filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Chris"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithUniqueCompositeIndex_WithGreaterOrEqualFilterOnFirstField_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {name: {_ne: "Keenan"}, age: {_ge: 44},}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _ge filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "age"}, {field: "name"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Roy"},
						{"name": "Chris"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithUniqueCompositeIndex_WithGreaterOrEqualFilterOnSecondField_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {age: {_ge: 44}, name: {_ne: "Keenan"}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _ge filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Chris"},
						{"name": "Roy"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithUniqueCompositeIndex_WithLessThanFilterOnFirstField_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {age: {_lt: 24}, name: {_ne: "Shahzad"}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _lt filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "age"}, {field: "name"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Bruno"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithUniqueCompositeIndex_WithLessThanFilterOnSecondField_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {age: {_lt: 24}, name: {_ne: "Shahzad"}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _lt filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Bruno"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithUniqueCompositeIndex_WithLessOrEqualFilterOnFirstField_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {age: {_le: 28}, name: {_ne: "Bruno"}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _le filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "age"}, {field: "name"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
						{"name": "Fred"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithUniqueCompositeIndex_WithLessOrEqualFilterOnSecondField_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {age: {_le: 28}, name: {_ne: "Bruno"}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _le filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Fred"},
						{"name": "Shahzad"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithUniqueCompositeIndex_WithNotEqualFilter_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {name: {_ne: "Islam"}, age: {_ne: 28}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _ne filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Addo"},
						{"name": "Andy"},
						{"name": "Bruno"},
						{"name": "Chris"},
						{"name": "John"},
						{"name": "Keenan"},
						{"name": "Roy"},
						{"name": "Shahzad"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithUniqueCompositeIndex_WithInForFirstAndEqForRest_ShouldFetchEfficiently(t *testing.T) {
	req := `query {
		User(filter: {age: {_eq: 33}, name: {_in: ["Addo", "Andy", "Fred"]}}) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _in filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Addo",
						"age":	33
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Addo",
						"age":	88
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Andy",
						"age":	33
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Andy",
						"age":	70
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Andy",
						"age":	51
					}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Addo", "age": 33},
						{"name": "Andy", "age": 33},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithUniqueCompositeIndex_WithInFilter_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {age: {_in: [20, 28, 33]}, name: {_in: ["Addo", "Andy", "Fred"]}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _in filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Addo",
						"age":	10
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Addo",
						"age":	88
					}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
						{"name": "Fred"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(5),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithUniqueCompositeIndex_WithNotInFilter_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {age: {_nin: [20, 23, 28, 42]}, name: {_nin: ["John", "Andy", "Chris"]}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _nin filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Islam"},
						{"name": "Keenan"},
						{"name": "Roy"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithUniqueCompositeIndex_WithLikeFilter_ShouldFetch(t *testing.T) {
	req1 := `query {
			User(filter: {email: {_like: "a%"}, name: {_like: "%o"}}) {
				name
			}
		}`
	req2 := `query {
			User(filter: {email: {_like: "%d@gmail.com"}, name: {_like: "F%"}}) {
				name
			}
		}`
	req3 := `query {
			User(filter: {email: {_like: "%e%"}, name: {_like: "%n%"}}) {
				name
			}
		}`
	req4 := `query {
		User(filter: {email: {_like: "fred@gmail.com"}, name: {_like: "Fred"}}) {
			name
		}
	}`
	req5 := `query {
		User(filter: {email: {_like: "a%@gmail.com"}, name: {_like: "%dd%"}}) {
			name
		}
	}`
	req6 := `query {
		User(filter: {email: {_like: "a%com%m"}}) {
			name
		}
	}`
	req7 := `query {
		User(filter: {email: {_like: "s%"}, name: {_like: "s%h%d"}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _like filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "name"}, {field: "email"}]) {
						name: String 
						email: String 
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Addo"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(10),
			},
			testUtils.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Fred"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(10),
			},
			testUtils.Request{
				Request: req3,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Keenan"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req3),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(10),
			},
			testUtils.Request{
				Request: req4,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Fred"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req4),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(10),
			},
			testUtils.Request{
				Request: req5,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Addo"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req5),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(10),
			},
			testUtils.Request{
				Request: req6,
				Results: map[string]any{
					"User": []map[string]any{},
				},
			},
			testUtils.Request{
				Request: req7,
				Results: map[string]any{
					"User": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithUniqueCompositeIndex_WithNotLikeFilter_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {name: {_nlike: "%h%"}, email: {_nlike: "%d%"}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _nlike filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "name"}, {field: "email"}]) {
						name: String 
						email: String 
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Bruno"},
						{"name": "Islam"},
						{"name": "Keenan"},
						{"name": "Roy"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithUniqueCompositeIndex_WithNotCaseInsensitiveLikeFilter_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {name: {_nilike: "j%"}, email: {_nlike: "%d%"}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _nilike and _nlike filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "name"}, {field: "email"}]) {
						name: String 
						email: String 
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Bruno"},
						{"name": "Chris"},
						{"name": "Islam"},
						{"name": "Keenan"},
						{"name": "Roy"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithUniqueCompositeIndex_IfFirstFieldIsNotInFilter_ShouldNotUseIndex(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test if index is not used when first field is not in filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: `query @explain(type: execute) {
					User(filter: {age: {_eq: 32}}) {
							name
						}
					}`,
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(11).WithIndexFetches(0),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithUniqueCompositeIndex_WithEqualFilterOnNilValueOnFirst_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _eq filter on nil value on first field",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"age":	32
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {name: {_eq: null}}) {
							name
							age
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": nil, "age": 32},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithUniqueCompositeIndex_WithMultipleNilOnFirstFieldAndNilFilter_ShouldFetchAll(t *testing.T) {
	req := `query {
			User(filter: {name: {_eq: null}}) {
				name
				age
				email
			}
		}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _eq filter on nil value on first field with multiple matches",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22,
						"email": "alice@gmail.com"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"age":	32,
						"email": "bob@gmail.com"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"age":	32,
						"email": "cate@gmail.com"
					}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": nil, "age": 32, "email": "bob@gmail.com"},
						{"name": nil, "age": 32, "email": "cate@gmail.com"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(2).WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithUniqueCompositeIndex_WithEqualFilterOnNilValueOnSecond_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _eq filter on nil value on second field",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						about: String
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22,
						"about": "alice_22"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Bob",
						"about": "bob_nil"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"about": "alice_nil"
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {name: {_eq: "Alice"}, age: {_eq: null}}) {
							age
							about
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"age":   nil,
							"about": "alice_nil",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithUniqueCompositeIndex_WithMultipleNilOnSecondFieldsAndNilFilter_ShouldFetchAll(t *testing.T) {
	req := `query {
			User(filter: {name: {_eq: "Bob"}, age: {_eq: null}}) {
				name
				age
				email
			}
		}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _eq filter on nil value on second field",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Bob",
						"age":	22,
						"email": "bob_age_22@gmail.com"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Bob",
						"age":	44,
						"email": "bob_age_44@gmail.com"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Bob",
						"email": "bob1@gmail.com"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Bob",
						"email": "bob2@gmail.com"
					}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Bob", "age": nil, "email": "bob1@gmail.com"},
						{"name": "Bob", "age": nil, "email": "bob2@gmail.com"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(2).WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithUniqueCompositeIndex_WithMultipleNilOnBothFieldsAndNilFilter_ShouldFetchAll(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _eq filter on nil value on both fields",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						about: String
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Bob",
						"age":	22,
						"about": "bob_22"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Bob",
						"about": "bob_nil"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"age":	22,
						"about": "nil_22"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"about": "nil_nil_1"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"about": "nil_nil_2"
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {name: {_eq: null}, age: {_eq: null}}) {
							about
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"about": "nil_nil_2"},
						{"about": "nil_nil_1"},
					},
				},
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {name: {_eq: null}}) {
							about
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"about": "nil_nil_2"},
						{"about": "nil_nil_1"},
						{"about": "nil_22"},
					},
				},
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {age: {_eq: null}}) {
							about
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"about": "bob_nil"},
						{"about": "nil_nil_2"},
						{"about": "nil_nil_1"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithUniqueCompositeIndex_AfterUpdateOnNilFields_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index querying on nil values works after values update",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						about: String
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Bob",
						"age":	22,
						"about": "bob_22 -> bob_nil"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Bob",
						"about": "bob_nil -> nil_nil"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"age":	22,
						"about": "nil_22 -> bob_nil"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"about": "nil_nil -> bob_nil"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"about": "nil_nil -> nil_22"
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `
					{
						"age":	null
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        1,
				Doc: `
					{
						"name": null
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        2,
				Doc: `
					{
						"name": "Bob",
						"age": null
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        3,
				Doc: `
					{
						"name": "Bob"
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        4,
				Doc: `
					{
						"age": 22
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {name: {_eq: null}, age: {_eq: null}}) {
							about
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"about": "bob_nil -> nil_nil"},
					},
				},
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {name: {_eq: null}}) {
							about
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"about": "bob_nil -> nil_nil"},
						{"about": "nil_nil -> nil_22"},
					},
				},
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {age: {_eq: null}}) {
							about
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{"about": "bob_22 -> bob_nil"},
						{"about": "nil_22 -> bob_nil"},
						{"about": "bob_nil -> nil_nil"},
						{"about": "nil_nil -> bob_nil"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithUniqueCompositeIndex_IfMiddleFieldIsNotInFilter_ShouldIgnoreValue(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test composite index with filter without middle field",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "name"}, {field: "email"}, {field: "age"}]) {
						name: String
						email: String
						age: Int
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"email": "alice@gmail.com",
						"age":	22
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"email": "alan@gmail.com",
						"age":	38
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Bob",
						"email": "bob@gmail.com",
						"age":	51
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {name: {_like: "%l%"}, age: {_gt: 30}}) {
							name
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alan",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
