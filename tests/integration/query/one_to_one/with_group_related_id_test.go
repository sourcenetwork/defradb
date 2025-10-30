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

func TestQueryOneToOneWithGroupRelatedID(t *testing.T) {
	test := testUtils.TestCase{
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
					Book(groupBy: [author_id]) {
						author_id
						_group {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"author_id": "bae-5181bbe5-c134-5e97-8928-30c33d3b83ad",
							"_group": []map[string]any{
								{
									"name": "Painted House",
								},
							},
						},
						{
							"author_id": "bae-b1a6f637-bbbb-59aa-8a54-938249e21cdd",
							"_group": []map[string]any{
								{
									"name": "Go Guide for Rust developers",
								},
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

func TestQueryOneToOneWithGroupRelatedIDFromSecondaryWithoutGroup(t *testing.T) {
	test := testUtils.TestCase{
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
					Book(groupBy: [author_id]) {
						author_id
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"author_id": "bae-f281e7e3-9ad5-5bbe-9e90-13e5ccbec2b5",
						},
						{
							"author_id": "bae-d92e6b41-9df9-519f-b823-c3e13f4e1b0b",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneWithGroupRelatedIDFromSecondaryWithoutGroupWithJoin(t *testing.T) {
	test := testUtils.TestCase{
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
					Book(groupBy: [author_id]) {
						author_id
						author {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"author_id": "bae-d92e6b41-9df9-519f-b823-c3e13f4e1b0b",
							"author": map[string]any{
								"name": "John Grisham",
							},
						},
						{
							"author_id": "bae-f281e7e3-9ad5-5bbe-9e90-13e5ccbec2b5",
							"author": map[string]any{
								"name": "Andrew Lone",
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

func TestQueryOneToOneWithGroupRelatedIDFromSecondaryWithGroup(t *testing.T) {
	test := testUtils.TestCase{
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
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"author_id": "bae-d92e6b41-9df9-519f-b823-c3e13f4e1b0b",
							"_group": []map[string]any{
								{
									"name": "Painted House",
								},
							},
						},
						{
							"author_id": "bae-b45f28db-1f51-5300-a8c2-51aa132e93bd",
							"_group": []map[string]any{
								{
									"name": "Go Guide for Rust developers",
								},
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

func TestQueryOneToOneWithGroupRelatedIDFromSecondaryWithGroupWithJoin(t *testing.T) {
	test := testUtils.TestCase{
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
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"author_id": "bae-d92e6b41-9df9-519f-b823-c3e13f4e1b0b",
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
							"author_id": "bae-f281e7e3-9ad5-5bbe-9e90-13e5ccbec2b5",
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
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
