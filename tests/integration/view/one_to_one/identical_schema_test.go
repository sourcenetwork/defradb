// Copyright 2023 Democratized Data Foundation
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

func TestView_OneToOneSameSchema(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one view with same schema",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type LeftHand {
						name: String
						holding: RightHand @primary @relation(name: "left_right")
						heldBy: RightHand @relation(name: "right_left")
					}
					type RightHand {
						name: String
						holding: LeftHand @primary @relation(name: "right_left")
						heldBy: LeftHand @relation(name: "left_right")
					}
				`,
			},
			testUtils.CreateView{
				Query: `
					LeftHand {
						name
						heldBy {
							name
						}
					}
				`,
				// todo - such a setup appears to work, yet prevents the querying of `RightHand`s as the primary return object
				// thought - although, perhaps if the view is defined as such, Left and right hands *could* be merged by us into a single table
				SDL: `
					type HandView {
						name: String
						holding: HandView @primary
						heldBy: HandView
					}
				`,
			},
			// bae-f3db7a4d-3db1-5d57-9996-32c3fdff99d3
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"Left hand 1"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name":	"Right hand 1",
					"holding_id": "bae-f3db7a4d-3db1-5d57-9996-32c3fdff99d3"
				}`,
			},
			testUtils.Request{
				Request: `query {
							HandView {
								name
								heldBy {
									name
								}
							}
						}`,
				Results: []map[string]any{
					{
						"name": "Left hand 1",
						"heldBy": map[string]any{
							"name": "Right hand 1",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
