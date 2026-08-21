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

package truncate

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestTruncateCollection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.Truncate{
				CollectionIndex: 0,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestTruncateCollection_TruncateTwice(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.Truncate{},
			&action.Truncate{},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestTruncateCollection_WithFilter(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { name: String }`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap:       map[string]any{"name": "John"},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap:       map[string]any{"name": "Jane"},
			},
			&action.Truncate{
				CollectionIndex: 0,
				DocIndexes:      []int{0},
				PruneHistory:    true,
			},
			&action.Truncate{
				CollectionIndex: 0,
				DocIndexes:      []int{0},
				PruneHistory:    true,
			},
			&action.Request{
				Request: `query { Users { name } }`,
				Results: map[string]any{
					"Users": []map[string]any{{"name": "Jane"}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
