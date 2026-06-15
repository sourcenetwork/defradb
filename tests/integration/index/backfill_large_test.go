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

// This file guards against regression of issue #4907: creating a secondary index on a
// collection whose existing documents exceed the storage engine's transaction size limit
// must succeed, which requires the backfill to run in batched transactions.

package index

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// TestIndexCreate_WithManyLargeExistingDocs_ShouldSucceed creates an index over
// ~12.5 MB of existing data: 250 documents, each with a ~50 KB indexed value.
// Index keys embed the field value, so this cannot fit in one ~11 MB transaction,
// while each individual key stays below the ~65 KB key cap.
func TestIndexCreate_WithManyLargeExistingDocs_ShouldSucceed(t *testing.T) {
	const numDocs = 250
	const valuePadLen = 50 * 1024

	actions := make([]any, 0, numDocs+5)

	actions = append(actions, &action.AddCollection{
		SDL: `
			type User {
				name: String
				age:  Int
			}
		`,
	})

	// One AddDoc action per document, so each insert runs in its own transaction
	// and only the backfill is put under pressure.
	for i := 0; i < numDocs; i++ {
		name := fmt.Sprintf("%04d", i) + strings.Repeat("a", valuePadLen)
		doc := fmt.Sprintf(`{"name": %q, "age": %d}`, name, i)
		actions = append(actions, &action.AddDoc{
			Doc: doc,
		})
	}

	actions = append(actions, &action.NewIndex{
		IndexName: "User_name",
		FieldName: "name",
	})

	actions = append(actions, &action.ListIndexes{
		ExpectedIndexes: []client.IndexDescription{
			{
				Name: "User_name",
				ID:   1,
				Fields: []client.IndexedFieldDescription{
					{Name: "name"},
				},
			},
		},
	})

	// Querying one document by its value proves the index was backfilled correctly.
	targetName := fmt.Sprintf("%04d", 42) + strings.Repeat("a", valuePadLen)
	req := fmt.Sprintf(
		`query { User(filter: {name: {_eq: %q}}) { age } }`,
		targetName,
	)

	actions = append(actions, &action.Request{
		Request: req,
		Results: map[string]any{
			"User": []map[string]any{
				{"age": int64(42)},
			},
		},
	})

	// The planner must use the index, not a full scan.
	actions = append(actions, &action.Request{
		Request:  makeExplainQuery(req),
		Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
	})

	test := testUtils.TestCase{
		Actions: actions,
	}

	testUtils.ExecuteTestCase(t, test)
}
