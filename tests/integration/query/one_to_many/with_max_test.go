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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToMany_WithMaxAliasFilter_ShouldMatchAll(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from many side with max alias",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Painted House",
					"rating":    4.9,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"rating":    4.5,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"rating":    4.8,
					"author_id": testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.Request{
				Request: `query {
					Author(filter: {_alias: {maxRating: {_gt: 0}}}) {
						name
						maxRating: _max(published: {field: rating})
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name":      "Cornelia Funke",
							"maxRating": 4.8,
						},
						{
							"name":      "John Grisham",
							"maxRating": 4.9,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToMany_WithMaxAliasFilter_ShouldMatchOne(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from many side with max alias",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Painted House",
					"rating":    4.9,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"rating":    4.5,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"rating":    4.8,
					"author_id": testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.Request{
				Request: `query {
					Author(filter: {_alias: {maxRating: {_gt: 4.8}}}) {
						name
						maxRating: _max(published: {field: rating})
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name":      "John Grisham",
							"maxRating": 4.9,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
