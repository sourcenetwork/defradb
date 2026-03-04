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

package one_to_many

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// Note: This test partially documents:
// https://github.com/sourcenetwork/defradb/issues/2113
func TestView_OneToManyWithCount_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			&action.AddView{
				Query: `
					Author {
						name
						COUNT(books: {})
					}
				`,
				SDL: `
					type AuthorView @materialized(if: false) {
						name: String
						COUNT: Int
					}
				`,
			},
			// bae-ef9cd756-08e1-5f23-abeb-7b3e6351a68d
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Harper Lee"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name":	"To Kill a Mockingbird",
					"_authorID": "bae-ef9cd756-08e1-5f23-abeb-7b3e6351a68d"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name":	"Go Set a Watchman",
					"_authorID": "bae-ef9cd756-08e1-5f23-abeb-7b3e6351a68d"
				}`,
			},
			&action.Request{
				Request: `query {
							AuthorView {
								name
								COUNT
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
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			&action.AddView{
				Query: `
					Author {
						name
						numberOfBooks: COUNT(books: {})
					}
				`,
				SDL: `
					type AuthorView @materialized(if: false) {
						name: String
						numberOfBooks: Int
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Harper Lee"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "To Kill a Mockingbird",
					"_authorID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "Go Set a Watchman",
					"_authorID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: `
					query {
						AuthorView {
							name
							numberOfBooks
						}
					}
				`,
				Results: map[string]any{
					"AuthorView": []map[string]any{
						{
							"name":          "Harper Lee",
							"numberOfBooks": 2,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_OneToManyWithCountInQueryButNotSDL(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			&action.AddView{
				Query: `
					Author {
						name
						COUNT(books: {})
					}
				`,
				SDL: `
					type AuthorView @materialized(if: false) {
						name: String
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Harper Lee"
				}`,
			},
			&action.Request{
				Request: `
					query {
						AuthorView {
							name
						}
					}
				`,
				Results: map[string]any{
					"AuthorView": []map[string]any{
						{
							"name": "Harper Lee",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
