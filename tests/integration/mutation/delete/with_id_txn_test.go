// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

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
			&action.AddSchema{
				Schema: `
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
					delete_User(docID: ["bae-390b4419-fe1c-506b-98bd-20847cdab2d9"]) {
						_docID
					}
				}`,
				Results: map[string]any{
					"delete_User": []map[string]any{
						{
							"_docID": "bae-390b4419-fe1c-506b-98bd-20847cdab2d9",
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
