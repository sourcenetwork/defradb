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

package one_to_one_to_one

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToOneToOne(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Publisher {
						name: String
						printed: Book @primary
					}

					type Book {
						name: String
						publisher: Publisher
						author: Author @primary
					}

					type Author {
						name: String
						published: Book
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name": "John Grisham",
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name": "Cornelia Funke",
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "Painted House",
					"_authorID": testUtils.NewDocIndex(2, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"_authorID": testUtils.NewDocIndex(2, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":       "Old Publisher",
					"_printedID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":       "New Publisher",
					"_printedID": testUtils.NewDocIndex(1, 1),
				},
			},
			&action.Request{
				Request: `query {
					Publisher {
						name
						printed {
							name
							author {
								name
							}
						}
					}
				}`,
				Results: map[string]any{
					"Publisher": []map[string]any{
						{
							"name": "New Publisher",
							"printed": map[string]any{
								"name": "Theif Lord",
								"author": map[string]any{
									"name": "Cornelia Funke",
								},
							},
						},
						{
							"name": "Old Publisher",
							"printed": map[string]any{
								"name": "Painted House",
								"author": map[string]any{
									"name": "John Grisham",
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

func TestQueryOneToOneToOneSecondaryThenPrimary(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Publisher {
						name: String
						printed: Book
					}

					type Book {
						name: String
						publisher: Publisher @primary
						author: Author @primary
					}

					type Author {
						name: String
						published: Book
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Old Publisher"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "New Publisher"
				}`,
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name": "John Grisham",
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name": "Cornelia Funke",
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Painted House",
					"_publisherID": testUtils.NewDocIndex(0, 0),
					"_authorID":    testUtils.NewDocIndex(2, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Theif Lord",
					"_publisherID": testUtils.NewDocIndex(0, 1),
					"_authorID":    testUtils.NewDocIndex(2, 1),
				},
			},
			&action.Request{
				Request: `query {
					Publisher {
						name
						printed {
							name
							author {
								name
							}
						}
					}
				}`,
				Results: map[string]any{
					"Publisher": []map[string]any{
						{
							"name": "New Publisher",
							"printed": map[string]any{
								"name": "Theif Lord",
								"author": map[string]any{
									"name": "Cornelia Funke",
								},
							},
						},
						{
							"name": "Old Publisher",
							"printed": map[string]any{
								"name": "Painted House",
								"author": map[string]any{
									"name": "John Grisham",
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

func TestQueryOneToOneToOnePrimaryThenSecondary(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Publisher {
						name: String
						printed: Book @primary
					}

					type Book {
						name: String
						publisher: Publisher
						author: Author
					}

					type Author {
						name: String
						published: Book @primary
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name": "Painted House",
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name": "Theif Lord",
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":       "Old Publisher",
					"_printedID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":       "New Publisher",
					"_printedID": testUtils.NewDocIndex(1, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":         "John Grisham",
					"_publishedID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":         "Cornelia Funke",
					"_publishedID": testUtils.NewDocIndex(1, 1),
				},
			},
			&action.Request{
				Request: `query {
					Publisher {
						name
						printed {
							name
							author {
								name
							}
						}
					}
				}`,
				Results: map[string]any{
					"Publisher": []map[string]any{
						{
							"name": "Old Publisher",
							"printed": map[string]any{
								"name": "Painted House",
								"author": map[string]any{
									"name": "John Grisham",
								},
							},
						},
						{
							"name": "New Publisher",
							"printed": map[string]any{
								"name": "Theif Lord",
								"author": map[string]any{
									"name": "Cornelia Funke",
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

func TestQueryOneToOneToOneSecondary(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Publisher {
						name: String
						printed: Book
					}

					type Book {
						name: String
						publisher: Publisher  @primary
						author: Author
					}

					type Author {
						name: String
						published: Book @primary
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Old Publisher"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "New Publisher"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Painted House",
					"_publisherID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Theif Lord",
					"_publisherID": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":         "John Grisham",
					"_publishedID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":         "Cornelia Funke",
					"_publishedID": testUtils.NewDocIndex(1, 1),
				},
			},
			&action.Request{
				Request: `query {
					Publisher {
						name
						printed {
							name
							author {
								name
							}
						}
					}
				}`,
				Results: map[string]any{
					"Publisher": []map[string]any{
						{
							"name": "New Publisher",
							"printed": map[string]any{
								"name": "Theif Lord",
								"author": map[string]any{
									"name": "Cornelia Funke",
								},
							},
						},
						{
							"name": "Old Publisher",
							"printed": map[string]any{
								"name": "Painted House",
								"author": map[string]any{
									"name": "John Grisham",
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
