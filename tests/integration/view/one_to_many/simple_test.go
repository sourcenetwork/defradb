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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestView_OneToMany(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
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
			&action.AddView{
				Query: `
					Author {
						name
						books {
							name
						}
					}
				`,
				SDL: `
					type AuthorView @materialized(if: false) {
						name: String
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
								name
								books {
									name
								}
							}
						}`,
				Results: map[string]any{
					"AuthorView": []map[string]any{
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_OneToManyWithMixedSDL(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
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
			&action.AddView{
				Query: `
					Author {
						name
						books {
							name
						}
					}
				`,
				SDL: `
					type AuthorView @materialized(if: false) {
						name: String
						books: [Book]
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
								name
								books {
									name
								}
							}
						}`,
				Results: map[string]any{
					"AuthorView": []map[string]any{
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestView_OneToManyFromInnerSide_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
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
			&action.AddView{
				Query: `
					Author {
						name
						books {
							name
						}
					}
				`,
				SDL: `
					type AuthorView @materialized(if: false) {
						name: String
						books: [BookView]
					}
					interface BookView {
						name: String
					}
				`,
			},
			&action.Request{
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
		Actions: []any{
			&action.AddSchema{
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
			&action.AddView{
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
					type AuthorView @materialized(if: false) {
						name: String
						books: [BookView]
					}
					interface BookView {
						name: String
					}
				`,
			},
			&action.Request{
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
		Actions: []any{
			&action.AddSchema{
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
			&action.AddView{
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
					type AuthorView @materialized(if: false) {
						name: String
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
			&action.Request{
				Request: `query {
							AuthorView {
								name
							}
						}`,
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

func TestView_OneToManyMultipleViewsWithEmbeddedSchema(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
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
			&action.AddView{
				Query: `
					Book {
						name
						author {
							name
						}
					}
				`,
				SDL: `
					type BookView @materialized(if: false) {
						name: String
						author: AuthorView
					}
					interface AuthorView {
						name: String
					}
				`,
			},
			&action.AddView{
				Query: `
					Book {
						name
						author {
							name
						}
					}
				`,
				SDL: `
					type BookView2 @materialized(if: false) {
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
		Actions: []any{
			&action.AddSchema{
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
			&action.AddView{
				Query: `
					Author {
						name
						books {
							name
						}
					}
				`,
				SDL: `
					type AuthorView @materialized(if: false) {
						name: String
						books: [BookView]
					}
					interface BookView {
						name: String
					}
				`,
			},
			&action.AddView{
				Query: `
					AuthorView {
						name
						books {
							name
						}
					}
				`,
				SDL: `
					type AuthorViewView @materialized(if: false) {
						name: String
						books: [BookViewView]
					}
					interface BookViewView {
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
							AuthorViewView {
								name
								books {
									name
								}
							}
						}`,
				Results: map[string]any{
					"AuthorViewView": []map[string]any{
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
