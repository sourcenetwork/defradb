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
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func gqlSchemaOneToManyToOne() testUtils.SchemaUpdate {
	return testUtils.SchemaUpdate{
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
			// bae-7aabc9d2-fbbc-5911-b0d0-b49a2a1d0e84, Has written 5 books
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
			// "bae-86f7a96a-be15-5b4d-91c7-bb6047aa4008", Has 1 Publisher
			DocMap: map[string]any{
				"name":      "Theif Lord",
				"rating":    4.8,
				"author_id": testUtils.NewDocIndex(0, 0),
			},
		},
		{
			CollectionID: 1,
			// "bae-5ce5698b-5af6-5f50-a6fb-633252be8d12", Has no Publisher.
			DocMap: map[string]any{
				"name":      "The Associate",
				"rating":    4.2,
				"author_id": testUtils.NewDocIndex(0, 0),
			},
		},
		{
			CollectionID: 1,
			// "bae-d890c705-8a7a-57ce-88b1-ddd7827438ea", Has 1 Publisher
			DocMap: map[string]any{
				"name":      "Painted House",
				"rating":    4.9,
				"author_id": testUtils.NewDocIndex(0, 0),
			},
		},
		{
			CollectionID: 1,
			// "bae-fc61b19e-646a-5537-82d6-69259e4f959a", Has 1 Publisher
			DocMap: map[string]any{
				"name":      "A Time for Mercy",
				"rating":    4.5,
				"author_id": testUtils.NewDocIndex(0, 0),
			},
		},
		{
			CollectionID: 1,
			// "bae-fc9f77fd-7b26-58c3-ad29-b2bd58a877be", Has 1 Publisher
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
