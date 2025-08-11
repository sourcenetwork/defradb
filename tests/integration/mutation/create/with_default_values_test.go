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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"

	"github.com/sourcenetwork/immutable"
)

func TestMutationCreate_WithDefaultValues_NoValuesProvided_SetsDefaultValue(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						active: Boolean @default(bool: true)
						created: DateTime @default(dateTime: "2000-07-23T03:00:00-00:00")
						name: String @default(string: "Bob")
						age: Int @default(int: 40)
						points: Float @default(float: 10)
						points32: Float32 @default(float32: 11)
						points64: Float64 @default(float64: 12)
						metadata: JSON @default(json: "{\"one\":1}")
						image: Blob @default(blob: "ff0099")
					}
				`,
			},
			testUtils.CreateDoc{
				// left empty to test default values
				DocMap: map[string]any{},
			},
			testUtils.Request{
				Request: `query {
					Users {
						age
						active
						name
						points
						points32
						points64
						created
						metadata
						image
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"age":      int64(40),
							"active":   true,
							"name":     "Bob",
							"points":   float64(10),
							"points32": float64(11),
							"points64": float64(12),
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
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						active: Boolean @default(bool: true)
						created: DateTime @default(dateTime: "2000-07-23T03:00:00-00:00")
						name: String @default(string: "Bob")
						age: Int @default(int: 40)
						points: Float @default(float: 10)
						points32: Float32 @default(float32: 11)
						points64: Float64 @default(float64: 12)
						metadata: JSON @default(json: "{\"one\":1}")
						image: Blob @default(blob: "ff0099")
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"age":      nil,
					"active":   nil,
					"name":     nil,
					"points":   nil,
					"points32": nil,
					"points64": nil,
					"created":  nil,
					"metadata": nil,
					"image":    nil,
				},
			},
			testUtils.Request{
				Request: `query {
					Users {
						age
						active
						name
						points
						points32
						points64
						created
						metadata
						image
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"age":      nil,
							"active":   nil,
							"name":     nil,
							"points":   nil,
							"points32": nil,
							"points64": nil,
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
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						active: Boolean @default(bool: true)
						created: DateTime @default(dateTime: "2000-07-23T03:00:00-00:00")
						name: String @default(string: "Bob")
						age: Int @default(int: 40)
						points: Float @default(float: 10)
						points32: Float @default(float: 11)
						points64: Float @default(float: 12)
						metadata: JSON @default(json: "{\"one\":1}")
						image: Blob @default(blob: "ff0099")
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"age":      int64(50),
					"active":   false,
					"name":     "Alice",
					"points":   float64(5),
					"points32": float32(6),
					"points64": float64(7),
					"created":  time.Time(time.Date(2024, time.June, 18, 1, 0, 0, 0, time.UTC)),
					"metadata": "{\"two\":2}",
					"image":    "aabb33",
				},
			},
			testUtils.Request{
				Request: `query {
					Users {
						age
						active
						name
						points
						points32
						points64
						created
						metadata
						image
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"age":      int64(50),
							"active":   false,
							"name":     "Alice",
							"points":   float64(5),
							"points32": float32(6),
							"points64": float64(7),
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
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			// This test will fail if using the collection save
			// method because it does not create two unique docs
			// and instead calls update on the second doc with
			// matching fields
			testUtils.CollectionNamedMutationType,
			testUtils.GQLRequestMutationType,
		}),
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String @default(string: "Bob")
						age: Int @default(int: 40)
					}
				`,
			},
			testUtils.CreateDoc{
				// left empty to test default values
				DocMap: map[string]any{},
			},
			testUtils.CreateDoc{
				// left empty to test default values
				DocMap:        map[string]any{},
				ExpectedError: "a document with the given ID already exists",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithDefaultValue_NoValueProvided_CreatedTwice_UniqueIndex_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			// This test will fail if using the collection save
			// method because it does not create two unique docs
			// and instead calls update on the second doc with
			// matching fields
			testUtils.CollectionNamedMutationType,
			testUtils.GQLRequestMutationType,
		}),
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String @default(string: "Bob") @index(unique: true)
						age: Int @default(int: 40)
					}
				`,
			},
			testUtils.CreateDoc{
				// left empty to test default values
				DocMap: map[string]any{},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"age": int64(50),
				},
				ExpectedError: "can not index a doc's field(s) that violates unique index",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
