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
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestOneToManyToOneJoinsAreLinkedProperly(t *testing.T) {
	test := testUtils.TestCase{
		Description: "1-N-1 Query to ensure joins are linked properly.",
		Actions: []any{
			gqlSchemaOneToManyToOne(),
			// Authors
			testUtils.CreateDoc{
				CollectionID: 0,
				// Has written 5 books
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// Has written 1 Book
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// Has written no Book
				Doc: `{
					"name": "Not a Writer",
					"age": 6,
					"verified": false
				}`,
			},
			// Books
			testUtils.CreateDoc{
				CollectionID: 1,
				// Has 1 Publisher
				DocMap: map[string]any{
					"name":      "The Rooster Bar",
					"rating":    4,
					"author_id": testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// Has 1 Publisher
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"rating":    4.8,
					"author_id": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// Has no Publisher.
				DocMap: map[string]any{
					"name":      "The Associate",
					"rating":    4.2,
					"author_id": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// Has 1 Publisher
				DocMap: map[string]any{
					"name":      "Painted House",
					"rating":    4.9,
					"author_id": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// Has 1 Publisher
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"rating":    4.5,
					"author_id": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// Has 1 Publisher
				DocMap: map[string]any{
					"name":      "Sooley",
					"rating":    3.2,
					"author_id": testUtils.NewDocIndex(0, 0),
				},
			},
			// Publishers
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":       "Only Publisher of The Rooster Bar",
					"address":    "1 Rooster Ave., Waterloo, Ontario",
					"yearOpened": 2022,
					"book_id":    testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":       "Only Publisher of Theif Lord",
					"address":    "1 Theif Lord, Waterloo, Ontario",
					"yearOpened": 2020,
					"book_id":    testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":       "Only Publisher of Painted House",
					"address":    "600 Madison Ave., New York, New York",
					"yearOpened": 1995,
					"book_id":    testUtils.NewDocIndex(1, 3),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":       "Only Publisher of A Time for Mercy",
					"address":    "123 Andrew Street, Flin Flon, Manitoba",
					"yearOpened": 2013,
					"book_id":    testUtils.NewDocIndex(1, 4),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":       "Only Publisher of Sooley",
					"address":    "11 Sooley Ave., Waterloo, Ontario",
					"yearOpened": 1999,
					"book_id":    testUtils.NewDocIndex(1, 5),
				},
			},
			testUtils.Request{
				Request: `query {
					Author {
						_docID
						name
						book {
							_docID
							name
							publisher {
								_docID
								name
							}
						}
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"_docID": "bae-489b4e01-4764-56f6-913f-b3c92dffcaa3",
							"book":   []map[string]any{},
							"name":   "Not a Writer",
						},
						{
							"name":   "John Grisham",
							"_docID": "bae-7aabc9d2-fbbc-5911-b0d0-b49a2a1d0e84",
							"book": []map[string]any{
								{
									"_docID":    "bae-5ce5698b-5af6-5f50-a6fb-633252be8d12",
									"name":      "The Associate",
									"publisher": nil,
								},
								{
									"_docID": "bae-86f7a96a-be15-5b4d-91c7-bb6047aa4008",
									"name":   "Theif Lord",
									"publisher": map[string]any{
										"_docID": "bae-6223fba1-5461-5e47-9682-6c769c8e5518",
										"name":   "Only Publisher of Theif Lord",
									},
								},
								{
									"_docID": "bae-d890c705-8a7a-57ce-88b1-ddd7827438ea",
									"name":   "Painted House",
									"publisher": map[string]any{
										"_docID": "bae-de7d087b-d33f-5b4b-b0e4-79de4335d9ed",
										"name":   "Only Publisher of Painted House",
									},
								},
								{
									"_docID": "bae-fc61b19e-646a-5537-82d6-69259e4f959a",
									"name":   "A Time for Mercy",
									"publisher": map[string]any{
										"_docID": "bae-5fd29915-86c6-5e9f-863a-a03292206b8c",
										"name":   "Only Publisher of A Time for Mercy",
									},
								},
								{
									"_docID": "bae-fc9f77fd-7b26-58c3-ad29-b2bd58a877be",
									"name":   "Sooley",
									"publisher": map[string]any{
										"_docID": "bae-e2cc19bd-4b3e-5cbe-9146-fb24f5913566",
										"name":   "Only Publisher of Sooley",
									},
								},
							},
						},
						{
							"name":   "Cornelia Funke",
							"_docID": "bae-fb2a1852-3951-5ce9-a3bf-6825202f201b",
							"book": []map[string]any{
								{
									"_docID": "bae-5a5ef6dd-0c2b-5cd0-a644-f0c47a640565",
									"name":   "The Rooster Bar",
									"publisher": map[string]any{
										"_docID": "bae-0020b43b-500c-57d0-81b3-43342c9d8d1d",
										"name":   "Only Publisher of The Rooster Bar",
									},
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
