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

func TestQueryOneToManyWithChildDocIDs(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from one side with child docIDs",
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
					"name": "The Firm",
					"rating": 4.5,
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
					Author {
						name
						published (
								docID: ["bae-064f13c1-7726-5d53-8eec-c395d94da4d0", "bae-649c8101-76b8-5d18-a701-21c97a5c66b3"]
							) {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John Grisham",
							"published": []map[string]any{
								{
									"name": "Painted House",
								},
								{
									"name": "The Associate",
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
