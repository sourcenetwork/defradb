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

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithVersionAndOrder(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Chris",
				},
			},
			&action.Request{
				Request: `query {
					Users(order: {name: ASC}) {
						name
						_version {
							fieldName
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Chris",
							"_version": []map[string]any{
								{
									"fieldName": "_C",
								},
							},
						},
						{
							"name": "John",
							"_version": []map[string]any{
								{
									"fieldName": "_C",
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
