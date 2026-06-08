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

// This test runs AddView inside of a transaction, and illustrates that committing the transaction
// results in the view being usable.
func TestTxn_AddView_WithCommit_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.AddView{
				TransactionID: immutable.Some(1),
				Query: `
					User {
						name
					}
				`,
				SDL: `
					type UserView @materialized(if: false) {
						name: String
					}
				`,
			},
			&action.CommitTransaction{
				TransactionID: 1,
			},
			&action.AddDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			&action.Request{
				Request: `query {
							UserView {
								name
							}
						}`,
				Results: map[string]any{
					"UserView": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test runs AddView inside of a transaction, and illustrates that notcommitting the transaction
// results in the view notbeing usable.
func TestTxn_AddView_WithoutCommit_Fails(t *testing.T) {
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
					}
				`,
			},
			&action.AddView{
				TransactionID: immutable.Some(1),
				Query: `
					User {
						name
					}
				`,
				SDL: `
					type UserView @materialized(if: false) {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			&action.Request{
				Request: `query {
							UserView {
								name
							}
						}`,
				ExpectedError: "Cannot query field \"UserView\" on type \"Query\".",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
