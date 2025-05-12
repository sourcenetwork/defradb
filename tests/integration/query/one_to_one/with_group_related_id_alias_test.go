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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToOneWithGroupRelatedIDAlias(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation query with group by related id alias (primary side)",
		Actions: []any{
			&action.AddSchema{
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
							"author_id": "bae-46209ee9-ef8c-5bf1-9c99-fe764cec3148",
							"author": map[string]any{
								"name": "John Grisham",
							},
							"_group": []map[string]any{
								{
									"name": "Painted House",
								},
							},
						},
						{
							"author_id": "bae-aad433b7-fe14-5a31-a5da-94735bedcd4f",
							"author": map[string]any{
								"name": "Andrew Lone",
							},
							"_group": []map[string]any{
								{
									"name": "Go Guide for Rust developers",
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
							"author_id": "bae-0362a1da-4a34-5c53-97a3-f5bdcea5d78f",
						},
						{
							"author_id": "bae-382a1634-1fde-536f-8812-5021d924da66",
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
							"author_id": "bae-0362a1da-4a34-5c53-97a3-f5bdcea5d78f",
							"author": map[string]any{
								"name": "John Grisham",
							},
						},
						{
							"author_id": "bae-382a1634-1fde-536f-8812-5021d924da66",
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

func TestQueryOneToOneWithGroupRelatedIDAliasFromSecondaryWithInnerGroup(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one relation query with group by related id alias (secondary side)",
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
							"author_id": "bae-0362a1da-4a34-5c53-97a3-f5bdcea5d78f",
							"_group": []map[string]any{
								{
									"name": "Painted House",
								},
							},
						},
						{
							"author_id": "bae-382a1634-1fde-536f-8812-5021d924da66",
							"_group": []map[string]any{
								{
									"name": "Go Guide for Rust developers",
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
							"author_id": "bae-0362a1da-4a34-5c53-97a3-f5bdcea5d78f",
							"author": map[string]any{
								"name": "John Grisham",
							},
							"_group": []map[string]any{
								{
									"name": "Painted House",
								},
							},
						},
						{
							"author_id": "bae-382a1634-1fde-536f-8812-5021d924da66",
							"author": map[string]any{
								"name": "Andrew Lone",
							},
							"_group": []map[string]any{
								{
									"name": "Go Guide for Rust developers",
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
