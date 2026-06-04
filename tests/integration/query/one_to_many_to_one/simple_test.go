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

package one_to_many_to_one

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToOneRelations(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			gqlSchemaOneToManyToOne(),
			// Authors
			&action.AddDoc{
				CollectionID: 0,
				// bae-9e70648f-c722-5875-97f5-574ec6f703e9, Has written 5 books
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04, Has written 1 Book
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				// Has written no Book
				Doc: `{
					"name": "Not a Writer",
					"age": 6,
					"verified": false
				}`,
			},
			// Books
			&action.AddDoc{
				CollectionID: 1,
				// "bae-080d7580-a791-541e-90bd-49bf69f858e1", Has 1 Publisher
				DocMap: map[string]any{
					"name":      "The Rooster Bar",
					"rating":    4,
					"_authorID": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				// "bae-4e3f217c-0dd4-5ff3-bee6-5740d8fe3ae6", Has 1 Publisher
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"rating":    4.8,
					"_authorID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				// "bae-efa4a57f-e766-530f-a5a6-4a5669106c74", Has no Publisher.
				DocMap: map[string]any{
					"name":      "The Associate",
					"rating":    4.2,
					"_authorID": testUtils.NewDocIndex(0, 0),
				},
			},
			// Publishers
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":       "Only Publisher of The Rooster Bar",
					"address":    "1 Rooster Ave., Waterloo, Ontario",
					"yearOpened": 2022,
					"_bookID":    testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":       "Only Publisher of Theif Lord",
					"address":    "1 Theif Lord, Waterloo, Ontario",
					"yearOpened": 2020,
					"_bookID":    testUtils.NewDocIndex(1, 1),
				},
			},
			&action.Request{
				Request: `query {
					Book {
						name
						author {
							name
						}
						publisher {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name": "The Associate",
							"author": map[string]any{
								"name": "John Grisham",
							},
							"publisher": nil,
						},
						{
							"name": "The Rooster Bar",
							"author": map[string]any{
								"name": "Cornelia Funke",
							},
							"publisher": map[string]any{
								"name": "Only Publisher of The Rooster Bar",
							},
						},
						{
							"name": "Theif Lord",
							"author": map[string]any{
								"name": "John Grisham",
							},
							"publisher": map[string]any{
								"name": "Only Publisher of Theif Lord",
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
