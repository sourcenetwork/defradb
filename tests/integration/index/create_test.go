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

	"github.com/sourcenetwork/defradb/internal/request/graphql/schema"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestIndexCreateWithCollection_ShouldNotHinderQuerying(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Creation of index with collection should not hinder querying",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String @indexField
						age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-d4303725-7db9-53d2-b324-f3ee44020e52
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
				Request: `
					query  {
						Users {
							name
							age
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{
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

func TestIndexCreate_ShouldNotHinderQuerying(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Creation of index separately from a collection should not hinder querying",
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
				// bae-d4303725-7db9-53d2-b324-f3ee44020e52
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.CreateIndex{
				CollectionID: 0,
				IndexName:    "some_index",
				FieldName:    "name",
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

func TestIndexCreate_IfInvalidIndexName_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "If invalid index name is provided, return error",
		Actions: []any{
			testUtils.SchemaUpdate{
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
