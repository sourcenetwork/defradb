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
				0: { // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
					`{
						"name": "Painted House",
						"rating": 4.9
					}`,
				},
				//authors
				1: { // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
					`{
						"name": "John Grisham",
						"age": 65,
						"verified": true,
						"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
					}`,
				},
			},
			Results: []map[string]any{
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
				0: { // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
					`{
						"name": "Painted House",
						"rating": 4.9
					}`,
				},
				//authors
				1: { // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
					`{
						"name": "John Grisham",
						"age": 65,
						"verified": true,
						"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
					}`,
				},
			},
			Results: []map[string]any{
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
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQueryOneToOneWithMultipleRecords(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-one relation primary direction, multiple records",
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
				// bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				`{
					"name": "Painted House",
					"rating": 4.9
				}`,
				// "bae-d3bc0f38-a2e1-5a26-9cc9-5b3fdb41c6db"
				`{
					"name": "Go Guide for Rust developers",
					"rating": 5.0
				}`,
			},
			//authors
			1: {
				// "bae-3bfe0092-e31f-5ebe-a3ba-fa18fac448a6"
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
				}`,
				// "bae-756c2bf0-4767-57fd-b12b-393915feae68",
				`{
					"name": "Andrew Lone",
					"age": 30,
					"verified": true,
					"published_id": "bae-d3bc0f38-a2e1-5a26-9cc9-5b3fdb41c6db"
				}`,
			},
		},
		Results: []map[string]any{
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
	}

	executeTestCase(t, test)
}

func TestQueryOneToOneWithMultipleRecordsSecondaryDirection(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-one-to-one relation secondary direction",
		Request: `query {
			Author {
				name
				published {
					name
				}
			}
		}`,
		Docs: map[int][]string{
			//books
			0: {
				// "bae-3d236f89-6a31-5add-a36a-27971a2eac76"
				`{
					"name": "Painted House"
				}`,
				// "bae-c2f3f08b-53f2-5b53-9a9f-da1eee096321"
				`{
					"name": "Theif Lord"
				}`,
			},
			//authors
			1: {
				`{
					"name": "John Grisham",
					"published_id": "bae-3d236f89-6a31-5add-a36a-27971a2eac76"
				}`,
				`{
					"name": "Cornelia Funke",
					"published_id": "bae-c2f3f08b-53f2-5b53-9a9f-da1eee096321"
				}`,
			},
		},
		Results: []map[string]any{
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
	}

	executeTestCase(t, test)
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
		Results: []map[string]any{
			{
				"name":      "John Grisham",
				"published": nil,
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
		Results: []map[string]any{
			{
				"name":   "Painted House",
				"author": nil,
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
				// bae-3d236f89-6a31-5add-a36a-27971a2eac76
				CollectionID: 0,
				Doc: `{
					"name": "Painted House"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"published_id": "bae-3d236f89-6a31-5add-a36a-27971a2eac76"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Author {
						name
						published_id
					}
				}`,
				Results: []map[string]any{
					{
						"name":         "John Grisham",
						"published_id": "bae-3d236f89-6a31-5add-a36a-27971a2eac76",
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
				// bae-3d236f89-6a31-5add-a36a-27971a2eac76
				CollectionID: 0,
				Doc: `{
					"name": "Painted House"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"published_id": "bae-3d236f89-6a31-5add-a36a-27971a2eac76"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book {
						name
						author_id
					}
				}`,
				Results: []map[string]any{
					{
						"name":      "Painted House",
						"author_id": "bae-6b624301-3d0a-5336-bd2c-ca00bca3de85",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
