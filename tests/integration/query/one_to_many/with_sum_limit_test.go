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

func TestQueryOneToManyWithSumWithLimit(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from many side with sum with limit",
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
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Associate",
					"rating": 4.2,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-72e8c691-9f20-55e7-9228-8af1cf54cace"
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
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			testUtils.Request{
				Request: `query {
					Author {
						name
						_sum(published: {field: rating, limit: 2})
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "Cornelia Funke",
							"_sum": 4.8,
						},
						{
							"name": "John Grisham",
							// .00...1 is float math thing
							"_sum": 9.100000000000001,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
