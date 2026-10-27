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

package update

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/immutable"
)

// This test covers the scenario outlined in:
// https://github.com/sourcenetwork/defradb/issues/5257
// where the document fetcher was reading beyond the documents
// that mattered, causing unnessecary transaction conflicts.
func TestMutationUpdate_InterleavedTxn(t *testing.T) {
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
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Fred",
				},
			},
			&action.UpdateDoc{
				DocID:         0,
				TransactionID: immutable.Some(0),
				Doc: `{
					"name": "Johnny"
				}`,
			},
			&action.UpdateDoc{
				DocID:         1,
				TransactionID: immutable.Some(1),
				Doc: `{
					"name": "Freddy"
				}`,
			},
			&action.CommitTransaction{
				TransactionID: 0,
			},
			&action.CommitTransaction{
				TransactionID: 1,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
