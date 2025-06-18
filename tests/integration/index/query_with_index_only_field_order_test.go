// Copyright 2023 Democratized Data Foundation
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

func TestQueryWithIndex_IfIntFieldInDescOrderWithGt_ShouldFetchInRevertedOrder(t *testing.T) {
	req := `
		query {
			User(filter: {age: {_gt: 20}}) {
				name
				age
			}
		}`

	test := testUtils.TestCase{
		Description: "If indexed int field is in DESC order with _gt, it should be fetched in reverted order",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						age: Int @index(direction: DESC)
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"John",
						"age":	20
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Fred",
						"age":	18
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Bob",
						"age":	24
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Kate",
						"age":	23
					}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Bob",
							"age":  24,
						},
						{
							"name": "Kate",
							"age":  23,
						},
						{
							"name": "Alice",
							"age":  22,
						},
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

func TestQueryWithIndex_IfIntFieldInDescOrderWithGe_ShouldFetchInRevertedOrder(t *testing.T) {
	req := `
		query {
			User(filter: {age: {_ge: 22}}) {
				name
				age
			}
		}`

	test := testUtils.TestCase{
		Description: "If indexed int field is in DESC order with _ge, it should be fetched in reverted order",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						age: Int @index(direction: DESC)
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"John",
						"age":	20
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Fred",
						"age":	18
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Bob",
						"age":	24
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Kate",
						"age":	23
					}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Bob",
							"age":  24,
						},
						{
							"name": "Kate",
							"age":  23,
						},
						{
							"name": "Alice",
							"age":  22,
						},
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

func TestQueryWithIndex_IfIntFieldInDescOrderWithLt_ShouldFetchInRevertedOrder(t *testing.T) {
	req := `
		query {
			User(filter: {age: {_lt: 22}}) {
				name
				age
			}
		}`

	test := testUtils.TestCase{
		Description: "If indexed int field is in DESC order with _lt, it should be fetched in reverted order",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						age: Int @index(direction: DESC)
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"John",
						"age":	20
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Fred",
						"age":	18
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Bob",
						"age":	24
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Kate",
						"age":	23
					}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
							"age":  20,
						},
						{
							"name": "Fred",
							"age":  18,
						},
					},
				},
			},
			testUtils.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_IfIntFieldInDescOrderWithLe_ShouldFetchInRevertedOrder(t *testing.T) {
	req := `
		query {
			User(filter: {age: {_le: 22}}) {
				name
				age
			}
		}`

	test := testUtils.TestCase{
		Description: "If indexed int field is in DESC order with _le, it should be fetched in reverted order",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						age: Int @index(direction: DESC)
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Alice",
						"age":	22
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"John",
						"age":	20
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Fred",
						"age":	18
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Bob",
						"age":	24
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Kate",
						"age":	23
					}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Alice",
							"age":  22,
						},
						{
							"name": "John",
							"age":  20,
						},
						{
							"name": "Fred",
							"age":  18,
						},
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

func TestQueryWithIndex_IfFloatFieldInDescOrderWithLt_ShouldFetchInRevertedOrder(t *testing.T) {
	req := `
		query {
			User(filter: {iq: {_lt: 0.35}}) {
				name
				iq
			}
		}`

	test := testUtils.TestCase{
		Description: "If indexed float field is in DESC order with _lt, it should be fetched in reverted order",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						iq: Float @index(direction: DESC)
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Alice",
						"iq":	0.2
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Bob",
						"iq":	0.4
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Kate",
						"iq":	0.3
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"David",
						"iq":	0.5
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Emma",
						"iq":	0.1
					}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Kate",
							"iq":   0.3,
						},
						{
							"name": "Alice",
							"iq":   0.2,
						},
						{
							"name": "Emma",
							"iq":   0.1,
						},
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

func TestQueryWithIndex_IfFloatFieldInDescOrderWithGt_ShouldFetchInRevertedOrder(t *testing.T) {
	req := `
		query {
			User(filter: {iq: {_gt: 0.25}}) {
				name
				iq
			}
		}`

	test := testUtils.TestCase{
		Description: "If indexed float field is in DESC order with _gt, it should be fetched in reverted order",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						iq: Float @index(direction: DESC)
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Alice",
						"iq":	0.2
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Bob",
						"iq":	0.4
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Kate",
						"iq":	0.3
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"David",
						"iq":	0.5
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Emma",
						"iq":	0.1
					}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "David",
							"iq":   0.5,
						},
						{
							"name": "Bob",
							"iq":   0.4,
						},
						{
							"name": "Kate",
							"iq":   0.3,
						},
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

func TestQueryWithIndex_IfFloatFieldInDescOrderWithGe_ShouldFetchInRevertedOrder(t *testing.T) {
	req := `
		query {
			User(filter: {iq: {_ge: 0.3}}) {
				name
				iq
			}
		}`

	test := testUtils.TestCase{
		Description: "If indexed float field is in DESC order with _ge, it should be fetched in reverted order",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						iq: Float @index(direction: DESC)
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Alice",
						"iq":	0.2
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Bob",
						"iq":	0.4
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Kate",
						"iq":	0.3
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"David",
						"iq":	0.5
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Emma",
						"iq":	0.1
					}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "David",
							"iq":   0.5,
						},
						{
							"name": "Bob",
							"iq":   0.4,
						},
						{
							"name": "Kate",
							"iq":   0.3,
						},
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

func TestQueryWithIndex_IfFloatFieldInDescOrderWithLe_ShouldFetchInRevertedOrder(t *testing.T) {
	req := `
		query {
			User(filter: {iq: {_le: 0.3}}) {
				name
				iq
			}
		}`

	test := testUtils.TestCase{
		Description: "If indexed float field is in DESC order with _le, it should be fetched in reverted order",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						iq: Float @index(direction: DESC)
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Alice",
						"iq":	0.2
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Bob",
						"iq":	0.4
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Kate",
						"iq":	0.3
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"David",
						"iq":	0.5
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Emma",
						"iq":	0.1
					}`,
			},
			testUtils.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Kate",
							"iq":   0.3,
						},
						{
							"name": "Alice",
							"iq":   0.2,
						},
						{
							"name": "Emma",
							"iq":   0.1,
						},
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

func TestQueryWithIndex_IfFloat32FieldInDescOrder_ShouldFetchInRevertedOrder(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If indexed float32 field is in DESC order, it should be fetched in reverted order",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						iq: Float32 @index(direction: DESC)
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Alice",
						"iq":	0.2
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Bob",
						"iq":	0.4
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Kate",
						"iq":	0.3
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {iq: {_lt: 1}}) {
							name
							iq
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Bob",
							"iq":   float32(0.4),
						},
						{
							"name": "Kate",
							"iq":   float32(0.3),
						},
						{
							"name": "Alice",
							"iq":   float32(0.2),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_IfStringFieldInDescOrder_ShouldFetchInRevertedOrder(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If indexed string field is in DESC order, it should be fetched in reverted order",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String @index(direction: DESC)
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Alice"
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Aaron"
					}`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"Andy"
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						User(filter: {name: {_like: "A%"}}) {
							name
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Andy",
						},
						{
							"name": "Alice",
						},
						{
							"name": "Aaron",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
