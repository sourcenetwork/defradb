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
	"github.com/sourcenetwork/immutable"
)

// This test runs RefreshViews inside of a transaction, and illustrates that committing the transaction
// results in the view being refreshed.
func TestTxn_RefreshView_WithCommit_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		SupportedViewTypes: immutable.Some([]testUtils.ViewType{
			testUtils.MaterializedViewType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			&action.AddView{
				Query: `
					User {
						name
					}
				`,
				SDL: `
					type UserView {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name":	"Fred"
				}`,
			},
			&action.RefreshViews{
				TransactionID: immutable.Some(1),
			},
			&action.CommitTransaction{
				TransactionID: 1,
			},
			&action.Request{
				DoNotRefreshViews: true,
				NonOrderedResults: true,
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
						{
							"name": "Fred",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
