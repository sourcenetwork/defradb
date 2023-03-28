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

func TestQueryOneToTwoManyWithOrder(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "One-to-many relation query from one side, order in opposite directions on children",
			Request: `query {
						Author {
							name
							written (order: {rating: ASC}) {
								name
							}
							reviewed (order: {rating: DESC}){
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
						"reviewedBy_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
					}`,
					`{
						"name": "A Time for Mercy",
						"rating": 4.5,
						"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3",
						"reviewedBy_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
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
					"reviewed": []map[string]any{
						{
							"name":   "Theif Lord",
							"rating": 4.8,
						},
						{
							"name":   "A Time for Mercy",
							"rating": 4.5,
						},
					},
					"written": []map[string]any{
						{
							"name": "A Time for Mercy",
						},
						{
							"name": "Painted House",
						},
					},
				},
				{
					"name": "Cornelia Funke",
					"reviewed": []map[string]any{
						{
							"name":   "Painted House",
							"rating": 4.9,
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
