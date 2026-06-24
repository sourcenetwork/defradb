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
	"time"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestTxnTTL_GivenIdleHTTPTransaction_DiscardsTransaction(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some([]state.ClientType{
			state.HTTPClientType,
			state.CLIClientType,
		}),
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
		}),
		HTTP: immutable.Some(options.NodeHTTPOptions{
			TxnTTL:        120 * time.Millisecond,
			TxnTTLTick:    20 * time.Millisecond,
			TxnTTLBuckets: 10,
		}),
		SkipChangeDetector: true,
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			&action.Request{
				TransactionID: immutable.Some(1),
				Request: `mutation {
					update_Users(filter: {name: {_eq: "John"}}, input: {age: 28}) {
						name
						age
					}
				}`,
				Results: map[string]any{
					"update_Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(28),
						},
					},
				},
			},
			&action.Wait{
				Duration: immutable.Some(300 * time.Millisecond),
			},
			&action.Request{
				TransactionID: immutable.Some(1),
				Request: `mutation {
					update_Users(filter: {name: {_eq: "John"}}, input: {age: 29}) {
						name
						age
					}
				}`,
				ExpectedError: "missing or expired transaction",
			},
			&action.Request{
				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"age":  int64(27),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
