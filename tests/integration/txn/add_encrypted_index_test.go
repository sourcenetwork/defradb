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

// This test runs AddEncryptedIndex inside of a transaction, and illustrates that committing the transaction
// results in the index being created.
func TestTxn_AddEncryptedIndex_WithCommit_Succeeds(t *testing.T) {
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
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.AddEncryptedIndex{
				TransactionID: immutable.Some(1),
				FieldName:     "name",
			},
			testUtils.TransactionCommit{
				TransactionID: 1,
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

// This test runs AddEncryptedIndex inside of a transaction, and illustrates that not committing the transaction
// results in the index not yet being created.
func TestTxn_AddEncryptedIndex_WithoutCommit_NoIndexes(t *testing.T) {
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
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.AddEncryptedIndex{
				TransactionID: immutable.Some(1),
				FieldName:     "name",
			},
			testUtils.ListEncryptedIndexes{
				CollectionID:    0,
				ExpectedIndexes: []client.EncryptedIndexDescription{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test runs ListEncryptedIndexes inside of a transaction with Add, and illustrates that
// the indexes are seen by the action.
func TestTxn_ListEncryptedIndexes_InsideTxn_Succeeds(t *testing.T) {
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
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.AddEncryptedIndex{
				TransactionID: immutable.Some(1),
				FieldName:     "name",
			},
			testUtils.ListEncryptedIndexes{
				TransactionID: immutable.Some(1),
				CollectionID:  0,
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
