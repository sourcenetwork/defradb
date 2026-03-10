// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package one_to_one

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationAddOneToOne_WithExplicitNullOnPrimarySide(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
