// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_one

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestOnetoOneSubTypeDscOrderByQueryWithFilterHavinghNoSubTypeSelections(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one subtype descending order query with filter, no subtype child fields selected.",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"published_id": "bae-be6d8024-4953-5a92-84b4-f042d25230c6"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false,
					"published_id": "bae-26a28d23-ae5b-5257-91b7-d4f2c6abef7b"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book(
						filter: {author: {age: {_gt: 5}}},
						order: {author: {age: DESC}}
					){
						name
						rating
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":   "Painted House",
							"rating": 4.9,
						},
						{
							"name":   "Theif Lord",
							"rating": 4.8,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestOnetoOneSubTypeAscOrderByQueryWithFilterHavinghNoSubTypeSelections(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-one subtype ascending order query with filter, no subtype child fields selected.",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"published_id": "bae-be6d8024-4953-5a92-84b4-f042d25230c6"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false,
					"published_id": "bae-26a28d23-ae5b-5257-91b7-d4f2c6abef7b"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book(
						filter: {author: {age: {_gt: 5}}},
						order: {author: {age: ASC}}
					){
						name
						rating
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":   "Theif Lord",
							"rating": 4.8,
						},
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
