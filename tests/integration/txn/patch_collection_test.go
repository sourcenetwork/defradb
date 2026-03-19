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

// This test runs PatchCollection inside of a transaction, and illustrates that commiting the transaction
// results in the patch being applied.
func TestTxn_PatchCollection_WithCommit_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
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
