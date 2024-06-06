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

func TestQueryWithCompositeIndex_WithEqualFilter_ShouldFetch(t *testing.T) {
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
					type User @index(fields: ["name", "age"]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req1,
				Results: []map[string]any{
					{"name": "Islam", "age": 32},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(1),
			},
			testUtils.Request{
				Request: req2,
				Results: []map[string]any{
					{"name": "Islam", "age": 32},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(1),
			},
			testUtils.Request{
				Request: req3,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithGreaterThanFilterOnFirstField_ShouldFetch(t *testing.T) {
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
					type User @index(fields: ["age", "name"]) {
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
				Results: []map[string]any{
					{"name": "Chris"},
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

func TestQueryWithCompositeIndex_WithGreaterThanFilterOnSecondField_ShouldFetch(t *testing.T) {
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
					type User @index(fields: ["name", "age"]) {
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
				Results: []map[string]any{
					{"name": "Chris"},
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

func TestQueryWithCompositeIndex_WithGreaterOrEqualFilterOnFirstField_ShouldFetch(t *testing.T) {
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
					type User @index(fields: ["age", "name"]) {
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
				Results: []map[string]any{
					{"name": "Roy"},
					{"name": "Chris"},
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

func TestQueryWithCompositeIndex_WithGreaterOrEqualFilterOnSecondField_ShouldFetch(t *testing.T) {
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
					type User @index(fields: ["name", "age"]) {
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
				Results: []map[string]any{
					{"name": "Chris"},
					{"name": "Roy"},
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

func TestQueryWithCompositeIndex_WithLessThanFilterOnFirstField_ShouldFetch(t *testing.T) {
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
					type User @index(fields: ["age", "name"]) {
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
				Results: []map[string]any{
					{"name": "Bruno"},
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

func TestQueryWithCompositeIndex_WithLessThanFilterOnSecondField_ShouldFetch(t *testing.T) {
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
					type User @index(fields: ["name", "age"]) {
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
				Results: []map[string]any{
					{"name": "Bruno"},
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

func TestQueryWithCompositeIndex_WithLessOrEqualFilterOnFirstField_ShouldFetch(t *testing.T) {
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
					type User @index(fields: ["age", "name"]) {
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
				Results: []map[string]any{
					{"name": "Shahzad"},
					{"name": "Fred"},
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

func TestQueryWithCompositeIndex_WithLessOrEqualFilterOnSecondField_ShouldFetch(t *testing.T) {
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
					type User @index(fields: ["name", "age"]) {
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
				Results: []map[string]any{
					{"name": "Fred"},
					{"name": "Shahzad"},
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

func TestQueryWithCompositeIndex_WithNotEqualFilter_ShouldFetch(t *testing.T) {
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
					type User @index(fields: ["name", "age"]) {
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
				Results: []map[string]any{
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
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithInFilter_ShouldFetch(t *testing.T) {
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
					type User @index(fields: ["name", "age"]) {
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
				Results: []map[string]any{
					{"name": "Andy"},
					{"name": "Fred"},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithNotInFilter_ShouldFetch(t *testing.T) {
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
					type User @index(fields: ["name", "age"]) {
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
				Results: []map[string]any{
					{"name": "Islam"},
					{"name": "Keenan"},
					{"name": "Roy"},
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

func TestQueryWithCompositeIndex_WithLikeFilter_ShouldFetch(t *testing.T) {
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
					type User @index(fields: ["name", "email"]) {
						name: String 
						email: String 
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req1,
				Results: []map[string]any{
					{"name": "Addo"},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(10),
			},
			testUtils.Request{
				Request: req2,
				Results: []map[string]any{
					{"name": "Fred"},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(10),
			},
			testUtils.Request{
				Request: req3,
				Results: []map[string]any{
					{"name": "Keenan"},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req3),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(10),
			},
			testUtils.Request{
				Request: req4,
				Results: []map[string]any{
					{"name": "Fred"},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req4),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(10),
			},
			testUtils.Request{
				Request: req5,
				Results: []map[string]any{
					{"name": "Addo"},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req5),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(0).WithIndexFetches(10),
			},
			testUtils.Request{
				Request: req6,
				Results: []map[string]any{},
			},
			testUtils.Request{
				Request: req7,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithNotLikeFilter_ShouldFetch(t *testing.T) {
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
					type User @index(fields: ["name", "email"]) {
						name: String 
						email: String 
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: []map[string]any{
					{"name": "Bruno"},
					{"name": "Islam"},
					{"name": "Keenan"},
					{"name": "Roy"},
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

func TestQueryWithCompositeIndex_IfFirstFieldIsNotInFilter_ShouldNotUseIndex(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test if index is not used when first field is not in filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(fields: ["name", "age"]) {
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

func TestQueryWithCompositeIndex_WithEqualFilterOnNilValueOnFirst_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _eq filter on nil value on first field",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(fields: ["name", "age"]) {
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
				Results: []map[string]any{
					{"name": nil, "age": 32},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithEqualFilterOnNilValueOnSecond_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test index filtering with _eq filter on nil value on second field",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(fields: ["name", "age"]) {
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
						"name":	"Bob"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice"
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {name: {_eq: "Alice"}, age: {_eq: null}}) {
							name
							age
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Alice",
						"age":  nil,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_IfMiddleFieldIsNotInFilter_ShouldIgnoreValue(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test composite index with filter without middle field",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(fields: ["name", "email", "age"]) {
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
				Results: []map[string]any{
					{
						"name": "Alan",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_IfConsecutiveEqOps_ShouldUseAllToOptimizeQuery(t *testing.T) {
	reqWithName := `query {
			User(filter: {name: {_eq: "Bob"}}) {
				about
			}
		}`
	reqWithNameAge := `query {
			User(filter: {name: {_eq: "Bob"}, age: {_eq: 22}}) {
				about
			}
		}`
	reqWithNameAgeNumChildren := `query {
			User(filter: {name: {_eq: "Bob"}, age: {_eq: 22}, numChildren: {_eq: 2}}) {
				about
			}
		}`
	test := testUtils.TestCase{
		Description: "Test index filtering with _eq filter on nil value on second field",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(fields: ["name", "age", "numChildren"]) {
						name: String
						age: Int
						numChildren: Int
						about: String
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Bob",
						"age":	22,
						"numChildren": 2,
						"about": "bob1"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Bob",
						"age":	22,
						"numChildren": 2,
						"about": "bob2"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Bob",
						"age":	22,
						"numChildren": 0,
						"about": "bob3"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Bob",
						"age":	44,
						"numChildren": 2,
						"about": "bob4"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22,
						"numChildren": 2,
						"about": "alice"
					}`,
			},
			testUtils.Request{
				Request: reqWithName,
				Results: []map[string]any{
					{"about": "bob3"},
					{"about": "bob2"},
					{"about": "bob1"},
					{"about": "bob4"},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(reqWithName),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(4).WithIndexFetches(4),
			},
			testUtils.Request{
				Request: reqWithNameAge,
				Results: []map[string]any{
					{"about": "bob3"},
					{"about": "bob2"},
					{"about": "bob1"},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(reqWithNameAge),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(3).WithIndexFetches(3),
			},
			testUtils.Request{
				Request: reqWithNameAgeNumChildren,
				Results: []map[string]any{
					{"about": "bob2"},
					{"about": "bob1"},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(reqWithNameAgeNumChildren),
				Asserter: testUtils.NewExplainAsserter().WithFieldFetches(2).WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
