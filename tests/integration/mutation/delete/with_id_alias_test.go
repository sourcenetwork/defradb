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

package delete

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationDeletion_WithIDAndAlias(t *testing.T) {
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
					"name": "Shahzad"
				}`,
			},
			&action.Request{
				Request: `mutation {
					delete_User(docID: ["bae-390b4419-fe1c-506b-98bd-20847cdab2d9"]) {
						fancyKey: _docID
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{
						{
							"fancyKey": "bae-390b4419-fe1c-506b-98bd-20847cdab2d9",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
