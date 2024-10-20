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

func TestCreateUniqueCompositeIndex_IfFieldValuesAreNotUnique_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If combination of fields is not unique, creating of unique index fails",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						age: Int
						email: String
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21,
						"email": "email@gmail.com"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21,
						"email": "another@gmail.com"
					}`,
			},
			testUtils.CreateIndex{
				CollectionID: 0,
				Fields:       []testUtils.IndexedField{{Name: "name"}, {Name: "age"}},
				Unique:       true,
				ExpectedError: db.NewErrCanNotIndexNonUniqueFields(
					"bae-c20024f0-bd72-56c2-85d5-865d3aa270b7",
					errors.NewKV("name", "John"), errors.NewKV("age", 21),
				).Error(),
			},
			testUtils.GetIndexes{
				CollectionID:    0,
				ExpectedIndexes: []client.IndexDescription{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestUniqueCompositeIndexCreate_UponAddingDocWithExistingFieldValue_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "adding a new doc with existing field combination for composite index should fail",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "name"}, {field: "age"}]) {
						name: String 
						age: Int 
						email: String
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21,
						"email": "email@gmail.com"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21,
						"email": "another@gmail.com"
					}`,
				ExpectedError: db.NewErrCanNotIndexNonUniqueFields(
					"bae-4da27b71-f735-59f6-b6b8-ea0fa181e3e3",
					errors.NewKV("name", "John"), errors.NewKV("age", 21)).Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestUniqueCompositeIndexCreate_IfFieldValuesAreUnique_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Description: "create unique composite index if all docs have unique fields combinations",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						age: Int 
						email: String
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21,
						"email": "some@gmail.com"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	35,
						"email": "another@gmail.com"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Andy",
						"age":	35,
						"email": "different@gmail.com"
					}`,
			},
			testUtils.CreateIndex{
				CollectionID: 0,
				Fields:       []testUtils.IndexedField{{Name: "name"}, {Name: "age"}},
				IndexName:    "name_age_unique_index",
				Unique:       true,
			},
			testUtils.GetIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "name_age_unique_index",
						ID:     1,
						Unique: true,
						Fields: []client.IndexedFieldDescription{
							{
								Name: "name",
							},
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
