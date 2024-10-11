// Copyright 2024 Democratized Data Foundation
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

func TestMutationCreateOneToOne_WithExplicitNullOnPrimarySide(t *testing.T) {
	test := testUtils.TestCase{
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
					"name": "Will Ferguson",
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name":   "How to Be a Canadian",
					"author": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Secrets at Maple Syrup Farm",
					"author": null
				}`,
			},
			testUtils.Request{
				Request: `
					query {
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
							"name":   "Secrets at Maple Syrup Farm",
							"author": nil,
						},
						{
							"name": "How to Be a Canadian",
							"author": map[string]any{
								"name": "Will Ferguson",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
