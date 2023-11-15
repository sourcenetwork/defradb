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
				0: { // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
					`{
						"name": "Painted House",
						"rating": 4.9,
						"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
					}`,
				},
				//authors
				1: { // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
					`{
						"name": "John Grisham",
						"age": 65,
						"verified": true
					}`,
				},
			},
			Results: []map[string]any{
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
				0: { // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
					`{
						"name": "Painted House",
						"rating": 4.9,
						"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
					}`,
					`{
						"name": "A Time for Mercy",
						"rating": 4.5,
						"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
						}`,
					`{
						"name": "Theif Lord",
						"rating": 4.8,
						"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
					}`,
				},
				//authors
				1: {
					// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
					`{
						"name": "John Grisham",
						"age": 65,
						"verified": true
					}`,
					// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
					`{
						"name": "Cornelia Funke",
						"age": 62,
						"verified": false
					}`,
				},
			},
			Results: []map[string]any{
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
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name":   "Painted House",
				"rating": 4.9,
				"author": nil,
			},
		},
	}

	executeTestCase(t, test)
}
