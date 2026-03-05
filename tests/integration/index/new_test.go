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
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/internal/request/graphql/schema"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestIndexNewWithCollection_ShouldNotHinderQuerying(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String @index
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `
					query  {
						User {
							name
							age
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
							"age":  int64(21),
						},
					},
				},
			},
			&action.ListIndexes{
				ExpectedIndexes: []client.IndexDescription{
					{
						Name: "User_name_ASC",
						ID:   1,
						Fields: []client.IndexedFieldDescription{
							{
								Name: "name",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexNew_ShouldNotHinderQuerying(t *testing.T) {
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
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.NewIndex{
				IndexName: "some_index",
				FieldName: "name",
			},
			&action.Request{
				Request: `
					query  {
						User {
							name
							age
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
							"age":  int64(21),
						},
					},
				},
			},
			&action.ListIndexes{
				ExpectedIndexes: []client.IndexDescription{
					{
						Name: "some_index",
						ID:   1,
						Fields: []client.IndexedFieldDescription{
							{
								Name: "name",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexNew_IfInvalidIndexName_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						Name: String 
						Age: Int
					}
				`,
			},
			&action.NewIndex{
				CollectionID:  0,
				IndexName:     "!",
				FieldName:     "Name",
				ExpectedError: schema.NewErrIndexWithInvalidName("!").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexNew_IfGivenSameIndexName_ShouldReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User @index(name: "age_index", includes: [{field: "age"}]) @index(name: "age_index", includes: [{field: "age"}]) {
						name: String 
						age: Int @index(name: "age_index")
					}
				`,
				ExpectedError: db.NewErrIndexWithNameAlreadyExists("age_index").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
