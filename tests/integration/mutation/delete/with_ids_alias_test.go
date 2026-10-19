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

func TestMutationDeletion_WithIDsAndSelectAlias(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int
						points: Float
						verified: Boolean
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad",
					"age":  26,
					"points": 48.48,
					"verified": true
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age":  26,
					"points": 48.48,
					"verified": true
				}`,
			},
			&action.Request{
				Request: `mutation {
					delete_User(docID: ["{{.DocID0_0}}", "{{.DocID0_1}}"]) {
						AliasID: _docID
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{
						{
							"AliasID": "{{.DocID0_0}}",
						},
						{
							"AliasID": "{{.DocID0_1}}",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
