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

// This file guards against regression of issue #4907: creating or deleting a secondary
// index on a collection whose existing documents exceed the storage engine's transaction
// size limit must succeed, which requires batched transactions for both operations.

package index

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	"github.com/sourcenetwork/defradb/tests/gen"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// largeDocCount × largeValuePadLen gives ~12.5 MB of data. Index keys embed the field value,
// so neither the backfill nor the deletion can fit in one ~11 MB transaction, while each key
// stays below the ~65 KB key cap. The name is unique per doc and derived from the doc index
// so the query target below is deterministic.
const largeDocCount = 250
const largeValuePadLen = 50 * 1024

func largeNameValue(i int) string {
	return fmt.Sprintf("%04d", i) + strings.Repeat("a", largeValuePadLen)
}

// TestIndex_OverManyLargeExistingDocs_CreatesAndDeletes covers both halves of #4907 over one
// expensive ~12.5 MB collection: creating the index backfills it (used by a filtered query),
// and deleting it removes every entry (query falls back to a full scan). Both operations must
// stay within the transaction size limit via batching.
func TestIndex_OverManyLargeExistingDocs_CreatesAndDeletes(t *testing.T) {
	req := fmt.Sprintf(`query { User(filter: {name: {_eq: %q}}) { age } }`, largeNameValue(42))

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age:  Int
					}
				`,
			},
			// Each doc's name is a unique ~50 KB value and its age equals the doc index,
			// both derived from i so the query target is deterministic.
			testUtils.GenerateDocs{
				Options: []gen.Option{
					gen.WithTypeDemand("User", largeDocCount),
					gen.WithFieldGenerator("User", "name", func(i int, _ func() any) any {
						return largeNameValue(i)
					}),
					gen.WithFieldGenerator("User", "age", func(i int, _ func() any) any {
						return i
					}),
				},
			},

			// Create: the backfill must not exceed the transaction size limit.
			&action.NewIndex{
				IndexName: "User_name",
				FieldName: "name",
			},
			&action.ListIndexes{
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "User_name",
						ID:     1,
						Fields: []client.IndexedFieldDescription{{Name: "name"}},
					},
				},
			},
			// Querying one document by its value proves the index was backfilled correctly,
			// and the explain proves the planner uses it rather than scanning.
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{{"age": int64(42)}},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},

			// Delete: the batched GC must not exceed the transaction size limit either.
			&action.DeleteIndex{
				IndexName: "User_name",
			},
			&action.ListIndexes{
				ExpectedIndexes: []client.IndexDescription{},
			},
			// The query still returns the correct document, now via a full scan.
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{{"age": int64(42)}},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
