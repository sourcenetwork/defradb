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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationAddOneToOne_WithExplicitNullOnPrimarySide(t *testing.T) {
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
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name": "Will Ferguson",
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name":   "How to Be a Canadian",
					"author": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				Doc: `{
					"name": "Secrets at Maple Syrup Farm",
					"author": null
				}`,
			},
			&action.Request{
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
							"name": "How to Be a Canadian",
							"author": map[string]any{
								"name": "Will Ferguson",
							},
						},
						{
							"name":   "Secrets at Maple Syrup Farm",
							"author": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
