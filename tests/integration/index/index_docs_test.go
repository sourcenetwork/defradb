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
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestIndexDocs_RegularIndex_VerifyIndexEntryCreated(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create index and verify index entry is created in datastore",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String @index
						age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Alice",
					"age": 25
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Bob",
					"age": 30
				}`,
			},
			testUtils.Datastore{
				Key: testUtils.NewKey().
					DatastoreIndex().
					Col(0).
					IndexID(1).
					Field(client.NewNormalString("Alice"), false).
					Field(testUtils.NewDocIndex(0, 0), false),
			},
			testUtils.Datastore{
				Key: testUtils.NewKey().
					DatastoreIndex().
					Col(0).
					IndexID(1).
					Field(client.NewNormalString("Bob"), false).
					Field(testUtils.NewDocIndex(0, 1), false),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndexDocs_CompositeIndex_VerifyIndexEntryCreated(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create composite index and verify index entry is created with DocIndex",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						age: Int
						city: String
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"age": 25,
					"city": "New York"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Jane",
					"age": 30,
					"city": "Boston"
				}`,
			},
			testUtils.CreateIndex{
				CollectionID: 0,
				IndexName:    "name_age_idx",
				Fields: []testUtils.IndexedField{
					{Name: "name", Descending: false},
					{Name: "age", Descending: false},
				},
			},
			testUtils.Datastore{
				Key: testUtils.NewKey().
					DatastoreIndex().
					Col(0).
					IndexID(1).
					Field(client.NewNormalString("John"), false).
					Field(client.NewNormalInt(25), false).
					Field(testUtils.NewDocIndex(0, 0), false),
			},
			testUtils.Datastore{
				Key: testUtils.NewKey().
					DatastoreIndex().
					Col(0).
					IndexID(1).
					Field(client.NewNormalString("Jane"), false).
					Field(client.NewNormalInt(30), false).
					Field(testUtils.DocIndex{CollectionIndex: 0, Index: 1}, false),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
