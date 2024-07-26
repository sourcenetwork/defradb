// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_one

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToOne(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "One-to-one relation query with no filter",
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
			Docs: map[int][]string{
				//books
				0: { // bae-be6d8024-4953-5a92-84b4-f042d25230c6
					`{
						"name": "Painted House",
						"rating": 4.9
					}`,
				},
				//authors
				1: { // bae-7aabc9d2-fbbc-5911-b0d0-b49a2a1d0e84
					`{
						"name": "John Grisham",
						"age": 65,
						"verified": true,
						"published_id": "bae-be6d8024-4953-5a92-84b4-f042d25230c6"
					}`,
				},
			},
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
		{
			Description: "One-to-one relation secondary direction, no filter",
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
			Docs: map[int][]string{
				//books
				0: { // bae-be6d8024-4953-5a92-84b4-f042d25230c6
					`{
						"name": "Painted House",
						"rating": 4.9
					}`,
				},
				//authors
				1: { // bae-7aabc9d2-fbbc-5911-b0d0-b49a2a1d0e84
					`{
						"name": "John Grisham",
						"age": 65,
						"verified": true,
						"published_id": "bae-be6d8024-4953-5a92-84b4-f042d25230c6"
					}`,
				},
			},
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
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQueryOneToOneWithMultipleRecords(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation primary direction, multiple records",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
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
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":   "Painted House",
					"rating": 4.9,
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":   "Go Guide for Rust developers",
					"rating": 5.0,
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "John Grisham",
					"age":          65,
					"verified":     true,
					"published_id": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Andrew Lone",
					"age":          30,
					"verified":     true,
					"published_id": testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.Request{
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
							"name": "Go Guide for Rust developers",
							"author": map[string]any{
								"name": "Andrew Lone",
							},
						},
						{
							"name": "Painted House",
							"author": map[string]any{
								"name": "John Grisham",
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
		Description: "One-to-one-to-one relation secondary direction",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
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
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "John Grisham",
					"published_id": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Cornelia Funke",
					"published_id": testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.Request{
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
							"name": "Cornelia Funke",
							"published": map[string]any{
								"name": "Theif Lord",
							},
						},
						{
							"name": "John Grisham",
							"published": map[string]any{
								"name": "Painted House",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneWithNilChild(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-one relation primary direction, nil child",
		Request: `query {
			Author {
				name
				published {
					name
				}
			}
		}`,
		Docs: map[int][]string{
			//authors
			1: {
				`{
					"name": "John Grisham"
				}`,
			},
		},
		Results: map[string]any{
			"Author": []map[string]any{
				{
					"name":      "John Grisham",
					"published": nil,
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToOneWithNilParent(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-one relation primary direction, nil parent",
		Request: `query {
			Book {
				name
				author {
					name
				}
			}
		}`,
		Docs: map[int][]string{
			//books
			0: {
				`{
					"name": "Painted House"
				}`,
			},
		},
		Results: map[string]any{
			"Book": []map[string]any{
				{
					"name":   "Painted House",
					"author": nil,
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToOne_WithRelationIDFromPrimarySide(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation primary direction, relation ID field",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
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
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "John Grisham",
					"published_id": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.Request{
				Request: `query {
					Author {
						name
						published_id
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name":         "John Grisham",
							"published_id": "bae-514f04b1-b218-5b8c-89ee-538f150a32b5",
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
		Description: "One-to-one relation secondary direction, relation ID field",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
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
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "John Grisham",
					"published_id": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.Request{
				Request: `query {
					Book {
						name
						author_id
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":      "Painted House",
							"author_id": "bae-420e72a6-e0c6-5a06-a958-2cc7adb7b3d0",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
