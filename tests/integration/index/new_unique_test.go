// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package index

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestAddUniqueIndex_IfFieldValuesAreNotUnique_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Andy",
						"age":	22
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Shahzad",
						"age":	21
					}`,
			},
			&action.NewIndex{
				CollectionID:  0,
				FieldName:     "age",
				Unique:        true,
				ExpectedError: "can not index a doc's field(s) that violates unique index.",
			},
			&action.ListIndexes{
				CollectionID:    0,
				ExpectedIndexes: []client.IndexDescription{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestUniqueIndexNew_UponAddingDocWithExistingFieldValue_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int @index(unique: true, name: "age_unique_index")
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Shahzad",
						"age":	21
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
				ExpectedError: "can not index a doc's field(s) that violates unique index.",
			},
			&action.Request{
				Request: `query {
					User(filter: {name: {_eq: "John"}}) {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{},
				},
			},
			&action.ListIndexes{
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

func TestUniqueIndexNew_IfFieldValuesAreUnique_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int 
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Shahzad",
						"age":	22
					}`,
			},
			&action.NewIndex{
				CollectionID: 0,
				IndexName:    "age_unique_index",
				FieldName:    "age",
				Unique:       true,
			},
			&action.ListIndexes{
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

func TestUniqueIndexNew_WithMultipleNilFields_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Andy"
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Keenan"
					}`,
			},
			&action.NewIndex{
				CollectionID: 0,
				IndexName:    "age_unique_index",
				FieldName:    "age",
				Unique:       true,
			},
			&action.ListIndexes{
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

func TestUniqueIndexNew_AddingDocWithNilValue_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int @index(unique: true)
					}
				`,
			},
			&action.AddDoc{
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

func TestUniqueIndexNew_UponAddingDocWithExistingNilValue_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int @index(unique: true)
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Keenan"
					}`,
			},
			&action.AddDoc{
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
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						birthday: DateTime @index(unique: true)
					}`,
			},
			&action.AddDoc{
				Doc: `{
						"name":	"Fred",
						"birthday": "2000-07-23T03:00:00-00:00"
					}`,
			},
			&action.AddDoc{
				Doc: `{
						"name":	"Andy",
						"birthday": "2000-07-23T03:00:00-00:00"
					}`,
				ExpectedError: "can not index a doc's field(s) that violates unique index",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
