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

package one_to_one_to_one

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToOneToOneWithNestedOrder(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name": "John Grisham",
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name": "Cornelia Funke",
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "Painted House",
					"_authorID": testUtils.NewDocIndex(2, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"_authorID": testUtils.NewDocIndex(2, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":       "Old Publisher",
					"_printedID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":       "New Publisher",
					"_printedID": testUtils.NewDocIndex(1, 1),
				},
			},
			&action.Request{
				Request: `query {
					Publisher(order: {printed: {author: {name: ASC}}}) {
						name
						printed {
							name
							author {
								name
							}
						}
					}
				}`,
				Results: map[string]any{
					"Publisher": []map[string]any{
						{
							"name": "New Publisher",
							"printed": map[string]any{
								"name": "Theif Lord",
								"author": map[string]any{
									"name": "Cornelia Funke",
								},
							},
						},
						{
							"name": "Old Publisher",
							"printed": map[string]any{
								"name": "Painted House",
								"author": map[string]any{
									"name": "John Grisham",
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
