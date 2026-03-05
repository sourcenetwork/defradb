// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package index

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryWithUniqueCompositeIndex_WithFilterOnIndexedRelation_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						devices: [Device]
					}

					type Device  {
						manufacturer: String 
						owner: User @index(unique: true, includes: [{field: "manufacturer"}])
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"manufacturer": "Apple",
					"_ownerID":     testUtils.NewDocIndex(0, 0),
				},
			},
			&action.Request{
				Request: `query {
					User {
						name
						devices(filter: {_ownerID: {_eq: "bae-7f4197fe-c647-5cc6-91bb-5f32229fd4cd"}}) {
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
