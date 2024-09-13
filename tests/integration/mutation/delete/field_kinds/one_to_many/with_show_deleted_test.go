// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_many

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

var schemas = `
type Book {
	name: String
	rating: Float
	author: Author
}
type Author {
	name: String
	age: Int
	published: [Book]
}
`

func TestDeletionOfADocumentUsingSingleDocIDWithShowDeletedDocumentQuery(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to many delete document using single document id, show deleted.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: schemas,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John",
					"age": 30
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "John and the philosopher are stoned",
					"rating":    9.9,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "John has a chamber of secrets",
					"rating":    9.9,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.Request{
				Request: `mutation {
					delete_Book(docID: "bae-39db1d4b-72c0-5b7b-b6f2-c20870982128") {
							_docID
						}
					}`,
				Results: map[string]any{
					"delete_Book": []map[string]any{
						{
							"_docID": "bae-39db1d4b-72c0-5b7b-b6f2-c20870982128",
						},
					},
				},
			},
			testUtils.Request{
				Request: `query {
						Author(showDeleted: true) {
							_deleted
							name
							age
							published {
								_deleted
								name
								rating
							}
						}
					}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"_deleted": false,
							"name":     "John",
							"age":      int64(30),
							"published": []map[string]any{
								{
									"_deleted": true,
									"name":     "John and the philosopher are stoned",
									"rating":   9.9,
								},
								{
									"_deleted": false,
									"name":     "John has a chamber of secrets",
									"rating":   9.9,
								},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
