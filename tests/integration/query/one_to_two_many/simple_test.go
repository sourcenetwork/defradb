// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_two_many

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToTwoManyWithNilUnnamedRelationship(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "One-to-many relation query from one side",
			Request: `query {
						Book {
							name
							rating
							author {
								name
							}
							reviewedBy {
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
						"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3",
						"reviewedBy_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
					}`,
					`{
						"name": "A Time for Mercy",
						"rating": 4.5,
						"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3",
						"reviewedBy_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
					}`,
					`{
						"name": "Theif Lord",
						"rating": 4.8,
						"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04",
						"reviewedBy_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
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
					"name":   "Painted House",
					"rating": 4.9,
					"author": map[string]any{
						"name": "John Grisham",
					},
					"reviewedBy": map[string]any{
						"name": "Cornelia Funke",
						"age":  uint64(62),
					},
				},
				{
					"name":   "Theif Lord",
					"rating": 4.8,
					"author": map[string]any{
						"name": "Cornelia Funke",
					},
					"reviewedBy": map[string]any{
						"name": "John Grisham",
						"age":  uint64(65),
					},
				},
				{
					"name":   "A Time for Mercy",
					"rating": 4.5,
					"author": map[string]any{
						"name": "John Grisham",
					},
					"reviewedBy": map[string]any{
						"name": "Cornelia Funke",
						"age":  uint64(62),
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
					written {
						name
					}
					reviewed {
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
						"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3",
						"reviewedBy_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
					}`,
					`{
						"name": "A Time for Mercy",
						"rating": 4.5,
						"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3",
						"reviewedBy_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
					}`,
					`{
						"name": "Theif Lord",
						"rating": 4.8,
						"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04",
						"reviewedBy_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
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
					"age":  uint64(65),
					"reviewed": []map[string]any{
						{
							"name":   "Theif Lord",
							"rating": 4.8,
						},
					},
					"written": []map[string]any{
						{
							"name": "Painted House",
						},
						{
							"name": "A Time for Mercy",
						},
					},
				},
				{
					"name": "Cornelia Funke",
					"age":  uint64(62),
					"reviewed": []map[string]any{
						{
							"name":   "Painted House",
							"rating": 4.9,
						},
						{
							"name":   "A Time for Mercy",
							"rating": 4.5,
						},
					},
					"written": []map[string]any{
						{
							"name": "Theif Lord",
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

func TestQueryOneToTwoManyWithNamedAndUnnamedRelationships(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "One-to-many relation query from one side",
			Request: `query {
						Book {
							name
							rating
							author {
								name
							}
							reviewedBy {
								name
								age
							}
							price {
								currency
								value
							}
						}
					}`,
			Docs: map[int][]string{
				//books
				0: {
					`{
						"name": "Painted House",
						"rating": 4.9,
						"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3",
						"reviewedBy_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04",
						"price_id": "bae-fcc7a01d-6855-5e7a-abdd-261a46dcb9bd"
					}`,
					`{
						"name": "A Time for Mercy",
						"rating": 4.5,
						"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3",
						"reviewedBy_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04",
						"price_id": "bae-d64a5165-1e77-5a67-95f2-6b1ff14b2179"
					}`,
					`{
						"name": "Theif Lord",
						"rating": 4.8,
						"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04",
						"reviewedBy_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3",
						"price_id": "bae-fcc7a01d-6855-5e7a-abdd-261a46dcb9bd"
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
				2: {
					// bae-fcc7a01d-6855-5e7a-abdd-261a46dcb9bd
					`{
						"currency": "GBP",
						"value": 12.99
					}`,
					// bae-d64a5165-1e77-5a67-95f2-6b1ff14b2179
					`{
						"currency": "SEK",
						"value": 129
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name":   "Theif Lord",
					"rating": 4.8,
					"author": map[string]any{
						"name": "Cornelia Funke",
					},
					"reviewedBy": map[string]any{
						"name": "John Grisham",
						"age":  uint64(65),
					},
					"price": map[string]any{
						"currency": "GBP",
						"value":    12.99,
					},
				},
				{
					"name":   "A Time for Mercy",
					"rating": 4.5,
					"author": map[string]any{
						"name": "John Grisham",
					},
					"reviewedBy": map[string]any{
						"name": "Cornelia Funke",
						"age":  uint64(62),
					},
					"price": map[string]any{
						"currency": "SEK",
						"value":    float64(129),
					},
				},
				{
					"name":   "Painted House",
					"rating": 4.9,
					"author": map[string]any{
						"name": "John Grisham",
					},
					"reviewedBy": map[string]any{
						"name": "Cornelia Funke",
						"age":  uint64(62),
					},
					"price": map[string]any{
						"currency": "GBP",
						"value":    12.99,
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
					written {
						name
						price {
							value
						}
					}
					reviewed {
						name
						rating
					}
				}
			}`,
			Docs: map[int][]string{
				//books
				0: {
					`{
						"name": "Painted House",
						"rating": 4.9,
						"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3",
						"reviewedBy_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04",
						"price_id": "bae-fcc7a01d-6855-5e7a-abdd-261a46dcb9bd"
					}`,
					`{
						"name": "A Time for Mercy",
						"rating": 4.5,
						"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3",
						"reviewedBy_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04",
						"price_id": "bae-d64a5165-1e77-5a67-95f2-6b1ff14b2179"
					}`,
					`{
						"name": "Theif Lord",
						"rating": 4.8,
						"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04",
						"reviewedBy_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3",
						"price_id": "bae-fcc7a01d-6855-5e7a-abdd-261a46dcb9bd"
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
				2: {
					// bae-fcc7a01d-6855-5e7a-abdd-261a46dcb9bd
					`{
						"currency": "GBP",
						"value": 12.99
					}`,
					// bae-d64a5165-1e77-5a67-95f2-6b1ff14b2179
					`{
						"currency": "SEK",
						"value": 129
					}`,
				},
			},
			Results: []map[string]any{
				{
					"name": "John Grisham",
					"age":  uint64(65),
					"reviewed": []map[string]any{
						{
							"name":   "Theif Lord",
							"rating": 4.8,
						},
					},
					"written": []map[string]any{
						{
							"name": "A Time for Mercy",
							"price": map[string]any{
								"value": float64(129),
							},
						},
						{
							"name": "Painted House",
							"price": map[string]any{
								"value": 12.99,
							},
						},
					},
				},
				{
					"name": "Cornelia Funke",
					"age":  uint64(62),
					"reviewed": []map[string]any{
						{
							"name":   "A Time for Mercy",
							"rating": 4.5,
						},
						{
							"name":   "Painted House",
							"rating": 4.9,
						},
					},
					"written": []map[string]any{
						{
							"name": "Theif Lord",
							"price": map[string]any{
								"value": 12.99,
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
