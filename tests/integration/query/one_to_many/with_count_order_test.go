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

func TestQueryOneToMany_WithCountAliasOrder_ShouldOrderResults(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from many side with order alias",
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
					Author(order: {_alias: {publishedCount: DESC}}) {
						name
						publishedCount: _count(published: {})
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name":           "John Grisham",
							"publishedCount": int64(2),
						},
						{
							"name":           "Cornelia Funke",
							"publishedCount": int64(1),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
