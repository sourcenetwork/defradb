// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package collection_version

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCollectionVersion_WithDefaultFieldValues(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						active: Boolean @default(bool: true)
						created: DateTime @default(dateTime: "2000-07-23T03:00:00-00:00")
						name: String @default(string: "Bob")
						age: Int @default(int: 10)
						points: Float @default(float: 30)
						metadata: JSON @default(json: "{\"value\":1}")
						image: Blob @default(blob: "ff0099")
					}
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						IsActive:       true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
							},
							{
								Name:         "active",
								DefaultValue: true,
							},
							{
								Name:         "age",
								DefaultValue: float64(10),
							},
							{
								Name:         "created",
								DefaultValue: "2000-07-23T03:00:00Z",
							},
							{
								Name:         "image",
								DefaultValue: "ff0099",
							},
							{
								Name:         "metadata",
								DefaultValue: "{\"value\":1}",
							},
							{
								Name:         "name",
								DefaultValue: "Bob",
							},
							{
								Name:         "points",
								DefaultValue: float64(30),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersion_WithInvalidDefaultFieldValueType_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						active: Boolean @default(bool: invalid)
					}
				`,
				ExpectedError: "Argument \"bool\" has invalid value invalid",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersion_WithIncorrectDefaultFieldValueType_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						active: Boolean @default(int: 10)
					}
				`,
				ExpectedError: "default value type must match field type. Name: active, Expected: bool, Actual: int",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersion_WithMultipleDefaultFieldValueTypes_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String @default(string: "Bob", int: 10, bool: true, float: 10)
					}
				`,
				ExpectedError: "default value must specify one argument. Field: name",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersion_WithDefaultFieldValueOnRelation_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						friend: User @default(string: "Bob")
					}
				`,
				ExpectedError: "default value is not allowed for this field type. Name: friend, Type: User",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersion_WithDefaultFieldValueOnList_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						names: [String] @default(string: "Bob")
					}
				`,
				ExpectedError: "default value is not allowed for this field type. Name: names, Type: List",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
