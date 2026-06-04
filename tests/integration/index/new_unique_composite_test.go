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

func TestUniqueCompositeIndexNew_IfFieldValuesAreNotUnique_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int
						email: String
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21,
						"email": "email@gmail.com"
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21,
						"email": "another@gmail.com"
					}`,
			},
			&action.NewIndex{
				CollectionID:  0,
				Fields:        []client.IndexedFieldDescription{{Name: "name"}, {Name: "age"}},
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

func TestUniqueCompositeIndexNew_UponAddingDocWithExistingFieldValue_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User @index(unique: true, includes: [{field: "name"}, {field: "age"}]) {
						name: String 
						age: Int 
						email: String
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21,
						"email": "email@gmail.com"
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21,
						"email": "another@gmail.com"
					}`,
				ExpectedError: "can not index a doc's field(s) that violates unique index.",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestUniqueCompositeIndexNew_IfFieldValuesAreUnique_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int 
						email: String
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21,
						"email": "some@gmail.com"
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	35,
						"email": "another@gmail.com"
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Andy",
						"age":	35,
						"email": "different@gmail.com"
					}`,
			},
			&action.NewIndex{
				CollectionID: 0,
				Fields:       []client.IndexedFieldDescription{{Name: "name"}, {Name: "age"}},
				IndexName:    "name_age_unique_index",
				Unique:       true,
			},
			&action.ListIndexes{
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

func TestUniqueCompositeIndexNew_IfFieldValuesAreOrdered_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int 
						email: String
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21,
						"email": "some@gmail.com"
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	35,
						"email": "another@gmail.com"
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Andy",
						"age":	35,
						"email": "different@gmail.com"
					}`,
			},
			&action.NewIndex{
				CollectionID: 0,
				Fields: []client.IndexedFieldDescription{
					{Name: "name", Descending: true},
					{Name: "age", Descending: false}, {Name: "email"},
				},
				IndexName: "name_age_unique_index",
				Unique:    true,
			},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "name_age_unique_index",
						ID:     1,
						Unique: true,
						Fields: []client.IndexedFieldDescription{
							{
								Name:       "name",
								Descending: true,
							},
							{
								Name:       "age",
								Descending: false,
							},
							{
								Name:       "email",
								Descending: false,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
