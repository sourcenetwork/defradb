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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestDeletionOfADocumentUsingSingleKeyWithShowDeletedDocumentQuery(t *testing.T) {
	jsonString1 := `{
		"name": "John",
		"age": 30
	}`
	doc1, err := client.NewDocFromJSON([]byte(jsonString1))
	require.NoError(t, err)

	jsonString2 := fmt.Sprintf(`{
		"name": "John and the philosopher are stoned",
		"rating": 9.9,
		"author_id": "%s"
	}`, doc1.Key())
	doc2, err := client.NewDocFromJSON([]byte(jsonString2))
	require.NoError(t, err)

	jsonString3 := fmt.Sprintf(`{
		"name": "John has a chamber of secrets",
		"rating": 9.9,
		"author_id": "%s"
	}`, doc1.Key())
	// doc3, err := client.NewDocFromJSON([]byte(jsonString1))
	// require.NoError(t, err)

	test := testUtils.TestCase{
		Description: "One to many delete document using single key show deleted.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
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
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc:          jsonString1,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          jsonString2,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          jsonString3,
			},
			testUtils.Request{
				Request: fmt.Sprintf(`mutation {
						delete_Book(id: "%s") {
							_key
						}
					}`, doc2.Key()),
				Results: []map[string]any{
					{
						"_key": doc2.Key().String(),
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
				Results: []map[string]any{
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
	}

	testUtils.ExecuteTestCase(t, test)
}
