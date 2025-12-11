// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package create

import (
	"testing"
	"time"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestMutationWithTxnDoesNotAllowUpdateInSecondTransactionUser(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some([]state.ClientType{
			// TTL only supported on HTTP/CLI clients
			// state.CLIClientType,
			state.HTTPClientType,
		}),
		HTTPTxnTTLCache: testUtils.HTTPTxnTTLCache{
			TxnTTL: 1 * time.Second,
		},
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			testUtils.Request{
				TransactionID: immutable.Some(1),
				Request: `mutation {
					update_User(input: {age: 28}) {
						_docID
						name
						age
					}
				}`,
				Results: map[string]any{
					"update_User": []map[string]any{
						{
							"_docID": "bae-bb8ed746-4570-5651-ac69-39a21f733211",
							"name":   "John",
							"age":    int64(28),
						},
					},
				},
			},
			// // Wiat for the txn to expire
			testUtils.Wait{
				Duration: 2 * time.Second,
			},
			testUtils.Request{
				TransactionID: immutable.Some(1),
				Request: `mutation {
					update_User(input: {age: 29}) {
						_docID
						name
						age
					}
				}`,
				ExpectedError: "missing or expired transaction",
			},
			testUtils.Request{
				// Regular query after transaction has expired:
				Request: `query {
					User {
						_docID
						name
						age
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"_docID": "bae-bb8ed746-4570-5651-ac69-39a21f733211",
							"name":   "John",
							"age":    int64(27),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
