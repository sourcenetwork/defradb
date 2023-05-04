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

func TestQueryOneToManyWithChildDocKey(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many relation query from one side with child dockey",
		Request: `query {
					Author {
						name
						published (
								dockey: "bae-b9b83269-1f28-5c3b-ae75-3fb4c00d559d"
							) {
							name
						}
					}
				}`,
		Docs: map[int][]string{
			//books
			0: { // bae-b9b83269-1f28-5c3b-ae75-3fb4c00d559d
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
				"name": "John Grisham",
				"published": []map[string]any{
					{
						"name": "Painted House",
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
