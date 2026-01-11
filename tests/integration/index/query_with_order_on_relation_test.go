// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package index

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// TestOrderQueryWithIndex_WithOrderOnIndexedFieldOfOneToMany_ShouldRespectOrder tests that
// ordering on an indexed field of a one-to-many relation yields different results for ASC and DESC.
func TestOrderQueryWithIndex_WithOrderOnIndexedFieldOfOneToMany_ShouldRespectOrder(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Author {
						name: String
						published: [Book]
					}

					type Book {
						name: String
						rating: Float @index
						author: Author
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":   "A Time for Mercy",
					"rating": 4.5,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":   "Painted House",
					"rating": 4.9,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":   "Sooley",
					"rating": 4.2,
					"author": testUtils.NewDocIndex(0, 0),
				},
			},
			// Test ASC ordering
			testUtils.Request{
				Request: `query {
					Author {
						name
						published(order: {rating: ASC}) {
							name
							rating
						}
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John Grisham",
							"published": []map[string]any{
								{
									"name":   "Sooley",
									"rating": 4.2,
								},
								{
									"name":   "A Time for Mercy",
									"rating": 4.5,
								},
								{
									"name":   "Painted House",
									"rating": 4.9,
								},
							},
						},
					},
				},
			},
			// Test DESC ordering (should be different from ASC)
			testUtils.Request{
				Request: `query {
					Author {
						name
						published(order: {rating: DESC}) {
							name
							rating
						}
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John Grisham",
							"published": []map[string]any{
								{
									"name":   "Painted House",
									"rating": 4.9,
								},
								{
									"name":   "A Time for Mercy",
									"rating": 4.5,
								},
								{
									"name":   "Sooley",
									"rating": 4.2,
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
