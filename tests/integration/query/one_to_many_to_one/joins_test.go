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
									"_docID": "bae-1d0dcbed-300a-567a-9b48-c23cd026d165",
									"name":   "A Time for Mercy",
									"publisher": map[string]any{
										"_docID": "bae-2bad7de3-0f1a-56c0-b499-a552debef4b8",
										"name":   "Only Publisher of A Time for Mercy",
									},
								},
								{
									"_docID":    "bae-374998e0-e84d-5f6b-9e87-5edaaa2d9c7d",
									"name":      "The Associate",
									"publisher": nil,
								},
								{
									"_docID": "bae-7697f14d-7b32-5884-8677-344e183c14bf",
									"name":   "Theif Lord",
									"publisher": map[string]any{
										"_docID": "bae-d43823c0-0bb6-58a9-a098-1826dffa4e4a",
										"name":   "Only Publisher of Theif Lord",
									},
								},
								{
									"_docID": "bae-aef1d940-5ac1-5924-a87f-63ac40758b22",
									"name":   "Painted House",
									"publisher": map[string]any{
										"_docID": "bae-a104397b-7804-5cd0-93e5-c3986b4e5e71",
										"name":   "Only Publisher of Painted House",
									},
								},
								{
									"_docID": "bae-ee6b8339-8a9e-58a9-9a0d-dbd8d44fa149",
									"name":   "Sooley",
									"publisher": map[string]any{
										"_docID": "bae-efeca601-cce1-5289-b392-85fa5b7bc0f7",
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
									"_docID": "bae-1867d7cb-01b3-572f-a993-1c3f22f46526",
									"name":   "The Rooster Bar",
									"publisher": map[string]any{
										"_docID": "bae-09af7e39-8596-584f-8825-cb430c4156b3",
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
