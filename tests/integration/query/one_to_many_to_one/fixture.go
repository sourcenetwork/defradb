// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_many_to_one

import (
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func gqlSchemaOneToManyToOne() *action.AddSchema {
	return &action.AddSchema{
		Schema: (`
			type Author {
				name: String
				age: Int
				verified: Boolean
				favouritePageNumbers: [Int!]
				book: [Book]
			}

			type Book {
				name: String
				rating: Float
				author: Author
				publisher: Publisher
			}

			type Publisher {
				name: String
				address: String
				yearOpened: Int
				book: Book @primary
			}
		`),
	}
}

func createDocsWith6BooksAnd5Publishers() []testUtils.CreateDoc {
	return []testUtils.CreateDoc{
		// Authors
		{
			CollectionID: 0,
			// bae-9e70648f-c722-5875-97f5-574ec6f703e9, Has written 5 books
			Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
		},
		{
			CollectionID: 0,
			// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04, Has written 1 Book
			Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
		},
		{
			CollectionID: 0,
			// Has written no Book
			Doc: `{
					"name": "Not a Writer",
					"age": 6,
					"verified": false
				}`,
		},
		// Books
		{
			CollectionID: 1,
			// "bae-080d7580-a791-541e-90bd-49bf69f858e1", Has 1 Publisher
			DocMap: map[string]any{
				"name":      "The Rooster Bar",
				"rating":    4,
				"author_id": testUtils.NewDocIndex(0, 1),
			},
		},
		{
			CollectionID: 1,
			// "bae-4e3f217c-0dd4-5ff3-bee6-5740d8fe3ae6", Has 1 Publisher
			DocMap: map[string]any{
				"name":      "Theif Lord",
				"rating":    4.8,
				"author_id": testUtils.NewDocIndex(0, 0),
			},
		},
		{
			CollectionID: 1,
			// "bae-efa4a57f-e766-530f-a5a6-4a5669106c74", Has no Publisher.
			DocMap: map[string]any{
				"name":      "The Associate",
				"rating":    4.2,
				"author_id": testUtils.NewDocIndex(0, 0),
			},
		},
		{
			CollectionID: 1,
			// "bae-622eeb8a-2c78-552d-91ed-f6ad86bdbb0a", Has 1 Publisher
			DocMap: map[string]any{
				"name":      "Painted House",
				"rating":    4.9,
				"author_id": testUtils.NewDocIndex(0, 0),
			},
		},
		{
			CollectionID: 1,
			// "bae-ed3a12b5-a362-5278-a26e-b3c4595463f6", Has 1 Publisher
			DocMap: map[string]any{
				"name":      "A Time for Mercy",
				"rating":    4.5,
				"author_id": testUtils.NewDocIndex(0, 0),
			},
		},
		{
			CollectionID: 1,
			// "bae-f624d6f8-957f-57fd-89ed-76b22bf092b9", Has 1 Publisher
			DocMap: map[string]any{
				"name":      "Sooley",
				"rating":    3.2,
				"author_id": testUtils.NewDocIndex(0, 0),
			},
		},
		// Publishers
		{
			CollectionID: 2,
			DocMap: map[string]any{
				"name":       "Only Publisher of The Rooster Bar",
				"address":    "1 Rooster Ave., Waterloo, Ontario",
				"yearOpened": 2022,
				"book_id":    testUtils.NewDocIndex(1, 0),
			},
		},
		{
			CollectionID: 2,
			DocMap: map[string]any{
				"name":       "Only Publisher of Theif Lord",
				"address":    "1 Theif Lord, Waterloo, Ontario",
				"yearOpened": 2020,
				"book_id":    testUtils.NewDocIndex(1, 1),
			},
		},
		{
			CollectionID: 2,
			DocMap: map[string]any{
				"name":       "Only Publisher of Painted House",
				"address":    "600 Madison Ave., New York, New York",
				"yearOpened": 1995,
				"book_id":    testUtils.NewDocIndex(1, 3),
			},
		},
		{
			CollectionID: 2,
			DocMap: map[string]any{
				"name":       "Only Publisher of A Time for Mercy",
				"address":    "123 Andrew Street, Flin Flon, Manitoba",
				"yearOpened": 2013,
				"book_id":    testUtils.NewDocIndex(1, 4),
			},
		},
		{
			CollectionID: 2,
			DocMap: map[string]any{
				"name":       "Only Publisher of Sooley",
				"address":    "11 Sooley Ave., Waterloo, Ontario",
				"yearOpened": 1999,
				"book_id":    testUtils.NewDocIndex(1, 5),
			},
		},
	}
}
