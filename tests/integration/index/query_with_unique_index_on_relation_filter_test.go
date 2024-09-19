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

func TestQueryWithUniqueCompositeIndex_WithIndexComprising2RelationsAndFilterOnIt_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test if we can filter on a unique composite index comprising at least 2 relations and filter on them",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String @index(unique: true)
						devices: [Device]
					}

					type Manufacturer {
						name: String
						devices: [Device]
					}
					
					type Device @index(unique: true, includes: [{name: "owner_id"}, {name: "manufacturer_id"}, {name: "model"}]) {
						owner: User 
						manufacturer: Manufacturer 
						model: String
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
					"name": "Apple",
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"model":           "iPhone",
					"owner_id":        testUtils.NewDocIndex(0, 0),
					"manufacturer_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"model":           "MacBook",
					"owner_id":        testUtils.NewDocIndex(0, 0),
					"manufacturer_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.Request{
				Request: `query {
					byUserId: Device (filter: {
						manufacturer_id: {_eq: "bae-18c7d707-c44d-552f-b6d6-9e3d05bbf9c1"},
						owner_id: {_eq: "bae-1ef746f8-821e-586f-99b2-4cb1fb9b782f"}
					}) {
						owner {
							name
						}
					}
					byUserName: Device (filter: {
						manufacturer_id: {_eq: "bae-18c7d707-c44d-552f-b6d6-9e3d05bbf9c1"},
						owner: {name: {_eq: "John"}}
					}) {
						owner {
							name
						}
					}
				}`,
				Results: map[string]any{
					"byUserId": []map[string]any{
						{
							"owner": map[string]any{
								"name": "John",
							},
						},
						{
							"owner": map[string]any{
								"name": "John",
							},
						},
					},
					"byUserName": []map[string]any{
						{
							"owner": map[string]any{
								"name": "John",
							},
						},
						{
							"owner": map[string]any{
								"name": "John",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
