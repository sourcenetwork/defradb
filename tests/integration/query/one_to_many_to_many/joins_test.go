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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestOneToManyToManyJoinsAreLinkedProperly(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Author {
						name: String
						age: Int
						verified: Boolean
						book: [Book]
					}

					type Book {
						name: String
						rating: Float
						author: Author
						publisher: [Publisher]
					}

					type Publisher {
						name: String
						address: String
						yearOpened: Int
						book: Book
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Not a Writer",
					"age": 6,
					"verified": false
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "The Rooster Bar",
					"rating":    4,
					"_authorID": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"rating":    4.8,
					"_authorID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "The Associate",
					"rating":    4.2,
					"_authorID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "Painted House",
					"rating":    4.9,
					"_authorID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"rating":    4.5,
					"_authorID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "Sooley",
					"rating":    3.2,
					"_authorID": testUtils.NewDocIndex(0, 0),
				},
			},
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
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":       "Only Publisher of Painted House",
					"address":    "600 Madison Ave., New York, New York",
					"yearOpened": 1995,
					"_bookID":    testUtils.NewDocIndex(1, 3),
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":       "Only Publisher of A Time for Mercy",
					"address":    "123 Andrew Street, Flin Flon, Manitoba",
					"yearOpened": 2013,
					"_bookID":    testUtils.NewDocIndex(1, 4),
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":       "First of Two Publishers of Sooley",
					"address":    "11 Sooley Ave., Waterloo, Ontario",
					"yearOpened": 1999,
					"_bookID":    testUtils.NewDocIndex(1, 5),
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":       "Second of Two Publishers of Sooley",
					"address":    "22 Sooley Ave., Waterloo, Ontario",
					"yearOpened": 2000,
					"_bookID":    testUtils.NewDocIndex(1, 5),
				},
			},
			&action.Request{
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
							"_docID": testUtils.NewDocIndex(0, 2),
							"book":   []map[string]any{},
							"name":   "Not a Writer",
						},
						{
							"name":   "John Grisham",
							"_docID": testUtils.NewDocIndex(0, 0),
							"book": []map[string]any{
								{
									"_docID":    testUtils.NewDocIndex(1, 2),
									"name":      "The Associate",
									"publisher": []map[string]any{},
								},
								{
									"_docID": testUtils.NewDocIndex(1, 3),
									"name":   "Painted House",
									"publisher": []map[string]any{
										{
											"_docID": testUtils.NewDocIndex(2, 2),
											"name":   "Only Publisher of Painted House",
										},
									},
								},
								{
									"_docID": testUtils.NewDocIndex(1, 1),
									"name":   "Theif Lord",
									"publisher": []map[string]any{
										{
											"_docID": testUtils.NewDocIndex(2, 1),
											"name":   "Only Publisher of Theif Lord",
										},
									},
								},
								{
									"_docID": testUtils.NewDocIndex(1, 4),
									"name":   "A Time for Mercy",
									"publisher": []map[string]any{
										{
											"_docID": testUtils.NewDocIndex(2, 3),
											"name":   "Only Publisher of A Time for Mercy",
										},
									},
								},
								{
									"_docID": testUtils.NewDocIndex(1, 5),
									"name":   "Sooley",
									"publisher": []map[string]any{
										{
											"_docID": testUtils.NewDocIndex(2, 4),
											"name":   "First of Two Publishers of Sooley",
										},
										{
											"_docID": testUtils.NewDocIndex(2, 5),
											"name":   "Second of Two Publishers of Sooley",
										},
									},
								},
							},
						},
						{
							"name":   "Cornelia Funke",
							"_docID": testUtils.NewDocIndex(0, 1),
							"book": []map[string]any{
								{
									"_docID": testUtils.NewDocIndex(1, 0),
									"name":   "The Rooster Bar",
									"publisher": []map[string]any{
										{
											"_docID": testUtils.NewDocIndex(2, 0),
											"name":   "Only Publisher of The Rooster Bar",
										},
									},
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
