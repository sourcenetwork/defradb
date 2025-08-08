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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToOne(t *testing.T) {
	tests := []testUtils.TestCase{
		{
			Description: "One-to-one relation query with no filter",
			Actions: []any{
				testUtils.CreateDoc{
					CollectionID: 0,
					Doc: `{
						"name": "Painted House",
						"rating": 4.9
					}`,
				},
				testUtils.CreateDoc{
					CollectionID: 1,
					Doc: `{
						"name": "John Grisham",
						"age": 65,
						"verified": true,
						"published_id": "bae-818aecea-02f9-5064-9e17-c8b7cc20e63f"
					}`,
				},
				testUtils.Request{
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
		},
		{
			Description: "One-to-one relation secondary direction, no filter",
			Actions: []any{
				testUtils.CreateDoc{
					CollectionID: 0,
					Doc: `{
						"name": "Painted House",
						"rating": 4.9
					}`,
				},
				testUtils.CreateDoc{
					CollectionID: 1,
					Doc: `{
						"name": "John Grisham",
						"age": 65,
						"verified": true,
						"published_id": "bae-818aecea-02f9-5064-9e17-c8b7cc20e63f"
					}`,
				},
				testUtils.Request{
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
			&action.AddSchema{
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
		Description: "One-to-one-to-one relation secondary direction",
		Actions: []any{
			&action.AddSchema{
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneWithNilChild(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation primary direction, nil child",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham"
				}`,
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
		Description: "One-to-one relation primary direction, nil parent",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "Painted House"
				}`,
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
		Description: "One-to-one relation primary direction, relation ID field",
		Actions: []any{
			&action.AddSchema{
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
							"published_id": "bae-131c8f1b-2f61-599e-824f-afc313812c58",
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
			&action.AddSchema{
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
							"author_id": "bae-0362a1da-4a34-5c53-97a3-f5bdcea5d78f",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
