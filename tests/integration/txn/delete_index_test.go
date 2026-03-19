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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

// This test runs DeleteIndex inside of a transaction, and illustrates that committing the transaction
// results in the index being created.
func TestTxn_DeleteIndex_WithCommit_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String @index
					}
				`,
			},
			&action.DeleteIndex{
				TransactionID: immutable.Some(1),
				IndexName:     "User_name_ASC",
			},
			testUtils.CommitTransaction{
				TransactionID: 1,
			},
			&action.ListIndexes{
				ExpectedIndexes: []client.IndexDescription{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test runs DeleteIndex inside of a transaction, and illustrates that not committing the transaction
// results in the index still existing.
func TestTxn_DeleteIndex_WithoutCommit_DoesNotDelete(t *testing.T) {
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
					type User {
						name: String @index
					}
				`,
			},
			&action.DeleteIndex{
				TransactionID: immutable.Some(1),
				IndexName:     "User_name_ASC",
			},
			&action.ListIndexes{
				ExpectedIndexes: []client.IndexDescription{
					{
						Name: "User_name_ASC",
						ID:   1,
						Fields: []client.IndexedFieldDescription{
							{
								Name: "name",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test runs DeleteIndex inside of a transaction, and illustrates that transactional isolation
// is maintained, and it can see indexes on schemas created in the same transaction.
func TestTxn_DeleteIndex_ExhibitsTransactionalIsolation_Succeeds(t *testing.T) {
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
				TransactionID: immutable.Some(1),
				SDL: `
					type User {
						name: String @index
					}
				`,
			},
			&action.DeleteIndex{
				TransactionID: immutable.Some(1),
				IndexName:     "User_name_ASC",
			},
			&action.ListIndexes{
				TransactionID:   immutable.Some(1),
				ExpectedIndexes: []client.IndexDescription{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
