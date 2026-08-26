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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestTruncateCollection_WithFilterIsIdempotent(t *testing.T) {
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
			// Retrying the same filtered truncate is a no-op.
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

func TestTruncateCollection_WithLargeIntegerFilter_RemovesOnlyExactMatch(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				state.GoClientType,
				state.HTTPClientType,
				state.CLIClientType,
				state.CClientType,
			},
		),
		SupportedMutationTypes: immutable.Some(
			[]state.MutationType{
				state.CollectionSaveMutationType,
				state.CollectionNamedMutationType,
			},
		),
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { name: String, age: Int }`,
			},
			&action.AddDoc{Doc: `{"name":"Fred","age":9007199254740992}`},
			&action.AddDoc{Doc: `{"name":"John","age":9007199254740993}`},
			&action.Truncate{
				Filter: map[string]any{
					"age": map[string]any{"_eq": int64(9007199254740993)},
				},
			},
			&action.Request{
				Request: `query { Users { name } }`,
				Results: map[string]any{
					"Users": []map[string]any{{"name": "Fred"}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
