// Copyright 2026 Democratized Data Foundation
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
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

// This test runs Request inside of a transaction, and illustrates that committing the transaction
// results in the mutation adding a document to the database.
func TestTxn_Request_WithCommit_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			&action.Request{
				TransactionID: immutable.Some(1),
				Request: `
					mutation {
						add_Users(input: [
							{ name: "John", age: 27 }
						]) {
							_docID
						}
					}
				`,
				Results: map[string]any{
					"add_Users": []map[string]any{
						{
							"_docID": "bae-32e84498-d467-5f01-b93e-fc2dca59be76",
						},
					},
				},
			},
			testUtils.CommitTransaction{
				TransactionID: 1,
			},
			&action.Request{
				Request: `
					query {
						Users {
							_docID
							name
							age
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_docID": "bae-32e84498-d467-5f01-b93e-fc2dca59be76",
							"name":   "John",
							"age":    int64(27),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test runs Request inside of a transaction, and illustrates that not committing the transaction
// results in the document not yet being in the database.
func TestTxn_Request_WithoutCommit_EmptyResults(t *testing.T) {
	test := testUtils.TestCase{
		// LevelDB does not support concurrent transactions
		// todo: https://github.com/sourcenetwork/defradb/issues/4442
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			&action.Request{
				TransactionID: immutable.Some(1),
				Request: `
					mutation {
						add_Users(input: [
							{ name: "John", age: 27 }
						]) {
							_docID
						}
					}
				`,
				Results: map[string]any{
					"add_Users": []map[string]any{
						{
							"_docID": "bae-32e84498-d467-5f01-b93e-fc2dca59be76",
						},
					},
				},
			},
			&action.Request{
				Request: `
					query {
						Users {
							_docID
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
