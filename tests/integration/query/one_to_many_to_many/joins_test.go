// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_many_to_many

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestOneToManyToManyJoinsAreLinkedProperly(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "1-N-M Query to ensure joins are linked properly.",
		Request: `query {
			Author {
				_key
				name
				book {
					_key
					name
					publisher {
						_key
						name
					}
				}
			}
		}`,

		Docs: map[int][]string{
			// Authors
			0: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3, Has written 5 books
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04, Has written 1 Book
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
				// Has written no Book
				`{
					"name": "Not a Writer",
					"age": 6,
					"verified": false
				}`,
			},

			// Books
			1: {
				// "bae-b6c078f2-3427-5b99-bafd-97dcd7c2e935", Has 1 Publisher
				`{
					"name": "The Rooster Bar",
					"rating": 4,
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
				}`,
				// "bae-b8091c4f-7594-5d7a-98e8-272aadcedfdf", Has 1 Publisher
				`{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				// "bae-4fb9e3e9-d1d3-5404-bf15-10e4c995d9ca", Has no Publisher.
				`{
					"name": "The Associate",
					"rating": 4.2,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				// "bae-b9b83269-1f28-5c3b-ae75-3fb4c00d559d", Has 1 Publisher
				`{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				// "bae-c674e3b0-ebb6-5b89-bfa3-d1128288d21a", Has 1 Publisher
				`{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				// "bae-7ba73251-c935-5f44-ac04-d2061149cc14", Has 2 Publishers
				`{
					"name": "Sooley",
					"rating": 3.2,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
			},

			// Publishers
			2: {
				`{
					"name": "Only Publisher of The Rooster Bar",
					"address": "1 Rooster Ave., Waterloo, Ontario",
					"yearOpened": 2022,
					"book_id": "bae-b6c078f2-3427-5b99-bafd-97dcd7c2e935"
			    }`,
				`{
					"name": "Only Publisher of Theif Lord",
					"address": "1 Theif Lord, Waterloo, Ontario",
					"yearOpened": 2020,
					"book_id": "bae-b8091c4f-7594-5d7a-98e8-272aadcedfdf"
			    }`,
				`{
					"name": "Only Publisher of Painted House",
					"address": "600 Madison Ave., New York, New York",
					"yearOpened": 1995,
					"book_id": "bae-b9b83269-1f28-5c3b-ae75-3fb4c00d559d"
			    }`,
				`{
					"name": "Only Publisher of A Time for Mercy",
					"address": "123 Andrew Street, Flin Flon, Manitoba",
					"yearOpened": 2013,
					"book_id": "bae-c674e3b0-ebb6-5b89-bfa3-d1128288d21a"
			    }`,
				`{
					"name": "First of Two Publishers of Sooley",
					"address": "11 Sooley Ave., Waterloo, Ontario",
					"yearOpened": 1999,
					"book_id": "bae-7ba73251-c935-5f44-ac04-d2061149cc14"
			    }`,
				`{
					"name": "Second of Two Publishers of Sooley",
					"address": "22 Sooley Ave., Waterloo, Ontario",
					"yearOpened": 2000,
					"book_id": "bae-7ba73251-c935-5f44-ac04-d2061149cc14"
			    }`,
			},
		},

		Results: []map[string]any{
			{
				"name": "John Grisham",
				"_key": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3",
				"book": []map[string]any{

					{
						"_key":      "bae-4fb9e3e9-d1d3-5404-bf15-10e4c995d9ca",
						"name":      "The Associate",
						"publisher": []map[string]any{},
					},

					{
						"_key": "bae-7ba73251-c935-5f44-ac04-d2061149cc14",
						"name": "Sooley",
						"publisher": []map[string]any{
							{
								"_key": "bae-cecb7841-fb4c-5403-a6d7-3654694dd073",
								"name": "First of Two Publishers of Sooley",
							},
							{
								"_key": "bae-d7e35ac3-dcf3-5537-91dd-3d27e378ba5d",
								"name": "Second of Two Publishers of Sooley",
							},
						},
					},

					{
						"_key": "bae-b8091c4f-7594-5d7a-98e8-272aadcedfdf",
						"name": "Theif Lord",
						"publisher": []map[string]any{
							{
								"_key": "bae-1a3ca715-3f3c-5934-9133-d7b489d57f88",
								"name": "Only Publisher of Theif Lord",
							},
						},
					},

					{
						"_key": "bae-b9b83269-1f28-5c3b-ae75-3fb4c00d559d",
						"name": "Painted House",
						"publisher": []map[string]any{
							{
								"_key": "bae-6412f5ff-a69a-5472-8647-18bf2b247697",
								"name": "Only Publisher of Painted House",
							},
						},
					},
					{
						"_key": "bae-c674e3b0-ebb6-5b89-bfa3-d1128288d21a",
						"name": "A Time for Mercy",
						"publisher": []map[string]any{
							{
								"_key": "bae-2f83fa75-241f-517d-9b47-3715feee43c1",
								"name": "Only Publisher of A Time for Mercy",
							},
						},
					},
				},
			},

			{
				"_key": "bae-7ba214a4-5ac8-5878-b221-dae6c285ef41",
				"book": []map[string]any{},
				"name": "Not a Writer",
			},

			{
				"name": "Cornelia Funke",
				"_key": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04",
				"book": []map[string]any{
					{
						"_key": "bae-b6c078f2-3427-5b99-bafd-97dcd7c2e935",
						"name": "The Rooster Bar",
						"publisher": []map[string]any{
							{
								"_key": "bae-3f0f19eb-b292-5e0b-b885-67e7796375f9",
								"name": "Only Publisher of The Rooster Bar",
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
