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

func TestQueryOneToManyWithNumericGreaterThanFilterOnParentAndNumericSortAscendingAndLimitOnChild(
	t *testing.T,
) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many relation query from the many side, asc. order & limit on sub",
		Request: `query {
			Author(filter: {age: {_gt: 63}}) {
				name
				age
				published(order: {rating: ASC}, limit: 1) {
					name
					rating
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
				`{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
				`{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-72e8c691-9f20-55e7-9228-8af1cf54cace"
				}`,
			},
			//authors
			1: {
				// bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-72e8c691-9f20-55e7-9228-8af1cf54cace
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},
		Results: map[string]any{
			"Author": []map[string]any{
				{
					"name": "John Grisham",
					"age":  int64(65),
					"published": []map[string]any{
						{
							"name":   "A Time for Mercy",
							"rating": 4.5,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToManyWithNumericGreaterThanFilterOnParentAndNumericSortDescendingAndLimitOnChild(
	t *testing.T,
) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many relation query from the many side, desc. order & limit on sub",
		Request: `query {
			Author(filter: {age: {_gt: 63}}) {
				name
				age
				published(order: {rating: DESC}, limit: 1) {
					name
					rating
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
				`{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
				`{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-72e8c691-9f20-55e7-9228-8af1cf54cace"
				}`,
			},
			//authors
			1: {
				// bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-72e8c691-9f20-55e7-9228-8af1cf54cace
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},
		Results: map[string]any{
			"Author": []map[string]any{
				{
					"name": "John Grisham",
					"age":  int64(65),
					"published": []map[string]any{
						{
							"name":   "Painted House",
							"rating": 4.9,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
