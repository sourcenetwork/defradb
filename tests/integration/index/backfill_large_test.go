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

// largeDocCount × largeNameValue gives ~12.5 MB of data. Index keys embed the field value,
// so the backfill cannot fit in one ~11 MB transaction, while each key stays below the
// ~65 KB key cap. The name is unique per doc and derived from the doc index so the query
// target below is deterministic.
const largeDocCount = 250
const largeValuePadLen = 50 * 1024

func largeNameValue(i int) string {
	return fmt.Sprintf("%04d", i) + strings.Repeat("a", largeValuePadLen)
}

// genLargeUsers generates largeDocCount User docs. Each doc's name is a unique ~50 KB value
// and its age equals the doc index, both derived from i so the query target is deterministic.
func genLargeUsers() testUtils.GenerateDocs {
	return testUtils.GenerateDocs{
		Options: []gen.Option{
			gen.WithTypeDemand("User", largeDocCount),
			gen.WithFieldGenerator("User", "name", func(i int, _ func() any) any {
				return largeNameValue(i)
			}),
			gen.WithFieldGenerator("User", "age", func(i int, _ func() any) any {
				return i
			}),
		},
	}
}

// TestIndexCreate_WithManyLargeExistingDocs_ShouldSucceed builds an index over ~12.5 MB of
// existing data and checks the index is ready and a filtered query uses it.
func TestIndexCreate_WithManyLargeExistingDocs_ShouldSucceed(t *testing.T) {
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
			genLargeUsers(),
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
			// Querying one document by its value proves the index was backfilled correctly.
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{{"age": int64(42)}},
				},
			},
			// The planner must use the index, not a full scan.
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestIndexDelete_WithManyLargeExistingDocs_ShouldSucceed is the deletion half of the #4907
// regression test: deleting an index over ~12.5 MB of data must not exceed the transaction
// size limit, after which the query falls back to a full scan.
func TestIndexDelete_WithManyLargeExistingDocs_ShouldSucceed(t *testing.T) {
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
			genLargeUsers(),
			&action.NewIndex{
				IndexName: "User_name",
				FieldName: "name",
			},
			&action.DeleteIndex{
				IndexName: "User_name",
			},
			&action.ListIndexes{
				ExpectedIndexes: []client.IndexDescription{},
			},
			// The filtered query still returns the correct document via a full scan.
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
