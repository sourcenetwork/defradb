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

func TestQueryWithIndex_IfIntFieldInDescOrder_ShouldFetchInRevertedOrder(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If indexed int field is in DESC order, it should be fetched in reverted order",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						age: Int @indexField(direction: DESC)
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
				Request: `
					query {
						User(filter: {age: {_gt: 1}}) {
							name
							age
						}
					}`,
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryWithIndex_IfFloatFieldInDescOrder_ShouldFetchInRevertedOrder(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If indexed float field is in DESC order, it should be fetched in reverted order",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						iq: Float @indexField(direction: DESC)
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
							"iq":   0.4,
						},
						{
							"name": "Kate",
							"iq":   0.3,
						},
						{
							"name": "Alice",
							"iq":   0.2,
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
						name: String @indexField(direction: DESC)
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
