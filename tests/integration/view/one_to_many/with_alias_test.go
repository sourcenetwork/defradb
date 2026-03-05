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

func TestView_OneToManyWithAliasOnOuter(t *testing.T) {
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
						fullName: name
						books {
							name
						}
					}
				`,
				SDL: `
					type AuthorView @materialized(if: false) {
						fullName: String
						books: [BookView]
					}
					interface BookView {
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
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "To Kill a Mockingbird",
					"_authorID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: `query {
							AuthorView {
								fullName
								books {
									name
								}
							}
						}`,
				Results: map[string]any{
					"AuthorView": []map[string]any{
						{
							"fullName": "Harper Lee",
							"books": []map[string]any{
								{
									"name": "To Kill a Mockingbird",
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

func TestView_OneToManyWithAliasOnInner(t *testing.T) {
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
						books {
							fullName: name
						}
					}
				`,
				SDL: `
					type AuthorView @materialized(if: false) {
						name: String
						books: [BookView]
					}
					interface BookView {
						fullName: String
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
			&action.Request{
				Request: `
					query {
						AuthorView {
							name
							books {
								fullName
							}
						}
					}
				`,
				Results: map[string]any{
					"AuthorView": []map[string]any{
						{
							"name": "Harper Lee",
							"books": []map[string]any{
								{
									"fullName": "To Kill a Mockingbird",
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
