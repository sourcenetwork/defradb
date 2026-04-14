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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

// This test runs AddCollection inside of a transaction, and illustrates that committing the transaction
// results in the collection being added to the database.
func TestTxn_AddCollection_WithCommit_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				TransactionID: immutable.Some(1),
				SDL: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			testUtils.CommitTransaction{
				TransactionID: 1,
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections(),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						VersionID:      "bafyreihsneodeja4lfer5puptim3lkwvketyckrmkhfpgxm67ch5wenjwq",
						IsMaterialized: true,
						IsActive:       true,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test runs AddSchema inside of a transaction, and illustrates that not committing the transaction
// results in the collection not yet being added to the database.
func TestTxn_AddCollection_WithoutCommit_EmptyResults(t *testing.T) {
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
					type Users {
						name: String
						age: Int
					}
				`,
			},
			&action.GetCollections{
				FilterOptions:   options.GetCollections(),
				ExpectedResults: []client.CollectionVersion{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
