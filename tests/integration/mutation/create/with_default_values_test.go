// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package create

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationCreate_WithDefaultValues_NoValuesProvided_SetsDefaultValue(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation, with default values and no values provided",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						age: Int @default(int: 40)
						active: Boolean @default(bool: true)
						name: String @default(string: "Bob")
						points: Float @default(float: 10)
					}
				`,
			},
			testUtils.Request{
				Request: `mutation {
							create_Users(input: {}) {
								age
								active
								name
								points
							}
						}`,
				Results: map[string]any{
					"create_Users": []map[string]any{
						{
							"age":    int64(40),
							"active": true,
							"name":   "Bob",
							"points": float64(10),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithDefaultValues_NilValuesProvided_SetsNilValue(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation, with default values and null values provided",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						age: Int @default(int: 40)
						active: Boolean @default(bool: true)
						name: String @default(string: "Bob")
						points: Float @default(float: 10)
					}
				`,
			},
			testUtils.Request{
				Request: `mutation {
							create_Users(input: {age: null, active: null, name: null, points: null}) {
								age
								active
								name
								points
							}
						}`,
				Results: map[string]any{
					"create_Users": []map[string]any{
						{
							"age":    nil,
							"active": nil,
							"name":   nil,
							"points": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithDefaultValues_ValuesProvided_SetsValue(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation, with default values and values provided",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						age: Int @default(int: 40)
						active: Boolean @default(bool: true)
						name: String @default(string: "Bob")
						points: Float @default(float: 10)
					}
				`,
			},
			testUtils.Request{
				Request: `mutation {
							create_Users(input: {age: 50, active: false, name: "Alice", points: 5}) {
								age
								active
								name
								points
							}
						}`,
				Results: map[string]any{
					"create_Users": []map[string]any{
						{
							"age":    int64(50),
							"active": false,
							"name":   "Alice",
							"points": float64(5),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
