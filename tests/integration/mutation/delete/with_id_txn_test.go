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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationDeletion_WithIDAndTxn(t *testing.T) {
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
				TransactionID: immutable.Some(0),
				Request: `mutation {
					delete_User(docID: ["{{.DocID0_0}}"]) {
						_docID
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{
						{
							"_docID": "{{.DocID0_0}}",
						},
					},
				},
			},
			&action.Request{
				TransactionID: immutable.Some(0),
				Request: `query {
					User {
						_docID
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
