// Copyright 2023 Democratized Data Foundation
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

func TestQueryOneToOneWithGroupRelatedID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation query with group by related id (primary side)",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Book {
						name: String
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
				Doc: `{
					"name": "Painted House"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Go Guide for Rust developers"
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
					"name":         "John Grisham",
					"published_id": testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.Request{
				Request: `query {
					Book(groupBy: [author_id]) {
						author_id
						_group {
							name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"author_id": "bae-077b5e8d-5a86-5ae7-a321-ac7e423bb260",
						"_group": []map[string]any{
							{
								"name": "Painted House",
							},
						},
					},
					{
						"author_id": "bae-cfee1ed9-ede8-5b80-a6fa-78c727a076ac",
						"_group": []map[string]any{
							{
								"name": "Go Guide for Rust developers",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneWithGroupRelatedIDFromSecondaryWithoutGroup(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation query with group by related id (secondary side)",
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
				CollectionID: 0,
				Doc: `{
					"name": "Go Guide for Rust developers"
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
					"name":         "Andrew Lone",
					"published_id": testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.Request{
				Request: `query {
					Book(groupBy: [author_id]) {
						author_id
					}
				}`,
				Results: []map[string]any{
					{
						"author_id": "bae-3c308f94-dc9e-5262-b0ce-ef4e8e545820",
					},
					{
						"author_id": "bae-420e72a6-e0c6-5a06-a958-2cc7adb7b3d0",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneWithGroupRelatedIDFromSecondaryWithoutGroupWithJoin(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation query with group by related id (secondary side)",
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
				CollectionID: 0,
				Doc: `{
					"name": "Go Guide for Rust developers"
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
					"name":         "Andrew Lone",
					"published_id": testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.Request{
				Request: `query {
					Book(groupBy: [author_id]) {
						author_id
						author {
							name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"author_id": "bae-3c308f94-dc9e-5262-b0ce-ef4e8e545820",
						"author": map[string]any{
							"name": "Andrew Lone",
						},
					},
					{
						"author_id": "bae-420e72a6-e0c6-5a06-a958-2cc7adb7b3d0",
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

func TestQueryOneToOneWithGroupRelatedIDFromSecondaryWithGroup(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation query with group by related id (secondary side)",
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
				CollectionID: 0,
				Doc: `{
					"name": "Go Guide for Rust developers"
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
					"name":         "John Grisham",
					"published_id": testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.Request{
				Request: `query {
					Book(groupBy: [author_id]) {
						author_id
						_group {
							name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"author_id": "bae-bb4d6e89-e8b4-5eec-bfeb-6f7aa4840950",
						"_group": []map[string]any{
							{
								"name": "Go Guide for Rust developers",
							},
						},
					},
					{
						"author_id": "bae-420e72a6-e0c6-5a06-a958-2cc7adb7b3d0",
						"_group": []map[string]any{
							{
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

func TestQueryOneToOneWithGroupRelatedIDFromSecondaryWithGroupWithJoin(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation query with group by related id (secondary side)",
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
				CollectionID: 0,
				Doc: `{
					"name": "Go Guide for Rust developers"
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
					"name":         "Andrew Lone",
					"published_id": testUtils.NewDocIndex(0, 1),
				},
			},
			testUtils.Request{
				Request: `query {
					Book(groupBy: [author_id]) {
						author_id
						author {
							name
						}
						_group {
							name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"author_id": "bae-3c308f94-dc9e-5262-b0ce-ef4e8e545820",
						"author": map[string]any{
							"name": "Andrew Lone",
						},
						"_group": []map[string]any{
							{
								"name": "Go Guide for Rust developers",
							},
						},
					},
					{
						"author_id": "bae-420e72a6-e0c6-5a06-a958-2cc7adb7b3d0",
						"author": map[string]any{
							"name": "John Grisham",
						},
						"_group": []map[string]any{
							{
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
