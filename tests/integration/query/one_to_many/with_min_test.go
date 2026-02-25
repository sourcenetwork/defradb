// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_many

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToMany_WithMinAliasFilter_ShouldMatchAll(t *testing.T) {
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
					Author(filter: {_alias: {minRating: {_gt: 0}}}) {
						name
						minRating: MIN(published: {field: rating})
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name":      "John Grisham",
							"minRating": 4.5,
						},
						{
							"name":      "Cornelia Funke",
							"minRating": 4.8,
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToMany_WithMinAliasFilter_ShouldMatchOne(t *testing.T) {
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
					Author(filter: {_alias: {minRating: {_lt: 4.7}}}) {
						name
						minRating: MIN(published: {field: rating})
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name":      "John Grisham",
							"minRating": 4.5,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
