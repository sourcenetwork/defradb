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
			// bae-41598f0c-19bc-5da6-813b-e80f14a10df3, Has written 5 books
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
			// "bae-b6c078f2-3427-5b99-bafd-97dcd7c2e935", Has 1 Publisher
			Doc: `{
					"name": "The Rooster Bar",
					"rating": 4,
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				}`,
		},
		{
			CollectionID: 1,
			// "bae-b8091c4f-7594-5d7a-98e8-272aadcedfdf", Has 1 Publisher
			Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
		},
		{
			CollectionID: 1,
			// "bae-4fb9e3e9-d1d3-5404-bf15-10e4c995d9ca", Has no Publisher.
			Doc: `{
					"name": "The Associate",
					"rating": 4.2,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
		},
		{
			CollectionID: 1,
			// "bae-b9b83269-1f28-5c3b-ae75-3fb4c00d559d", Has 1 Publisher
			Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
		},
		{
			CollectionID: 1,
			// "bae-c674e3b0-ebb6-5b89-bfa3-d1128288d21a", Has 1 Publisher
			Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
		},
		{
			CollectionID: 1,
			// "bae-7ba73251-c935-5f44-ac04-d2061149cc14", Has 1 Publisher
			Doc: `{
					"name": "Sooley",
					"rating": 3.2,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
		},
		// Publishers
		{
			CollectionID: 2,
			Doc: `{
					"name": "Only Publisher of The Rooster Bar",
					"address": "1 Rooster Ave., Waterloo, Ontario",
					"yearOpened": 2022,
					"book_id": "bae-b6c078f2-3427-5b99-bafd-97dcd7c2e935"
			    }`,
		},
		{
			CollectionID: 2,
			Doc: `{
					"name": "Only Publisher of Theif Lord",
					"address": "1 Theif Lord, Waterloo, Ontario",
					"yearOpened": 2020,
					"book_id": "bae-b8091c4f-7594-5d7a-98e8-272aadcedfdf"
			    }`,
		},
		{
			CollectionID: 2,
			Doc: `{
					"name": "Only Publisher of Painted House",
					"address": "600 Madison Ave., New York, New York",
					"yearOpened": 1995,
					"book_id": "bae-b9b83269-1f28-5c3b-ae75-3fb4c00d559d"
			    }`,
		},
		{
			CollectionID: 2,
			Doc: `{
					"name": "Only Publisher of A Time for Mercy",
					"address": "123 Andrew Street, Flin Flon, Manitoba",
					"yearOpened": 2013,
					"book_id": "bae-c674e3b0-ebb6-5b89-bfa3-d1128288d21a"
			    }`,
		},
		{
			CollectionID: 2,
			Doc: `{
					"name": "Only Publisher of Sooley",
					"address": "11 Sooley Ave., Waterloo, Ontario",
					"yearOpened": 1999,
					"book_id": "bae-7ba73251-c935-5f44-ac04-d2061149cc14"
			    }`,
		},
	}
}
