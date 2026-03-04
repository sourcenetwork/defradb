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

package one_to_many

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToMany_WithAverageAliasFilter_ShouldMatchAll(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Painted House",
					"rating":    4.9,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"rating":    4.5,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"rating":    4.8,
					"_authorID": testUtils.NewDocIndex(1, 1),
				},
			},
			&action.Request{
				Request: `query {
					Author(filter: {_alias: {averageRating: {_gt: 0}}}) {
						name
						averageRating: AVG(published: {field: rating})
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name":          "John Grisham",
							"averageRating": 4.7,
						},
						{
							"name":          "Cornelia Funke",
							"averageRating": 4.8,
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToMany_WithAverageAliasFilter_ShouldMatchOne(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Painted House",
					"rating":    4.9,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"rating":    4.5,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"rating":    4.8,
					"_authorID": testUtils.NewDocIndex(1, 1),
				},
			},
			&action.Request{
				Request: `query {
					Author(filter: {_alias: {averageRating: {_lt: 4.8}}}) {
						name
						averageRating: AVG(published: {field: rating})
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name":          "John Grisham",
							"averageRating": 4.7,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
