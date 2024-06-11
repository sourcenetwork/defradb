// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_one_multiple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToOneMultiple_FromPrimary(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Multiple one-to-one joins from primary direction",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Publisher {
						name: String
						printed: Book
					}

					type Author {
						name: String
						published: Book
					}

					type Book {
						name: String
						publisher: Publisher @primary
						author: Author @primary
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Old Publisher"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "New Publisher"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Cornelia Funke"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":         "Painted House",
					"publisher_id": testUtils.NewDocIndex(0, 0),
					"author_id":    testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":         "Theif Lord",
					"publisher_id": testUtils.NewDocIndex(0, 1),
					"author_id":    testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.Request{
				Request: `query {
					Book {
						name
						publisher {
							name
						}
						author {
							name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Painted House",
						"publisher": map[string]any{
							"name": "Old Publisher",
						},
						"author": map[string]any{
							"name": "John Grisham",
						},
					},
					{
						"name": "Theif Lord",
						"publisher": map[string]any{
							"name": "New Publisher",
						},
						"author": map[string]any{
							"name": "Cornelia Funke",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneMultiple_FromMixedPrimaryAndSecondary(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Multiple one-to-one joins from primary direction",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Publisher {
						name: String
						printed: Book @primary
					}

					type Author {
						name: String
						published: Book
					}

					type Book {
						name: String
						publisher: Publisher
						author: Author @primary
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Old Publisher"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "New Publisher"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Cornelia Funke"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":         "Painted House",
					"publisher_id": testUtils.NewDocIndex(0, 0),
					"author_id":    testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":         "Theif Lord",
					"publisher_id": testUtils.NewDocIndex(0, 1),
					"author_id":    testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.Request{
				Request: `query {
					Book {
						name
						publisher {
							name
						}
						author {
							name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Theif Lord",
						"publisher": map[string]any{
							"name": "New Publisher",
						},
						"author": map[string]any{
							"name": "Cornelia Funke",
						},
					},
					{
						"name": "Painted House",
						"publisher": map[string]any{
							"name": "Old Publisher",
						},
						"author": map[string]any{
							"name": "John Grisham",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneMultiple_FromSecondary(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Multiple one-to-one joins from primary direction",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Publisher {
						name: String
						printed: Book @primary
					}

					type Author {
						name: String
						published: Book @primary
					}

					type Book {
						name: String
						publisher: Publisher
						author: Author
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Old Publisher"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "New Publisher"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Cornelia Funke"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":         "Painted House",
					"publisher_id": testUtils.NewDocIndex(0, 0),
					"author_id":    testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":         "Theif Lord",
					"publisher_id": testUtils.NewDocIndex(0, 1),
					"author_id":    testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.Request{
				Request: `query {
					Book {
						name
						publisher {
							name
						}
						author {
							name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Painted House",
						"publisher": map[string]any{
							"name": "Old Publisher",
						},
						"author": map[string]any{
							"name": "John Grisham",
						},
					},
					{
						"name": "Theif Lord",
						"publisher": map[string]any{
							"name": "New Publisher",
						},
						"author": map[string]any{
							"name": "Cornelia Funke",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
