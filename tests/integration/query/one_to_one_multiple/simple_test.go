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
				// "bae-1f4cc394-08a8-5825-87b9-b02de2f25f7d"
				Doc: `{
					"name": "Old Publisher"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// "bae-a3cd6fac-13c0-5c8f-970b-0ce7abbb49a5"
				Doc: `{
					"name": "New Publisher"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-b6ea52b8-a5a5-5127-b9c0-5df4243457a3
				Doc: `{
					"name": "Cornelia Funke"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				Doc: `{
					"name": "Painted House",
					"publisher_id": "bae-1f4cc394-08a8-5825-87b9-b02de2f25f7d",
					"author_id": "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				Doc: `{
					"name": "Theif Lord",
					"publisher_id": "bae-a3cd6fac-13c0-5c8f-970b-0ce7abbb49a5",
					"author_id": "bae-b6ea52b8-a5a5-5127-b9c0-5df4243457a3"
				}`,
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
				// "bae-1f4cc394-08a8-5825-87b9-b02de2f25f7d"
				Doc: `{
					"name": "Old Publisher"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// "bae-a3cd6fac-13c0-5c8f-970b-0ce7abbb49a5"
				Doc: `{
					"name": "New Publisher"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-b6ea52b8-a5a5-5127-b9c0-5df4243457a3
				Doc: `{
					"name": "Cornelia Funke"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				Doc: `{
					"name": "Painted House",
					"publisher_id": "bae-1f4cc394-08a8-5825-87b9-b02de2f25f7d",
					"author_id": "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				Doc: `{
					"name": "Theif Lord",
					"publisher_id": "bae-a3cd6fac-13c0-5c8f-970b-0ce7abbb49a5",
					"author_id": "bae-b6ea52b8-a5a5-5127-b9c0-5df4243457a3"
				}`,
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
				// "bae-1f4cc394-08a8-5825-87b9-b02de2f25f7d"
				Doc: `{
					"name": "Old Publisher"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// "bae-a3cd6fac-13c0-5c8f-970b-0ce7abbb49a5"
				Doc: `{
					"name": "New Publisher"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-b6ea52b8-a5a5-5127-b9c0-5df4243457a3
				Doc: `{
					"name": "Cornelia Funke"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				Doc: `{
					"name": "Painted House",
					"publisher_id": "bae-1f4cc394-08a8-5825-87b9-b02de2f25f7d",
					"author_id": "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				Doc: `{
					"name": "Theif Lord",
					"publisher_id": "bae-a3cd6fac-13c0-5c8f-970b-0ce7abbb49a5",
					"author_id": "bae-b6ea52b8-a5a5-5127-b9c0-5df4243457a3"
				}`,
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
