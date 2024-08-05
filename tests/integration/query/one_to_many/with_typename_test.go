// Copyright 2022 Democratized Data Foundation
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

func TestQueryOneToManyWithTypeName(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from one side with typename",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book {
						name
						__typename
						author {
							name
							__typename
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":       "Painted House",
							"__typename": "Book",
							"author": map[string]any{
								"name":       "John Grisham",
								"__typename": "Author",
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
