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

func TestQueryOneToOneWithGroupRelatedIDAlias(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation query with group by related id alias (primary side)",
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
				CollectionID: 1,
				DocMap: map[string]any{
					"name": "John Grisham",
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name": "Andrew Lone",
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Painted House",
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Go Guide for Rust developers",
					"author_id": testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.Request{
				Request: `query {
					Book(groupBy: [author]) {
						author_id
						author {
							name
						}
						_group {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"author_id": "bae-547eb3d8-7fc8-5c21-bcef-590813451e55",
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
							"author_id": "bae-ee5973cf-73c3-558f-8aec-8b590b8e77cf",
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneWithGroupRelatedIDAliasFromSecondaryWithoutInnerGroup(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation query with group by related id alias (secondary side)",
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
					Book(groupBy: [author]) {
						author_id
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"author_id": "bae-23a33112-7345-52f1-8816-0481747645f2",
						},
						{
							"author_id": "bae-35fc1c36-4347-5bf4-a41f-bf676b145075",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneWithGroupRelatedIDAliasFromSecondaryWithoutInnerGroupWithJoin(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation query with group by related id alias (secondary side)",
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
					Book(groupBy: [author]) {
						author_id
						author {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"author_id": "bae-23a33112-7345-52f1-8816-0481747645f2",
							"author": map[string]any{
								"name": "Andrew Lone",
							},
						},
						{
							"author_id": "bae-35fc1c36-4347-5bf4-a41f-bf676b145075",
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

func TestQueryOneToOneWithGroupRelatedIDAliasFromSecondaryWithInnerGroup(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation query with group by related id alias (secondary side)",
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
					Book(groupBy: [author]) {
						author_id
						_group {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"author_id": "bae-23a33112-7345-52f1-8816-0481747645f2",
							"_group": []map[string]any{
								{
									"name": "Go Guide for Rust developers",
								},
							},
						},
						{
							"author_id": "bae-35fc1c36-4347-5bf4-a41f-bf676b145075",
							"_group": []map[string]any{
								{
									"name": "Painted House",
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

func TestQueryOneToOneWithGroupRelatedIDAliasFromSecondaryWithInnerGroupWithJoin(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation query with group by related id alias (secondary side)",
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
					Book(groupBy: [author]) {
						author_id
						author {
							name
						}
						_group {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"author_id": "bae-23a33112-7345-52f1-8816-0481747645f2",
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
							"author_id": "bae-35fc1c36-4347-5bf4-a41f-bf676b145075",
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
