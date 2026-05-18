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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/multiplier"
	"github.com/sourcenetwork/defradb/tests/state"
)

// This test runs PatchCollection inside of a transaction, and illustrates that commiting the transaction
// results in the patch being applied.
func TestTxn_PatchCollection_WithCommit_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		// The secondary-index multiplier adds @index on all fields. Removing a field that
		// has a dependent index fails with "the given field does not exist".
		// https://github.com/sourcenetwork/defradb/issues/4722
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						email: String
					}
				`,
			},
			&action.PatchCollection{
				TransactionID: immutable.Some(1),
				Patch: `
					[
						{ "op": "remove", "path": "/Users/Fields/2" }
					]
				`,
			},
			&action.CommitTransaction{
				TransactionID: 1,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						email
					}
				}`,
				ExpectedError: "Cannot query field \"name\" on type \"Users\".",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test runs PatchCollection inside of a transaction, and illustrates that notcommiting the transaction
// results in the patch not yet being applied.
func TestTxn_PatchCollection_WithoutCommit_PatchNotApplied(t *testing.T) {
	test := testUtils.TestCase{
		// The secondary-index multiplier adds @index on all fields. Removing a field that
		// has a dependent index fails with "the given field does not exist".
		// https://github.com/sourcenetwork/defradb/issues/4722
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
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
						email: String
					}
				`,
			},
			&action.PatchCollection{
				TransactionID: immutable.Some(1),
				Patch: `
					[
						{ "op": "remove", "path": "/Users/Fields/2" }
					]
				`,
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
