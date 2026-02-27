// Copyright 2025 Democratized Data Foundation
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
	"github.com/sourcenetwork/immutable"
)

// This test runs AddSchema inside of a transaction, and illustrates that committing the transaction
// results in the schema being added to the database.
func TestTxn_AddSchema_WithCommit_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				TransactionID: immutable.Some(1),
				Schema: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			testUtils.TransactionCommit{
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
// results in the schema not yet being added to the database.
func TestTxn_AddSchema_WithoutCommit_EmptyResults(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				TransactionID: immutable.Some(1),
				Schema: `
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
