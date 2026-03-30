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
