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

// This test runs AddEncryptedIndex inside of a transaction, and illustrates that transactional isolation
// is maintained, and that it can see indexes on schemas created in the same transaction.
func TestTxn_AddEncryptedIndex_ExhibitsTransactionalIsolation_Succeeds(t *testing.T) {
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
				CollectionID:  0,
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
