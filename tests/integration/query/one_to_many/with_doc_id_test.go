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

func TestQueryOneToManyWithChildDocID(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many relation query from one side with child docID",
		Request: `query {
					Author {
						name
						published (
								docID: "bae-5366ba09-54e8-5381-8169-a770aa9282ae"
							) {
							name
						}
					}
				}`,
		Docs: map[int][]string{
			//books
			0: { // bae-5366ba09-54e8-5381-8169-a770aa9282ae
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
