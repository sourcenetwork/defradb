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
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/internal/request/graphql/schema"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestIndexCreateWithCollection_ShouldNotHinderQuerying(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Creation of index with collection should not hinder querying",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String @index
						age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
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
			testUtils.GetIndexes{
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

func TestIndexCreate_ShouldNotHinderQuerying(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Creation of index separately from a collection should not hinder querying",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.CreateIndex{
				IndexName: "some_index",
				FieldName: "name",
			},
			testUtils.Request{
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
			testUtils.GetIndexes{
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

func TestIndexCreate_IfInvalidIndexName_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If invalid index name is provided, return error",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						Name: String 
						Age: Int
					}
				`,
			},
			testUtils.CreateIndex{
				CollectionID:  0,
				IndexName:     "!",
				FieldName:     "Name",
				ExpectedError: schema.NewErrIndexWithInvalidName("!").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexCreate_IfGivenSameIndexName_ShouldReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If given same index name, should return error",
		Actions: []any{
			&action.AddSchema{
				Schema: `
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
