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

package update

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdate_WithIdAndVersion_ReturnResults(t *testing.T) {
	updateCid := testUtils.NewSameValue()
	createCid := testUtils.NewSameValue()

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						points: Float
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"points": 42.1
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Bob",
					"points": 66.6
				}`,
			},
			&action.Request{
				Request: `mutation {
					update_Users(docID: "bae-9466cfe3-c011-5d44-b1cd-f0c5a46d9202", input: {points: 59}) {
						name
						_version {
							cid
						}
					}
				}`,
				Results: map[string]any{
					"update_Users": []map[string]any{
						{
							"name": "John",
							"_version": []map[string]any{
								{
									"cid": updateCid,
								},
								{
									"cid": createCid,
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
