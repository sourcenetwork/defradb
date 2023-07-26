// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_one_to_one

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToOneToOne(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one-to-one relation primary direction",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
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
				// "bae-a6cdabfc-17dd-5662-b213-c596ee4c3292"
				Doc: `{
					"name": "Painted House",
					"publisher_id": "bae-1f4cc394-08a8-5825-87b9-b02de2f25f7d"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// "bae-bc198c5f-6238-5b50-8072-68dec9c7a16b"
				Doc: `{
					"name": "Theif Lord",
					"publisher_id": "bae-a3cd6fac-13c0-5c8f-970b-0ce7abbb49a5"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				Doc: `{
					"name": "John Grisham",
					"published_id": "bae-a6cdabfc-17dd-5662-b213-c596ee4c3292"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				Doc: `{
					"name": "Cornelia Funke",
					"published_id": "bae-bc198c5f-6238-5b50-8072-68dec9c7a16b"
				}`,
			},
			testUtils.Request{
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
				Results: []map[string]any{
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
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneToOneSecondaryThenPrimary(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one-to-one relation, secondary then primary direction",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
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
				// "bae-a6cdabfc-17dd-5662-b213-c596ee4c3292"
				Doc: `{
					"name": "Painted House",
					"publisher_id": "bae-1f4cc394-08a8-5825-87b9-b02de2f25f7d"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// "bae-bc198c5f-6238-5b50-8072-68dec9c7a16b"
				Doc: `{
					"name": "Theif Lord",
					"publisher_id": "bae-a3cd6fac-13c0-5c8f-970b-0ce7abbb49a5"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				Doc: `{
					"name": "John Grisham",
					"published_id": "bae-a6cdabfc-17dd-5662-b213-c596ee4c3292"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				Doc: `{
					"name": "Cornelia Funke",
					"published_id": "bae-bc198c5f-6238-5b50-8072-68dec9c7a16b"
				}`,
			},
			testUtils.Request{
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
				Results: []map[string]any{
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
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneToOnePrimaryThenSecondary(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one-to-one relation, primary then secondary direction",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
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
				// "bae-a6cdabfc-17dd-5662-b213-c596ee4c3292"
				Doc: `{
					"name": "Painted House",
					"publisher_id": "bae-1f4cc394-08a8-5825-87b9-b02de2f25f7d"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// "bae-bc198c5f-6238-5b50-8072-68dec9c7a16b"
				Doc: `{
					"name": "Theif Lord",
					"publisher_id": "bae-a3cd6fac-13c0-5c8f-970b-0ce7abbb49a5"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				Doc: `{
					"name": "John Grisham",
					"published_id": "bae-a6cdabfc-17dd-5662-b213-c596ee4c3292"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				Doc: `{
					"name": "Cornelia Funke",
					"published_id": "bae-bc198c5f-6238-5b50-8072-68dec9c7a16b"
				}`,
			},
			testUtils.Request{
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
				Results: []map[string]any{
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
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneToOneSecondary(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one-to-one relation, secondary direction",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
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
				// "bae-a6cdabfc-17dd-5662-b213-c596ee4c3292"
				Doc: `{
					"name": "Painted House",
					"publisher_id": "bae-1f4cc394-08a8-5825-87b9-b02de2f25f7d"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// "bae-bc198c5f-6238-5b50-8072-68dec9c7a16b"
				Doc: `{
					"name": "Theif Lord",
					"publisher_id": "bae-a3cd6fac-13c0-5c8f-970b-0ce7abbb49a5"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				Doc: `{
					"name": "John Grisham",
					"published_id": "bae-a6cdabfc-17dd-5662-b213-c596ee4c3292"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				Doc: `{
					"name": "Cornelia Funke",
					"published_id": "bae-bc198c5f-6238-5b50-8072-68dec9c7a16b"
				}`,
			},
			testUtils.Request{
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
				Results: []map[string]any{
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
	}

	testUtils.ExecuteTestCase(t, test)
}
