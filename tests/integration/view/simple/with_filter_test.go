// Copyright 2023 Democratized Data Foundation
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

func TestView_SimpleWithFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple view with filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			testUtils.CreateView{
				Query: `
					User(filter: {name: {_eq: "John"}}) {
						name
					}
				`,
				SDL: `
					type UserView {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"Fred"
				}`,
			},
			testUtils.Request{
				Request: `query {
							UserView {
								name
							}
						}`,
				Results: []map[string]any{
					{
						"name": "John",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_SimpleWithFilterOnViewAndQuery(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple view with filter",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			testUtils.CreateView{
				Query: `
					User(filter: {name: {_eq: "John"}}) {
						name
						age
					}
				`,
				SDL: `
					type UserView {
						name: String
						age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"John",
					"age": 30
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"John",
					"age": 31
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name":	"Fred",
					"age": 31
				}`,
			},
			testUtils.Request{
				Request: `query {
							UserView(filter: {age: {_eq: 31}}) {
								name
							}
						}`,
				Results: []map[string]any{
					{
						"name": "John",
						"age":  31,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
