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
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// Note: This test partially documents:
// https://github.com/sourcenetwork/defradb/issues/2113
func TestView_OneToManyWithCount_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to many view with count",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Author {
						name: String
						books: [Book]
					}
					type Book {
						name: String
						author: Author
					}
				`,
			},
			testUtils.CreateView{
				Query: `
					Author {
						name
						_count(books: {})
					}
				`,
				SDL: `
					type AuthorView {
						name: String
						_count: Int
					}
				`,
			},
			// bae-ef9cd756-08e1-5f23-abeb-7b3e6351a68d
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Harper Lee"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name":	"To Kill a Mockingbird",
					"author_id": "bae-ef9cd756-08e1-5f23-abeb-7b3e6351a68d"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name":	"Go Set a Watchman",
					"author_id": "bae-ef9cd756-08e1-5f23-abeb-7b3e6351a68d"
				}`,
			},
			testUtils.Request{
				Request: `query {
							AuthorView {
								name
								_count
							}
						}`,
				ExpectedError: "aggregate must be provided with a property to aggregate",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_OneToManyWithAliasedCount(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to many view with aliased count",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Author {
						name: String
						books: [Book]
					}
					type Book {
						name: String
						author: Author
					}
				`,
			},
			testUtils.CreateView{
				Query: `
					Author {
						name
						numberOfBooks: _count(books: {})
					}
				`,
				SDL: `
					type AuthorView {
						name: String
						numberOfBooks: Int
					}
				`,
			},
			// bae-ef9cd756-08e1-5f23-abeb-7b3e6351a68d
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Harper Lee"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name":	"To Kill a Mockingbird",
					"author_id": "bae-ef9cd756-08e1-5f23-abeb-7b3e6351a68d"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name":	"Go Set a Watchman",
					"author_id": "bae-ef9cd756-08e1-5f23-abeb-7b3e6351a68d"
				}`,
			},
			testUtils.Request{
				Request: `query {
							AuthorView {
								name
								numberOfBooks
							}
						}`,
				Results: []map[string]any{
					{
						"name":          "Harper Lee",
						"numberOfBooks": 2,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
