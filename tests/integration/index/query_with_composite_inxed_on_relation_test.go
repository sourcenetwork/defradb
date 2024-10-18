// Copyright 2024 Democratized Data Foundation
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

// This test was added during https://github.com/sourcenetwork/defradb/issues/2924
// The issue was that [multiScanNode] that keeps track of how many calls to [Next()] should
// be made, would call [Init()] and [Start()] every time without tracking, which would result
// in child nodes being initialized and started multiple times, which in turn created
// index iterators without closing them.
func TestQueryWithCompositeIndexOnManyToOne_WithMultipleIndexedChildNodes_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						devices: [Device]
					}

					type Device @index(includes: [{field: "owner_id"}, {field: "manufacturer_id"}]) {
						model: String
						owner: User 
						manufacturer: Manufacturer 
					}

					type Manufacturer {
						name: String
						devices: [Device]
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name": "Apple",
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name": "Sony",
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "MacBook Pro",
					"owner":        testUtils.NewDocIndex(0, 0),
					"manufacturer": testUtils.NewDocIndex(2, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"model":        "PlayStation 5",
					"owner":        testUtils.NewDocIndex(0, 0),
					"manufacturer": testUtils.NewDocIndex(2, 1),
				},
			},
			testUtils.Request{
				Request: `query {
					User {
						devices {
							model
							owner {
								name
							}
							manufacturer {
								name
							}
						}
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"devices": []map[string]any{
								{
									"model":        "MacBook Pro",
									"owner":        map[string]any{"name": "John"},
									"manufacturer": map[string]any{"name": "Apple"},
								},
								{
									"model":        "PlayStation 5",
									"owner":        map[string]any{"name": "John"},
									"manufacturer": map[string]any{"name": "Sony"},
								},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
