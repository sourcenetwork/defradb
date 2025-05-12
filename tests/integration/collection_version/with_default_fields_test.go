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
								Kind: client.FieldKind_DocID,
							},
							{
								Name:         "active",
								Kind:         client.FieldKind_NILLABLE_BOOL,
								Typ:          client.LWW_REGISTER,
								DefaultValue: true,
							},
							{
								Name:         "age",
								Kind:         client.FieldKind_NILLABLE_INT,
								Typ:          client.LWW_REGISTER,
								DefaultValue: float64(10),
							},
							{
								Name:         "created",
								Kind:         client.FieldKind_NILLABLE_DATETIME,
								Typ:          client.LWW_REGISTER,
								DefaultValue: "2000-07-23T03:00:00Z",
							},
							{
								Name:         "image",
								Kind:         client.FieldKind_NILLABLE_BLOB,
								Typ:          client.LWW_REGISTER,
								DefaultValue: "ff0099",
							},
							{
								Name:         "metadata",
								Kind:         client.FieldKind_NILLABLE_JSON,
								Typ:          client.LWW_REGISTER,
								DefaultValue: "{\"value\":1}",
							},
							{
								Name:         "name",
								Kind:         client.FieldKind_NILLABLE_STRING,
								Typ:          client.LWW_REGISTER,
								DefaultValue: "Bob",
							},
							{
								Name:         "points",
								Kind:         client.FieldKind_NILLABLE_FLOAT64,
								Typ:          client.LWW_REGISTER,
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
