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

package one_to_many

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

var userCollection = `
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
		Actions: []any{
			&action.AddCollection{
				SDL: userCollection,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John",
					"age": 30
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "John and the philosopher are stoned",
					"rating":    9.9,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "John has a chamber of secrets",
					"rating":    9.9,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.Request{
				Request: `mutation {
					delete_Book(docID: "bae-227565a8-81b1-5c96-90e2-30dbe75ad5bd") {
							_docID
						}
					}`,
				Results: map[string]any{
					"delete_Book": []map[string]any{
						{
							"_docID": "bae-227565a8-81b1-5c96-90e2-30dbe75ad5bd",
						},
					},
				},
			},
			&action.Request{
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
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
