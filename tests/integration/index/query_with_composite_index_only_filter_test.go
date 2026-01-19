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

	"github.com/sourcenetwork/defradb/tests/action"
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
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Islam", "age": 32},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
			&action.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Islam", "age": 32},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
			&action.Request{
				Request: req3,
				Results: map[string]any{
					"User": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithGreaterThanFilterOnFirstField_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {name: {_neq: "Keenan"}, age: {_gt: 44}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "age"}, {field: "name"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Chris"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithGreaterThanFilterOnSecondField_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {name: {_neq: "Keenan"}, age: {_gt: 44}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Chris"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithGreaterOrEqualFilterOnFirstField_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {name: {_neq: "Keenan"}, age: {_geq: 44},}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "age"}, {field: "name"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Roy"},
						{"name": "Chris"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithGreaterOrEqualFilterOnSecondField_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {age: {_geq: 44}, name: {_neq: "Keenan"}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Chris"},
						{"name": "Roy"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithLessThanFilterOnFirstField_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {age: {_lt: 24}, name: {_neq: "Shahzad"}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "age"}, {field: "name"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Bruno"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithLessThanFilterOnSecondField_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {age: {_lt: 24}, name: {_neq: "Shahzad"}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Bruno"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithLessOrEqualFilterOnFirstField_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {age: {_leq: 28}, name: {_neq: "Bruno"}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "age"}, {field: "name"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"},
						{"name": "Fred"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithLessOrEqualFilterOnSecondField_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {age: {_leq: 28}, name: {_neq: "Bruno"}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Fred"},
						{"name": "Shahzad"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithNotEqualFilter_ShouldFetch(t *testing.T) {
	req := `query {
		User(filter: {name: {_neq: "Islam"}, age: {_neq: 28}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
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
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(10),
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
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Andy"},
						{"name": "Fred"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
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
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Islam"},
						{"name": "Keenan"},
						{"name": "Roy"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(10),
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
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name"}, {field: "email"}]) {
						name: String 
						email: String 
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Addo"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(10),
			},
			&action.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Fred"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(10),
			},
			&action.Request{
				Request: req3,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Keenan"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req3),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(10),
			},
			&action.Request{
				Request: req4,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Fred"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req4),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(10),
			},
			&action.Request{
				Request: req5,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Addo"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req5),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(10),
			},
			&action.Request{
				Request: req6,
				Results: map[string]any{
					"User": []map[string]any{},
				},
			},
			&action.Request{
				Request: req7,
				Results: map[string]any{
					"User": []map[string]any{},
				},
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
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name"}, {field: "email"}]) {
						name: String 
						email: String 
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
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
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_IfFirstFieldIsNotInFilter_ShouldNotUseIndex(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name"}, {field: "age"}]) {
						name: String
						age: Int
						email: String
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			&action.Request{
				Request: `query @explain(type: execute) {
					User(filter: {age: {_eq: 32}}) {
							name
						}
					}`,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_WithEqualFilterOnNilValueOnFirst_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name"}, {field: "age"}]) {
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
			&action.Request{
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

func TestQueryWithCompositeIndex_WithEqualFilterOnNilValueOnSecond_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name"}, {field: "age"}]) {
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
			&action.Request{
				Request: `
					query {
						User(filter: {name: {_eq: "Alice"}, age: {_eq: null}}) {
							name
							age
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alice",
							"age":  nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithCompositeIndex_IfMiddleFieldIsNotInFilter_ShouldIgnoreValue(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name"}, {field: "email"}, {field: "age"}]) {
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
			&action.Request{
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
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name"}, {field: "age"}, {field: "numChildren"}]) {
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
			&action.Request{
				Request: reqWithName,
				Results: map[string]any{
					"User": []map[string]any{
						{"about": "bob3"},
						{"about": "bob1"},
						{"about": "bob2"},
						{"about": "bob4"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(reqWithName),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(4),
			},
			&action.Request{
				Request: reqWithNameAge,
				Results: map[string]any{
					"User": []map[string]any{
						{"about": "bob3"},
						{"about": "bob1"},
						{"about": "bob2"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(reqWithNameAge),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
			&action.Request{
				Request: reqWithNameAgeNumChildren,
				Results: map[string]any{
					"User": []map[string]any{
						{"about": "bob1"},
						{"about": "bob2"},
					},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(reqWithNameAgeNumChildren),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
