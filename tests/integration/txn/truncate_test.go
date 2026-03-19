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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/immutable"
)

// This test runs Truncate inside of a transaction, and illustrates that committing the transaction
// results in the truncation occurring.
func TestCollectionTruncate_WithCommit_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Truncate{
				CollectionIndex: 0,
				TransactionID:   immutable.Some(1),
			},
			testUtils.CommitTransaction{
				TransactionID: 1,
			},
			&action.Request{
				Request: `query {
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Note there is no test for the case of the commit not happening. This is because
// datastore read-locks are only released when the transaction closes. This is unique to
// the Truncate action.
