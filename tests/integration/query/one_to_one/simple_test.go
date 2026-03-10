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

package one_to_one

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToOne_PrimaryDirection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name": "Painted House",
						"rating": 4.9
					}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
						"name": "John Grisham",
						"age": 65,
						"verified": true,
						"_publishedID": "bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25"
					}`,
			},
			&action.Request{
				Request: `query {
						Book {
							name
							rating
							author {
								name
								age
							}
						}
					}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":   "Painted House",
							"rating": 4.9,
							"author": map[string]any{
								"name": "John Grisham",
								"age":  int64(65),
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToOne_SecondaryDirection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name": "Painted House",
						"rating": 4.9
					}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
						"name": "John Grisham",
						"age": 65,
						"verified": true,
						"_publishedID": "bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25"
					}`,
			},
			&action.Request{
				Request: `query {
						Author {
							name
							age
							published {
								name
								rating
							}
						}
					}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John Grisham",
							"age":  int64(65),
							"published": map[string]any{
								"name":   "Painted House",
								"rating": 4.9,
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToOneWithMultipleRecords(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						name: String
						rating: Float
						author: Author
					}

					type Author {
						name: String
						age: Int
						verified: Boolean
						published: Book @primary
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":   "Painted House",
					"rating": 4.9,
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":   "Go Guide for Rust developers",
					"rating": 5.0,
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "John Grisham",
					"age":          65,
					"verified":     true,
					"_publishedID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Andrew Lone",
					"age":          30,
					"verified":     true,
					"_publishedID": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.Request{
				Request: `query {
					Book {
						name
						author {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name": "Painted House",
							"author": map[string]any{
								"name": "John Grisham",
							},
						},
						{
							"name": "Go Guide for Rust developers",
							"author": map[string]any{
								"name": "Andrew Lone",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneWithMultipleRecordsSecondaryDirection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						name: String
						rating: Float
						author: Author
					}

					type Author {
						name: String
						age: Int
						verified: Boolean
						published: Book @primary
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "John Grisham",
					"_publishedID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Cornelia Funke",
					"_publishedID": testUtils.NewDocIndex(0, 1),
				},
			},
			&action.Request{
				Request: `query {
					Author {
						name
						published {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John Grisham",
							"published": map[string]any{
								"name": "Painted House",
							},
						},
						{
							"name": "Cornelia Funke",
							"published": map[string]any{
								"name": "Theif Lord",
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

func TestQueryOneToOneWithNilChild(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			&action.Request{
				Request: `query {
					Author {
						name
						published {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name":      "John Grisham",
							"published": nil,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToOneWithNilParent(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Painted House"
				}`,
			},
			&action.Request{
				Request: `query {
					Book {
						name
						author {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":   "Painted House",
							"author": nil,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToOne_WithRelationIDFromPrimarySide(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						name: String
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
					"name": "Painted House"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "John Grisham",
					"_publishedID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: `query {
					Author {
						name
						_publishedID
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name":         "John Grisham",
							"_publishedID": "bae-ffa6fd8c-8fd7-5da1-81d5-481bb4efd3c6",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOne_WithRelationIDFromSecondarySide(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						name: String
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
					"name": "Painted House"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "John Grisham",
					"_publishedID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: `query {
					Book {
						name
						_authorID
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":      "Painted House",
							"_authorID": "bae-e4ab9b93-bc93-52ff-8429-d7032bb914ab",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
