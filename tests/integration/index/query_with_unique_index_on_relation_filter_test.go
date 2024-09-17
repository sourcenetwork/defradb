// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package index

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryWithUniqueCompositeIndex_WithFilterOnIndexedRelation_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						devices: [Device]
					}

					type Device  {
						manufacturer: String 
						owner: User @index(unique: true, includes: [{name: "manufacturer"}])
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"manufacturer": "Apple",
					"owner_id":     testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.Request{
				Request: `query {
					User {
						name
						devices(filter: {owner_id: {_eq: "bae-1ef746f8-821e-586f-99b2-4cb1fb9b782f"}}) {
							manufacturer
						}
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
							"devices": []map[string]any{
								{"manufacturer": "Apple"},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
