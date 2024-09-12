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
	"time"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationCreate_WithDefaultValues_NoValuesProvided_SetsDefaultValue(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation, with default values and no values provided",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						active: Boolean @default(bool: true)
						created: DateTime @default(dateTime: "2000-07-23T03:00:00-00:00")
						name: String @default(string: "Bob")
						age: Int @default(int: 40)
						points: Float @default(float: 10)
						metadata: JSON @default(json: "{\"one\":1}")
						image: Blob @default(blob: "ff0099")
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
								created
								metadata
								image
							}
						}`,
				Results: map[string]any{
					"create_Users": []map[string]any{
						{
							"age":      int64(40),
							"active":   true,
							"name":     "Bob",
							"points":   float64(10),
							"created":  time.Time(time.Date(2000, time.July, 23, 3, 0, 0, 0, time.UTC)),
							"metadata": "{\"one\":1}",
							"image":    "ff0099",
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
						active: Boolean @default(bool: true)
						created: DateTime @default(dateTime: "2000-07-23T03:00:00-00:00")
						name: String @default(string: "Bob")
						age: Int @default(int: 40)
						points: Float @default(float: 10)
						metadata: JSON @default(json: "{\"one\":1}")
						image: Blob @default(blob: "ff0099")
					}
				`,
			},
			testUtils.Request{
				Request: `mutation {
							create_Users(input: {age: null, active: null, name: null, points: null, created: null, metadata: null, image: null}) {
								age
								active
								name
								points
								created
								metadata
								image
							}
						}`,
				Results: map[string]any{
					"create_Users": []map[string]any{
						{
							"age":      nil,
							"active":   nil,
							"name":     nil,
							"points":   nil,
							"created":  nil,
							"metadata": nil,
							"image":    nil,
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
						active: Boolean @default(bool: true)
						created: DateTime @default(dateTime: "2000-07-23T03:00:00-00:00")
						name: String @default(string: "Bob")
						age: Int @default(int: 40)
						points: Float @default(float: 10)
						metadata: JSON @default(json: "{\"one\":1}")
						image: Blob @default(blob: "ff0099")
					}
				`,
			},
			testUtils.Request{
				Request: `mutation {
							create_Users(input: {age: 50, active: false, name: "Alice", points: 5, created: "2024-06-18T01:00:00-00:00", metadata: "{\"two\":2}", image: "aabb33"}) {
								age
								active
								name
								points
								created
								metadata
								image
							}
						}`,
				Results: map[string]any{
					"create_Users": []map[string]any{
						{
							"age":      int64(50),
							"active":   false,
							"name":     "Alice",
							"points":   float64(5),
							"created":  time.Time(time.Date(2024, time.June, 18, 1, 0, 0, 0, time.UTC)),
							"metadata": "{\"two\":2}",
							"image":    "aabb33",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithDefaultValue_NoValueProvided_CreatedTwice_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation, with default value, no value provided, and created twice",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String @default(string: "Bob")
						age: Int @default(int: 40)
					}
				`,
			},
			testUtils.Request{
				Request: `mutation {
							create_Users(input: {}) {
								name
								age
							}
						}`,
				Results: map[string]any{
					"create_Users": []map[string]any{
						{
							"name": "Bob",
							"age":  int64(40),
						},
					},
				},
			},
			testUtils.Request{
				Request: `mutation {
							create_Users(input: {}) {
								name
								age
							}
						}`,
				ExpectedError: "a document with the given ID already exists",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithDefaultValue_NoValueProvided_CreatedTwice_UniqueIndex_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation, with default value, no value provided, created twice, and unique index",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String @default(string: "Bob") @index(unique: true)
						age: Int @default(int: 40)
					}
				`,
			},
			testUtils.Request{
				Request: `mutation {
							create_Users(input: {}) {
								name
								age
							}
						}`,
				Results: map[string]any{
					"create_Users": []map[string]any{
						{
							"name": "Bob",
							"age":  int64(40),
						},
					},
				},
			},
			testUtils.Request{
				Request: `mutation {
							create_Users(input: {age: 30}) {
								name
								age
							}
						}`,
				ExpectedError: "can not index a doc's field(s) that violates unique index",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
