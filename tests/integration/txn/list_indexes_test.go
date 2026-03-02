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

// This test runs ListIndexes inside of a transaction with AddSchema, and illustrates that
// the indexes are seen by the action.
func TestTxn_ListIndexes_InsideTxnWithAddSchema_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				TransactionID: immutable.Some(1),
				Schema: `
					type User {
						name: String @index
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
			&action.ListIndexes{
				TransactionID: immutable.Some(1),
				CollectionID:  0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name: "User_name_ASC",
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
