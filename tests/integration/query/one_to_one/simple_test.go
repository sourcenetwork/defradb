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

func TestQueryOneToOne(t *testing.T) {
	tests := []testUtils.QueryTestCase{
		{
			Description: "One-to-one relation query with no filter",
			Query: `query {
						book {
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
					"rating": 4.9
				}`,
				},
				//authors
				1: { // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
					`{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
				}`,
				},
			},
			Results: []map[string]interface{}{
				{
					"name":   "Painted House",
					"rating": 4.9,
					"author": map[string]interface{}{
						"name": "John Grisham",
						"age":  uint64(65),
					},
				},
			},
		},
		{
			Description: "One-to-one relation secondary direction, no filter",
			Query: `query {
						author {
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
					"rating": 4.9
					}`,
				},
				//authors
				1: { // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
					`{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
					}`,
				},
			},
			Results: []map[string]interface{}{
				{
					"name": "John Grisham",
					"age":  uint64(65),
					"published": map[string]interface{}{
						"name":   "Painted House",
						"rating": 4.9,
					},
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}
