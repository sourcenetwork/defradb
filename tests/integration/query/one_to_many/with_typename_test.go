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
	test := testUtils.RequestTestCase{
		Description: "One-to-many relation query from one side with typename",
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
		Docs: map[int][]string{
			//books
			0: { // bae-be6d8024-4953-5a92-84b4-f042d25230c6
				`{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			//authors
			1: { // bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
		},
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
	}

	executeTestCase(t, test)
}
