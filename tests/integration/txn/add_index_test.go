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
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/immutable"
)

// This test runs AddIndex inside of a transaction, and illustrates that committing the transaction
// results in the index being created.
func TestTxn_AddIndex_WithCommit_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.AddIndex{
				TransactionID: immutable.Some(1),
				IndexName:     "some_index",
				FieldName:     "name",
			},
			testUtils.TransactionCommit{
				TransactionID: 1,
			},
			&action.ListIndexes{
				ExpectedIndexes: []client.IndexDescription{
					{
						Name: "some_index",
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

// This test runs AddIndex inside of a transaction, and illustrates that not committing the transaction
// results in the index not yet being created.
func TestTxn_AddIndex_WithoutCommit_NoIndexes(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String 
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.AddIndex{
				TransactionID: immutable.Some(1),
				IndexName:     "some_index",
				FieldName:     "name",
			},
			&action.ListIndexes{
				ExpectedIndexes: []client.IndexDescription{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test runs AddIndex inside of a transaction, and illustrates that transactional isolation
// is maintained, and it can see documents created in the same transaction.
func TestTxn_AddIndex_ExhibitsTransactionalIsolation_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				TransactionID: immutable.Some(1),
				Schema: `
					type User {
						name: String 
						age: Int
					}
				`,
			},
			&action.AddDoc{
				TransactionID: immutable.Some(1),
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.AddIndex{
				TransactionID: immutable.Some(1),
				IndexName:     "some_index",
				FieldName:     "name",
			},
			testUtils.TransactionCommit{
				TransactionID: 1,
			},
			&action.ListIndexes{
				ExpectedIndexes: []client.IndexDescription{
					{
						Name: "some_index",
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
