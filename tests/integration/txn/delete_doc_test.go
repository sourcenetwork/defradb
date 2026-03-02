// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package txn_testing

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/immutable"
)

// This test runs DeleteDoc inside of a transaction, and illustrates that committing the transaction
// results in the document being deleted from the database.
func TestTxn_DeleteDoc_WithCommit_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			testUtils.DeleteDoc{
				TransactionID: immutable.Some(1),
				CollectionID:  0,
				DocID:         0,
			},
			testUtils.TransactionCommit{
				TransactionID: 1,
			},
			&action.Request{
				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test runs DeleteDoc inside of a transaction, and illustrates that not committing the transaction
// results in the document not yet being deleted from the database.
func TestTxn_DeleteDoc_WithoutCommit_DoesNotDelete(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			testUtils.DeleteDoc{
				TransactionID: immutable.Some(1),
				DocID:         0,
			},
			&action.Request{
				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(27),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
