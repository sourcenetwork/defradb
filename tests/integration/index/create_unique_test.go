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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/db"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

const johnDocID = "bae-774fbeea-813b-52c8-82b0-d08515a075d7"

func TestCreateUniqueIndex_IfFieldValuesAreNotUnique_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If field is not unique, creating of unique index fails",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Andy",
						"age":	22
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Shahzad",
						"age":	21
					}`,
			},
			testUtils.CreateIndex{
				CollectionID: 0,
				FieldName:    "age",
				Unique:       true,
				ExpectedError: db.NewErrCanNotIndexNonUniqueFields(
					johnDocID, errors.NewKV("age", 21)).Error(),
			},
			testUtils.GetIndexes{
				CollectionID:    0,
				ExpectedIndexes: []client.IndexDescription{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestUniqueIndexCreate_UponAddingDocWithExistingFieldValue_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "adding a new doc with existing value for indexed field should fail",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						age: Int @index(unique: true, name: "age_unique_index")
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Shahzad",
						"age":	21
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
				ExpectedError: db.NewErrCanNotIndexNonUniqueFields(
					johnDocID, errors.NewKV("age", 21)).Error(),
			},
			testUtils.Request{
				Request: `query {
					User(filter: {name: {_eq: "John"}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{},
				},
			},
			testUtils.GetIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "age_unique_index",
						ID:     1,
						Unique: true,
						Fields: []client.IndexedFieldDescription{
							{
								Name: "age",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestUniqueIndexCreate_IfFieldValuesAreUnique_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Description: "create unique index if all docs have unique field values",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						age: Int 
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Shahzad",
						"age":	22
					}`,
			},
			testUtils.CreateIndex{
				CollectionID: 0,
				IndexName:    "age_unique_index",
				FieldName:    "age",
				Unique:       true,
			},
			testUtils.GetIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "age_unique_index",
						ID:     1,
						Unique: true,
						Fields: []client.IndexedFieldDescription{
							{
								Name: "age",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestUniqueIndexCreate_WithMultipleNilFields_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If filter does not match any document, return empty result",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Andy"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Keenan"
					}`,
			},
			testUtils.CreateIndex{
				CollectionID: 0,
				IndexName:    "age_unique_index",
				FieldName:    "age",
				Unique:       true,
			},
			testUtils.GetIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "age_unique_index",
						ID:     1,
						Unique: true,
						Fields: []client.IndexedFieldDescription{
							{
								Name: "age",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestUniqueIndexCreate_AddingDocWithNilValue_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test adding a doc with nil value for indexed field should succeed",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						age: Int @index(unique: true)
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John"
					}`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestUniqueIndexCreate_UponAddingDocWithExistingNilValue_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If filter does not match any document, return empty result",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						age: Int @index(unique: true)
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Keenan"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Andy"
					}`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestUniqueQueryWithIndex_UponAddingDocWithSameDateTime_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						birthday: DateTime @index(unique: true)
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"name":	"Fred",
						"birthday": "2000-07-23T03:00:00-00:00"
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"name":	"Andy",
						"birthday": "2000-07-23T03:00:00-00:00"
					}`,
				ExpectedError: db.NewErrCanNotIndexNonUniqueFields(
					"bae-2000529a-8b27-539b-91e9-c35f431fb78e",
					errors.NewKV("birthday", testUtils.MustParseTime("2000-07-23T03:00:00-00:00")),
				).Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
