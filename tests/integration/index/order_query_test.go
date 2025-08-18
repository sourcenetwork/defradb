// Copyright 2025 Democratized Data Foundation
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

func TestOrderQueryWithIndex_WithAscendingOrder_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(order: {age: ASC}) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						age: Int @index
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Shahzad",
							"age":  int64(20),
						},
						{
							"name": "Bruno",
							"age":  int64(23),
						},
						{
							"name": "Fred",
							"age":  int64(28),
						},
						{
							"name": "John",
							"age":  int64(30),
						},
						{
							"name": "Islam",
							"age":  int64(32),
						},
						{
							"name": "Andy",
							"age":  int64(33),
						},
						{
							"name": "Addo",
							"age":  int64(42),
						},
						{
							"name": "Roy",
							"age":  int64(44),
						},
						{
							"name": "Keenan",
							"age":  int64(48),
						},
						{
							"name": "Chris",
							"age":  int64(55),
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOrderQueryWithIndex_WithLimitDescending_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(order: {age: DESC}, limit: 3) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						age: Int @index
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Chris",
							"age":  int64(55),
						},
						{
							"name": "Keenan",
							"age":  int64(48),
						},
						{
							"name": "Roy",
							"age":  int64(44),
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithLimit().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOrderQueryWithIndex_WithLimitAscending_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(order: {age: ASC}, limit: 3) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						age: Int @index
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Shahzad",
							"age":  int64(20),
						},
						{
							"name": "Bruno",
							"age":  int64(23),
						},
						{
							"name": "Fred",
							"age":  int64(28),
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithLimit().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOrderQueryWithIndex_WithFilterOnNonIndexedFieldAscending_ShouldUseIndexForOrdering(t *testing.T) {
	req := `query {
		User(order: {age: ASC}, filter: {name: {_like: "A%"}}) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						age: Int @index
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Andy",
							"age":  int64(33),
						},
						{
							"name": "Addo",
							"age":  int64(42),
						},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req),
				// we fetch all available docs with index
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOrderQueryWithIndex_WithFilterOnNonIndexedFieldDescending_ShouldUseIndexForOrdering(t *testing.T) {
	req := `query {
		User(order: {age: DESC}, filter: {name: {_like: "A%"}}) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						age: Int @index
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Addo",
							"age":  int64(42),
						},
						{
							"name": "Andy",
							"age":  int64(33),
						},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req),
				// we fetch all available docs with index
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(10),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOrderQueryWithIndex_WithFilterOnIndexedFieldAscending_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(order: {age: ASC}, filter: {age: {_gt: 22}}, limit: 3) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						age: Int @index
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Bruno",
							"age":  int64(23),
						},
						{
							"name": "Fred",
							"age":  int64(28),
						},
						{
							"name": "John",
							"age":  int64(30),
						},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req),
				// we fetch docs starting from the lowest age and skip the first one
				Asserter: testUtils.NewExplainAsserter().WithLimit().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOrderQueryWithIndex_WithFilterOnIndexedFieldDescending_ShouldUseIndex(t *testing.T) {
	req := `query {
		User(order: {age: DESC}, filter: {age: {_lt: 45}}, limit: 3) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						age: Int @index
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Roy",
							"age":  int64(44),
						},
						{
							"name": "Addo",
							"age":  int64(42),
						},
						{
							"name": "Andy",
							"age":  int64(33),
						},
					},
				},
			},
			testUtils.Request{
				Request: makeExplainQuery(req),
				// we fetch docs starting from the highest age, skipping the first 2
				Asserter: testUtils.NewExplainAsserter().WithLimit().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOrderQueryWithIndex_WithOrderOnNestedField_ShouldUseIndexForOrdering(t *testing.T) {
	req := `query {
		User(order: {device: {model: ASC}}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						device: Device 
					}

					type Device {
						model: String @index
						owner: User @primary
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Fred"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Addo"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Shahzad"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "walkman",
					"owner": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "iPhone",
					"owner": testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "pixel",
					"owner": testUtils.NewDocIndex(0, 2),
				},
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Addo"},    // iPhone
						{"name": "Shahzad"}, // pixel
						{"name": "Fred"},    // walkman
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOrderQueryWithIndex_WithOrderOnRelationIDField_ShouldUseIndexForOrdering(t *testing.T) {
	req := `query {
		Device(order: {owner_id: ASC}) {
			model
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						device: Device 
					}

					type Device {
						model: String 
						owner: User @primary @index
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Fred"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Addo"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Shahzad"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "walkman",
					"owner": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "iPhone",
					"owner": testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model": "pixel",
					"owner": testUtils.NewDocIndex(0, 2),
				},
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"Device": []map[string]any{
						{"model": "walkman"},
						{"model": "pixel"},
						{"model": "iPhone"},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOrderQueryWithIndex_WithAscendingQueryOnDescendingIndexedField_ShouldReturnInReverseOrder(t *testing.T) {
	req := `query {
		User(order: {age: ASC}, limit: 3) {
			name
			age
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						age: Int @index(direction: DESC)
					}`,
			},
			testUtils.CreatePredefinedDocs{
				Docs: getUserDocs(),
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Shahzad",
							"age":  int64(20),
						},
						{
							"name": "Bruno",
							"age":  int64(23),
						},
						{
							"name": "Fred",
							"age":  int64(28),
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithLimit().WithIndexFetches(3),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOrderQueryWithCompositeIndex_OrderMismatchASCAndDESC_ShouldNotUserIndex(t *testing.T) {
	req1 := `query {
		User(order: [{name: ASC}, {age: ASC}]) {
			name
			age
		}
	}`

	req2 := `query {
		User(order: [{name: DESC}, {age: DESC}]) {
			name
			age
		}
	}`

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
				type User @index(includes: [{field: "name"},  {field: "age", direction: DESC}]) {
					name: String
					age: Int
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
						"name":	"Alan",
						"age":	29
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24
					}`,
			},
			testUtils.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alan",
							"age":  29,
						},
						{
							"name": "Alice",
							"age":  22,
						},
						{
							"name": "Alice",
							"age":  24,
						},
						{
							"name": "Alice",
							"age":  38,
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithOrder().WithIndexFetches(0),
			},
			testUtils.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alice",
							"age":  22,
						},
						{
							"name": "Alice",
							"age":  24,
						},
						{
							"name": "Alice",
							"age":  38,
						},
						{
							"name": "Alan",
							"age":  29,
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithOrder().WithIndexFetches(0),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOrderQueryWithCompositeIndex_OrderMismatchDESCAndASC_ShouldNotUserIndex(t *testing.T) {
	req1 := `query {
		User(order: [{name: ASC}, {age: ASC}]) {
			name
			age
		}
	}`

	req2 := `query {
		User(order: [{name: DESC}, {age: DESC}]) {
			name
			age
		}
	}`

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User @index(includes: [{field: "name", direction: DESC},  {field: "age"}]) {
						name: String
						age: Int
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
						"name":	"Alan",
						"age":	29
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24
					}`,
			},
			testUtils.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alan",
							"age":  29,
						},
						{
							"name": "Alice",
							"age":  22,
						},
						{
							"name": "Alice",
							"age":  24,
						},
						{
							"name": "Alice",
							"age":  38,
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithOrder().WithIndexFetches(0),
			},
			testUtils.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alice",
							"age":  22,
						},
						{
							"name": "Alice",
							"age":  24,
						},
						{
							"name": "Alice",
							"age":  38,
						},
						{
							"name": "Alan",
							"age":  29,
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithOrder().WithIndexFetches(0),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOrderQueryWithCompositeIndex_OrderMismatchASCAndASC_ShouldNotUserIndex(t *testing.T) {
	req1 := `query {
		User(order: [{name: ASC}, {age: DESC}]) {
			name
			age
		}
	}`

	req2 := `query {
		User(order: [{name: DESC}, {age: ASC}]) {
			name
			age
		}
	}`

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
				type User @index(includes: [{field: "name"},  {field: "age"}]) {
					name: String
					age: Int
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
						"name":	"Alan",
						"age":	29
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24
					}`,
			},
			testUtils.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alan",
							"age":  29,
						},
						{
							"name": "Alice",
							"age":  22,
						},
						{
							"name": "Alice",
							"age":  24,
						},
						{
							"name": "Alice",
							"age":  38,
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithOrder().WithIndexFetches(0),
			},
			testUtils.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alice",
							"age":  22,
						},
						{
							"name": "Alice",
							"age":  24,
						},
						{
							"name": "Alice",
							"age":  38,
						},
						{
							"name": "Alan",
							"age":  29,
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithOrder().WithIndexFetches(0),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOrderQueryWithCompositeIndex_OrderMismatchDESCAndDESC_ShouldNotUserIndex(t *testing.T) {
	req1 := `query {
		User(order: [{name: ASC}, {age: DESC}]) {
			name
			age
		}
	}`

	req2 := `query {
		User(order: [{name: DESC}, {age: ASC}]) {
			name
			age
		}
	}`

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
				type User @index(includes: [{field: "name", direction: DESC},  {field: "age", direction: DESC}]) {
					name: String
					age: Int
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
						"name":	"Alan",
						"age":	29
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24
					}`,
			},
			testUtils.Request{
				Request: req1,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alan",
							"age":  29,
						},
						{
							"name": "Alice",
							"age":  22,
						},
						{
							"name": "Alice",
							"age":  24,
						},
						{
							"name": "Alice",
							"age":  38,
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req1),
				Asserter: testUtils.NewExplainAsserter().WithOrder().WithIndexFetches(0),
			},
			testUtils.Request{
				Request: req2,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alice",
							"age":  22,
						},
						{
							"name": "Alice",
							"age":  24,
						},
						{
							"name": "Alice",
							"age":  38,
						},
						{
							"name": "Alan",
							"age":  29,
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req2),
				Asserter: testUtils.NewExplainAsserter().WithOrder().WithIndexFetches(0),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOrderQueryWithCompositeIndex_WithOrderOnNonIndexInMiddle_ShouldNotUserIndex(t *testing.T) {
	req := `query {
		User(order: [{name: ASC}, {level: ASC}, {age: ASC}]) {
			name
			age
			level
		}
	}`

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
				type User @index(includes: [{field: "name"},  {field: "age"}]) {
					name: String
					age: Int
					level: Int
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22,
						"level": 1
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29,
						"level": 2
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38,
						"level": 3
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24,
						"level": 2
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24,
						"level": 1
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24,
						"level": 3
					}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name":  "Alan",
							"age":   29,
							"level": 2,
						},
						{
							"name":  "Alice",
							"age":   24,
							"level": 1,
						},
						{
							"name":  "Alice",
							"age":   24,
							"level": 2,
						},
						{
							"name":  "Alice",
							"age":   22,
							"level": 1,
						},
						{
							"name":  "Alice",
							"age":   24,
							"level": 3,
						},
						{
							"name":  "Alice",
							"age":   38,
							"level": 3,
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithOrder().WithIndexFetches(0),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOrderQueryWithCompositeIndex_WithOrderOnNonIndexInEnd_ShouldNotUserIndex(t *testing.T) {
	req := `query {
		User(order: [{name: ASC},  {age: ASC}, {level: ASC}]) {
			name
			age
			level
		}
	}`

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
				type User @index(includes: [{field: "name"},  {field: "age"}]) {
					name: String
					age: Int
					level: Int
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	22,
						"level": 1
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alan",
						"age":	29,
						"level": 2
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	38,
						"level": 3
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24,
						"level": 2
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24,
						"level": 1
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Alice",
						"age":	24,
						"level": 3
					}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name":  "Alan",
							"age":   29,
							"level": 2,
						},
						{
							"name":  "Alice",
							"age":   24,
							"level": 1,
						},
						{
							"name":  "Alice",
							"age":   24,
							"level": 2,
						},
						{
							"name":  "Alice",
							"age":   22,
							"level": 1,
						},
						{
							"name":  "Alice",
							"age":   24,
							"level": 3,
						},
						{
							"name":  "Alice",
							"age":   38,
							"level": 3,
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithOrder().WithIndexFetches(0),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOrderQueryWithIndexOnRelation_OrderByPrimaryDoc_ShouldOrderWithIndex(t *testing.T) {
	req := `query {
		User(order: {
			device: {model: ASC}
		}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						device: Device
					} 

					type Device {
						model: String @index
						manufacturer: String
						owner: User @primary
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Fred"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Playstation",
					"manufacturer": "Sony",
					"owner":        testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Andy"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "XBox",
					"manufacturer": "Microsoft",
					"owner":        testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Shahzad"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Arduino",
					"manufacturer": "Arduino",
					"owner":        testUtils.NewDocIndex(0, 2),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Keenan"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "Galaxy",
					"manufacturer": "Samsung",
					"owner":        testUtils.NewDocIndex(0, 3),
				},
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"}, // Arduino
						{"name": "Keenan"},  // Galaxy
						{"name": "Fred"},    // Playstation
						{"name": "Andy"},    // XBox
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(4),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestOrderQueryWithIndexOnRelation_OrderBySecondaryDoc_ShouldOrderWithIndex(t *testing.T) {
	req := `query {
		User(order: {
			device: {model: ASC}
		}) {
			name
		}
	}`
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Device {
						model: String @index
						manufacturer: String
						owner: User 
					}

					type User {
						name: String
						device: Device @primary
					} 
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"model":        "Playstation",
					"manufacturer": "Sony",
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":   "Fred",
					"device": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"model":        "XBox",
					"manufacturer": "Microsoft",
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":   "Andy",
					"device": testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"model":        "Arduino",
					"manufacturer": "Arduino",
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":   "Shahzad",
					"device": testUtils.NewDocIndex(0, 2),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"model":        "Galaxy",
					"manufacturer": "Samsung",
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":   "Keenan",
					"device": testUtils.NewDocIndex(0, 3),
				},
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Shahzad"}, // Arduino
						{"name": "Keenan"},  // Galaxy
						{"name": "Fred"},    // Playstation
						{"name": "Andy"},    // XBox
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(4),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
