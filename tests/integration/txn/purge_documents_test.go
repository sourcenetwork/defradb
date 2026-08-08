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
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

func TestTxn_PurgeDocuments_RespectsCommitAndDiscard(t *testing.T) {
	test := testUtils.TestCase{
		// LevelDB does not support concurrent transactions.
		// todo: https://github.com/sourcenetwork/defradb/issues/4442
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { name: String }`,
			},
			&action.AddDoc{
				Doc: `{"name":"John"}`,
			},
			&action.PurgeDocs{
				CollectionIndex: 0,
				DocIndexes:      []int{0},
				TransactionID:   immutable.Some(1),
			},
			&action.DiscardTransaction{
				TransactionID: 1,
			},
			&action.Request{
				Request: `query { Users { name } }`,
				Results: map[string]any{
					"Users": []map[string]any{{"name": "John"}},
				},
			},
			&action.PurgeDocs{
				CollectionIndex: 0,
				DocIndexes:      []int{0},
				TransactionID:   immutable.Some(2),
			},
			&action.CommitTransaction{
				TransactionID: 2,
			},
			&action.Request{
				Request: `query { Users { name } }`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
