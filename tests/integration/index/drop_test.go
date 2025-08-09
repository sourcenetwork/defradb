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
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestIndexDrop_ShouldNotHinderQuerying(t *testing.T) {
	test := testUtils.TestCase{
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
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.DropIndex{
				IndexName: "User_name_ASC",
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexDrop_ShouldRemoveIndexFromCollection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String @index
						age: Int @index
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
			testUtils.DropIndex{
				IndexName: "User_age_ASC",
			},
			testUtils.GetIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						ID:   1,
						Name: "User_name_ASC",
						Fields: []client.IndexedFieldDescription{
							{Name: "name"},
						},
					},
				},
			},
			testUtils.DropIndex{
				IndexName: "User_name_ASC",
			},
			testUtils.GetIndexes{
				CollectionID:    0,
				ExpectedIndexes: []client.IndexDescription{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexDrop_IfIndexDoesNotExist_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
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
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.DropIndex{
				CollectionID:  0,
				IndexName:     "non_existing_index",
				ExpectedError: "index with name doesn't exists. Name: non_existing_index",
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
