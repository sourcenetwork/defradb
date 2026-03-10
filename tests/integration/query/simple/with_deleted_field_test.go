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

func TestQuerySimple_WithDeletedField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Andy"
				}`,
			},
			testUtils.DeleteDoc{
				DocID: 0,
			},
			testUtils.DeleteDoc{
				DocID: 1,
			},
			&action.Request{
				Request: `query {
						User(showDeleted: true) {
							_deleted
							name
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"_deleted": true,
							"name":     "John",
						},
						{
							"_deleted": true,
							"name":     "Andy",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
