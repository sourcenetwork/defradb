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

func TestQueryOneToMany(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "One-to-many relation query from one side",
			Request: `query {
						Book {
							name
							rating
							author {
								name
								age
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
						"name":   "Painted House",
						"rating": 4.9,
						"author": map[string]any{
							"name": "John Grisham",
							"age":  int64(65),
						},
					},
				},
			},
		},
		{
			Description: "One-to-many relation query from many side",
			Request: `query {
				Author {
					name
					age
					published {
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
						"name": "Cornelia Funke",
						"age":  int64(62),
						"published": []map[string]any{
							{
								"name":   "Theif Lord",
								"rating": 4.8,
							},
						},
					},
					{
						"name": "John Grisham",
						"age":  int64(65),
						"published": []map[string]any{
							{
								"name":   "Painted House",
								"rating": 4.9,
							},
							{
								"name":   "A Time for Mercy",
								"rating": 4.5,
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQueryOneToManyWithNonExistantParent(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many relation query from one side with non-existant parent",
		Request: `query {
						Book {
							name
							rating
							author {
								name
								age
							}
						}
					}`,
		Docs: map[int][]string{
			//books
			0: {
				`{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
		},
		Results: map[string]any{
			"Book": []map[string]any{
				{
					"name":   "Painted House",
					"rating": 4.9,
					"author": nil,
				},
			},
		},
	}

	executeTestCase(t, test)
}
