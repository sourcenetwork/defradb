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
				// bae-3d236f89-6a31-5add-a36a-27971a2eac76
				Doc: `{
					"name": "Painted House"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-d6627fea-8bf7-511c-bcf9-bac4212bddd6
				Doc: `{
					"name": "Go Guide for Rust developers"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-6b624301-3d0a-5336-bd2c-ca00bca3de85
				Doc: `{
					"name": "John Grisham",
					"published_id": "bae-3d236f89-6a31-5add-a36a-27971a2eac76"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-92fa9dcb-c1ee-5b84-b2f6-e9437c7f261c
				Doc: `{
					"name": "Andrew Lone",
					"published_id": "bae-d6627fea-8bf7-511c-bcf9-bac4212bddd6"
				}`,
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
						"author_id": "bae-6b624301-3d0a-5336-bd2c-ca00bca3de85",
						"_group": []map[string]any{
							{
								"name": "Painted House",
							},
						},
					},
					{
						"author_id": "bae-92fa9dcb-c1ee-5b84-b2f6-e9437c7f261c",
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

	testUtils.ExecuteTEMP(t, test)
}

// This test documents unwanted behaviour, see:
// https://github.com/sourcenetwork/defradb/issues/1654
func TestQueryOneToOneWithGroupRelatedIDFromSecondary(t *testing.T) {
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
				// bae-3d236f89-6a31-5add-a36a-27971a2eac76
				Doc: `{
					"name": "Painted House"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-d6627fea-8bf7-511c-bcf9-bac4212bddd6
				Doc: `{
					"name": "Go Guide for Rust developers"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-6b624301-3d0a-5336-bd2c-ca00bca3de85
				Doc: `{
					"name": "John Grisham",
					"published_id": "bae-3d236f89-6a31-5add-a36a-27971a2eac76"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-92fa9dcb-c1ee-5b84-b2f6-e9437c7f261c
				Doc: `{
					"name": "Andrew Lone",
					"published_id": "bae-d6627fea-8bf7-511c-bcf9-bac4212bddd6"
				}`,
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
						"author_id": nil,
						"_group": []map[string]any{
							{
								"name": "Painted House",
							},
							{
								"name": "Go Guide for Rust developers",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTEMP(t, test)
}
