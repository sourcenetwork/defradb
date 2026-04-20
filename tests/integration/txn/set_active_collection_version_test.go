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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

// This test runs SetActiveCollectionVersion inside of a transaction, and illustrates that commiting
// the transaction results in the version being changed.
func TestTxn_SetActiveCollectionVersion_WithCommit_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.SetActiveCollectionVersion{
				TransactionID: immutable.Some(1),
				VersionID:     "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu",
			},
			testUtils.CommitTransaction{
				TransactionID: 1,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						email
					}
				}`,
				ExpectedError: `Cannot query field "email" on type "Users".`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// This test runs SetActiveCollectionVersion inside of a transaction, and illustrates that not commiting
// the transaction results in the version not yet being changed.
func TestTxn_SetActiveCollectionVersion_WithoutCommit_VersionNotChanged(t *testing.T) {
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
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.SetActiveCollectionVersion{
				TransactionID: immutable.Some(1),
				VersionID:     "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu",
			},

			&action.Request{
				Request: `query {
					Users {
						name
						email
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
