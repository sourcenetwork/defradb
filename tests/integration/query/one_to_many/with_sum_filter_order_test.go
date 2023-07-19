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

func TestOneToManyAscOrderAndFilterOnParentWithAggSumOnSubTypeField(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "1-N ascending order & filter on parent, with sum on on subtype field.",
		Request: `query {
			Author(order: {age: ASC}, filter: {age: {_gt: 8}}) {
				name
				_sum(published: {field: rating})
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
					"name": "The Associate",
					"rating": 4.2,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "Sooley",
					"rating": 3.2,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "The Rooster Bar",
					"rating": 4,
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
				`{
					"name": "Little Kid",
					"age": 6,
					"verified": true
				}`,
				`{
					"name": "Not a Writer",
					"age": 85,
					"verified": false
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name": "Cornelia Funke",
				"_sum": 4.8,
			},
			{
				"name": "John Grisham",
				"_sum": 20.799999999999997,
			},
			{
				"name": "Not a Writer",
				"_sum": 0.0,
			},
		},
	}

	executeTestCase(t, test)
}

func TestOneToManyDescOrderAndFilterOnParentWithAggSumOnSubTypeField(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "1-N descending order & filter on parent, with sum on on subtype field.",
		Request: `query {
			Author(order: {age: DESC}, filter: {age: {_gt: 8}}) {
				name
				_sum(published: {field: rating})
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
					"name": "The Associate",
					"rating": 4.2,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "Sooley",
					"rating": 3.2,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "The Rooster Bar",
					"rating": 4,
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
				`{
					"name": "Little Kid",
					"age": 6,
					"verified": true
				}`,
				`{
					"name": "Not a Writer",
					"age": 85,
					"verified": false
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name": "Not a Writer",
				"_sum": 0.0,
			},
			{
				"name": "John Grisham",
				"_sum": 20.799999999999997,
			},
			{
				"name": "Cornelia Funke",
				"_sum": 4.8,
			},
		},
	}

	executeTestCase(t, test)
}

func TestOnetoManySumBySubTypeFieldAndSumBySybTypeFieldWithDescOrderingOnFieldWithLimit(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "1-N sum subtype and sum subtype with desc. order on field with limit.",
		Request: `query {
			Author {
				name
				sum1: _sum(published: {field: rating})
				sum2: _sum(published: {field: rating, limit: 2, order: {rating: DESC}})
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
					"name": "The Associate",
					"rating": 4.2,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "Sooley",
					"rating": 3.2,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "The Rooster Bar",
					"rating": 4,
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
				`{
					"name": "Little Kid",
					"age": 6,
					"verified": true
				}`,
				`{
					"name": "Not a Writer",
					"age": 85,
					"verified": false
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name": "Little Kid",
				"sum1": 0.0,
				"sum2": 0.0,
			},
			{
				"name": "Not a Writer",
				"sum1": 0.0,
				"sum2": 0.0,
			},
			{
				"name": "John Grisham",
				"sum1": 20.799999999999997,
				"sum2": 4.9 + 4.5,
			},
			{
				"name": "Cornelia Funke",
				"sum1": 4.8,
				"sum2": 4.8,
			},
		},
	}

	executeTestCase(t, test)
}

func TestOnetoManySumBySubTypeFieldAndSumBySybTypeFieldWithAscOrderingOnFieldWithLimit(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "1-N sum subtype and sum subtype with asc. order on field with limit.",
		Request: `query {
			Author {
				name
				sum1: _sum(published: {field: rating})
				sum2: _sum(published: {field: rating, limit: 2, order: {rating: ASC}})
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
					"name": "The Associate",
					"rating": 4.2,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "Sooley",
					"rating": 3.2,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "The Rooster Bar",
					"rating": 4,
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
				`{
					"name": "Little Kid",
					"age": 6,
					"verified": true
				}`,
				`{
					"name": "Not a Writer",
					"age": 85,
					"verified": false
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name": "Little Kid",
				"sum1": 0.0,
				"sum2": 0.0,
			},
			{
				"name": "Not a Writer",
				"sum1": 0.0,
				"sum2": 0.0,
			},
			{
				"name": "John Grisham",
				"sum1": 20.799999999999997,
				"sum2": 4.0 + 3.2,
			},
			{
				"name": "Cornelia Funke",
				"sum1": 4.8,
				"sum2": 4.8,
			},
		},
	}

	executeTestCase(t, test)
}

func TestOneToManyLimitAscOrderSumOfSubTypeAndLimitAscOrderFieldsOfSubtype(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "1-N sum of subtype float field with limit and asc. order, and non-sum query of same subtype fields.",
		Request: `query {
			Author {
				LimitOrderSum: _sum(published: {field: rating, limit: 2, order: {rating: ASC}})
				LimitOrderFields: published(order: {rating: ASC}, limit: 2) {
					name
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
					"name": "The Associate",
					"rating": 4.2,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "Sooley",
					"rating": 3.2,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "The Rooster Bar",
					"rating": 4,
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
				`{
					"name": "Little Kid",
					"age": 6,
					"verified": true
				}`,
				`{
					"name": "Not a Writer",
					"age": 85,
					"verified": false
				}`,
			},
		},
		Results: []map[string]any{
			{
				"LimitOrderSum":    0.0,
				"LimitOrderFields": []map[string]any{},
			},
			{
				"LimitOrderSum":    0.0,
				"LimitOrderFields": []map[string]any{},
			},
			{
				"LimitOrderSum": 3.2 + 4.0,
				"LimitOrderFields": []map[string]any{
					{
						"name": "Sooley",
					},
					{
						"name": "The Rooster Bar",
					},
				},
			},
			{
				"LimitOrderSum": 4.8,
				"LimitOrderFields": []map[string]any{
					{
						"name": "Theif Lord",
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestOneToManyLimitDescOrderSumOfSubTypeAndLimitAscOrderFieldsOfSubtype(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "1-N sum of subtype float field with limit and desc. order, and non-sum query of same subtype fields.",
		Request: `query {
			Author {
				LimitOrderSum: _sum(published: {field: rating, limit: 2, order: {rating: DESC}})
				LimitOrderFields: published(order: {rating: DESC}, limit: 2) {
					name
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
					"name": "The Associate",
					"rating": 4.2,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "Sooley",
					"rating": 3.2,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "The Rooster Bar",
					"rating": 4,
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
				`{
					"name": "Little Kid",
					"age": 6,
					"verified": true
				}`,
				`{
					"name": "Not a Writer",
					"age": 85,
					"verified": false
				}`,
			},
		},
		Results: []map[string]any{
			{
				"LimitOrderSum":    0.0,
				"LimitOrderFields": []map[string]any{},
			},
			{
				"LimitOrderSum":    0.0,
				"LimitOrderFields": []map[string]any{},
			},
			{
				"LimitOrderSum": 4.9 + 4.5,
				"LimitOrderFields": []map[string]any{
					{
						"name": "Painted House",
					},
					{
						"name": "A Time for Mercy",
					},
				},
			},
			{
				"LimitOrderSum": 4.8,
				"LimitOrderFields": []map[string]any{
					{
						"name": "Theif Lord",
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
