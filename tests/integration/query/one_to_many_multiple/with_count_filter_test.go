// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_many_multiple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToManyMultipleWithCountOnMultipleJoinsWithAndWithoutFilter(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "One-to-many relation query from many side with counts with and without filters",
		Query: `query {
				author {
					name
					_count(books: {filter: {score: {_gt: 3}}}, articles: {})
				}
			}`,
		Docs: map[int][]string{
			//articles
			0: {
				(`{
					"name": "After Guantánamo, Another Injustice",
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3",
					"rating": 3
				}`),
				(`{
					"name": "To my dear readers",
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04",
					"rating": 2
				}`),
				(`{
					"name": "Twinklestar's Favourite Xmas Cookie",
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04",
					"rating": 1
				}`),
			},
			//books
			1: {
				(`{
					"name": "Painted House",
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3",
					"score": 1
				}`),
				(`{
					"name": "A Time for Mercy",
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3",
					"score": 2
				}`),
				(`{
					"name": "Sooley",
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3",
					"score": 3
				}`),
				(`{
					"name": "Theif Lord",
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04",
					"score": 4
				}`),
			},
			//authors
			2: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				(`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`),
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				(`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`),
			},
		},
		Results: []map[string]interface{}{
			{
				"name":   "John Grisham",
				"_count": 1,
			},
			{
				"name":   "Cornelia Funke",
				"_count": 3,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToManyMultipleWithCountOnMultipleJoinsWithFilters(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "One-to-many relation query from many side with counts with filters",
		Query: `query {
				author {
					name
					_count(books: {filter: {score: {_gt: 3}}}, articles: {filter: {rating: {_lt: 3}}})
				}
			}`,
		Docs: map[int][]string{
			//articles
			0: {
				(`{
					"name": "After Guantánamo, Another Injustice",
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3",
					"rating": 3
				}`),
				(`{
					"name": "To my dear readers",
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04",
					"rating": 2
				}`),
				(`{
					"name": "Twinklestar's Favourite Xmas Cookie",
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04",
					"rating": 1
				}`),
			},
			//books
			1: {
				(`{
					"name": "Painted House",
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3",
					"score": 1
				}`),
				(`{
					"name": "A Time for Mercy",
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3",
					"score": 2
				}`),
				(`{
					"name": "Sooley",
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3",
					"score": 3
				}`),
				(`{
					"name": "Theif Lord",
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04",
					"score": 4
				}`),
			},
			//authors
			2: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				(`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`),
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				(`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`),
			},
		},
		Results: []map[string]interface{}{
			{
				"name":   "John Grisham",
				"_count": 0,
			},
			{
				"name":   "Cornelia Funke",
				"_count": 3,
			},
		},
	}

	executeTestCase(t, test)
}
