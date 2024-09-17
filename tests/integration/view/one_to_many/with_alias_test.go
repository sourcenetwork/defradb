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

func TestView_OneToManyWithAliasOnOuter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to many view with alias on outer object",
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
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Harper Lee"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "To Kill a Mockingbird",
					"author_id": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.Request{
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
		Description: "One to many view with alias on inner object",
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
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Harper Lee"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "To Kill a Mockingbird",
					"author_id": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.Request{
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
