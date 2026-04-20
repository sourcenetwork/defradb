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

// This test runs DeleteEncryptedIndex inside of a transaction, and illustrates that committing the transaction
// results in the index being deleted.
func TestTxn_DeleteEncryptedIndex_WithCommit_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.NewEncryptedIndex{
				FieldName: "name",
			},
			testUtils.DeleteEncryptedIndex{
				TransactionID: immutable.Some(1),
				FieldName:     "name",
			},
			testUtils.CommitTransaction{
				TransactionID: 1,
			},
			testUtils.ListEncryptedIndexes{
				CollectionID:    0,
				ExpectedIndexes: []client.EncryptedIndexDescription{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test runs DeleteEncryptedIndex inside of a transaction, and illustrates that not committing the transaction
// results in the index not yet beingdeleted.
func TestTxn_DeleteEncryptedIndex_WithoutCommit_DoesNotDelete(t *testing.T) {
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
						name: String 
						age: Int
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.NewEncryptedIndex{
				FieldName: "name",
			},
			testUtils.DeleteEncryptedIndex{
				TransactionID: immutable.Some(1),
				FieldName:     "name",
			},
			testUtils.ListEncryptedIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.EncryptedIndexDescription{
					{
						FieldName: "name",
						Type:      client.EncryptedIndexTypeEquality,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test runs DeleteEncryptedIndex inside of a transaction, and illustrates that transactional isolation
// is maintained, and it can see indexes created in the same transaction.
func TestTxn_DeleteEncryptedIndex_ExhibitsTransactionalIsolation_Succeeds(t *testing.T) {
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
						name: String 
						age: Int
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.NewEncryptedIndex{
				TransactionID: immutable.Some(1),
				FieldName:     "name",
			},
			testUtils.DeleteEncryptedIndex{
				TransactionID: immutable.Some(1),
				FieldName:     "name",
			},
			testUtils.ListEncryptedIndexes{
				TransactionID:   immutable.Some(1),
				CollectionID:    0,
				ExpectedIndexes: []client.EncryptedIndexDescription{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
