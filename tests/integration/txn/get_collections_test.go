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
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

// This test runs GetCollections inside of a transaction with AddCollection, and illustrates that
// the collections are seen by the action.
func TestTxn_GetCollections_InsideTxnWithAddSchema_Succeeds(t *testing.T) {
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
			&action.GetCollections{
				TransactionID: immutable.Some(1),
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

// This test runs GetCollections inside of a transaction separate from the one that AddCollection is run in,
// and illustrates that the collections are not seen by the action.
func TestTxn_GetCollections_InsideTxnWithoutAddSchema_NoCollections(t *testing.T) {
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
				TransactionID:   immutable.Some(2),
				ExpectedResults: []client.CollectionVersion{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
