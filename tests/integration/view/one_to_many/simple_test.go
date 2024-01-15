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

func TestView_OneToMany(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to many view",
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
							name
						}
					}
				`,
				SDL: `
					type AuthorView {
						name: String
						books: [BookView]
					}
					interface BookView {
						name: String
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
			testUtils.Request{
				Request: `query {
							AuthorView {
								name
								books {
									name
								}
							}
						}`,
				Results: []map[string]any{
					{
						"name": "Harper Lee",
						"books": []map[string]any{
							{
								"name": "To Kill a Mockingbird",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_OneToManyWithMixedSDL_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to many view with mixed sdl errors",
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
							name
						}
					}
				`,
				SDL: `
					type AuthorView {
						name: String
						books: [Book]
					}
				`,
				ExpectedError: "relation must be defined on both schemas. Field: books, Type: Book",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_OneToManyFromInnerSide_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to many view from inner side",
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
							name
						}
					}
				`,
				SDL: `
					type AuthorView {
						name: String
						books: [BookView]
					}
					interface BookView {
						name: String
					}
				`,
			},
			testUtils.Request{
				Request: `query {
							BookView {
								name
								author {
									name
								}
							}
						}`,
				ExpectedError: `Cannot query field "BookView" on type "Query".`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_OneToManyOuterToInnerToOuter_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to many view from outer to inner to outer",
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
							name
							author {
								name
							}
						}
					}
				`,
				SDL: `
					type AuthorView {
						name: String
						books: [BookView]
					}
					interface BookView {
						name: String
					}
				`,
			},
			testUtils.Request{
				Request: `query {
							AuthorView {
								name
								books {
									name
									author {
										name
									}
								}
							}
						}`,
				ExpectedError: `Cannot query field "author" on type "BookView".`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_OneToManyWithRelationInQueryButNotInSDL(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to many view with relation in query but not SDL",
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
				// Query books via author but do not declare relation in SDL
				Query: `
					Author {
						name
						books {
							name
						}
					}
				`,
				SDL: `
					type AuthorView {
						name: String
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
			testUtils.Request{
				Request: `query {
							AuthorView {
								name
							}
						}`,
				Results: []map[string]any{
					{
						"name": "Harper Lee",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_OneToManyMultipleViewsWithEmbeddedSchema(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Multiple one to many views with embedded schemas",
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
					Book {
						name
						author {
							name
						}
					}
				`,
				SDL: `
					type BookView {
						name: String
						author: AuthorView
					}
					interface AuthorView {
						name: String
					}
				`,
			},
			testUtils.CreateView{
				Query: `
					Book {
						name
						author {
							name
						}
					}
				`,
				SDL: `
					type BookView2 {
						name: String
						author: AuthorView2
					}
					interface AuthorView2 {
						name: String
					}
				`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_OneToManyWithDoubleSidedRelation_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to many view",
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
							name
						}
					}
				`,
				SDL: `
					type AuthorView {
						name: String
						books: [BookView]
					}
					interface BookView {
						name: String
					}
				`,
			},
			testUtils.CreateView{
				Query: `
					AuthorView {
						name
						books {
							name
						}
					}
				`,
				SDL: `
					type AuthorViewView {
						name: String
						books: [BookViewView]
					}
					interface BookViewView {
						name: String
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
			testUtils.Request{
				Request: `query {
							AuthorViewView {
								name
								books {
									name
								}
							}
						}`,
				Results: []map[string]any{
					{
						"name": "Harper Lee",
						"books": []map[string]any{
							{
								"name": "To Kill a Mockingbird",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_OneToManyViewOfView(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to many view of view",
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
							name
						}
					}
				`,
				SDL: `
					type AuthorView {
						name: String
						books: [BookView]
					}
					interface BookView {
						name: String
						author: AuthorView
					}
				`,
				ExpectedError: "relations in views must only be defined on one schema",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
