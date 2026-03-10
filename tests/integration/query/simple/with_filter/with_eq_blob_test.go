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

func TestQuerySimple_WithEqOpOnBlobField_ShouldFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						data: Blob
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"data": "00FF99AA"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Andy",
					"data": "FA02CC45"
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {data: {_eq: "00FF99AA"}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "John"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
