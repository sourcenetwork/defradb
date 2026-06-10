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

// Package index contains integration tests for secondary indexes.
//
// This file is a regression guard for issue #4907: creating a secondary index on a
// collection whose existing documents total more than ~11MB must not fail. The current
// (unfixed) code runs backfill inside a single badger transaction that overflows
// badger's built-in txn-size cap, so NewIndex returns an ErrTxnTooBig error. This
// test is therefore EXPECTED TO FAIL on the unpatched code path and MUST PASS once
// the batched-backfill implementation is in place.

package index

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// TestIndexCreate_WithManyLargeExistingDocs_ShouldSucceed verifies that calling
// NewIndex on a collection that already holds many large documents succeeds without
// hitting badger's transaction-size limit.
//
// Sizing rationale: badger's per-transaction cap is ~11 MB. Index keys embed the
// indexed field value, so 250 documents each carrying a ~50 KB name value produce
// roughly 12.5 MB of index data — guaranteed to overflow a single-transaction
// backfill while keeping each individual key comfortably below badger's ~65 KB
// per-key limit. Each document is inserted via its own AddDoc action (separate
// transactions) so that only the backfill step is the one being exercised.
func TestIndexCreate_WithManyLargeExistingDocs_ShouldSucceed(t *testing.T) {
	const numDocs = 250
	const valuePadLen = 50 * 1024 // 50 KB of padding per name field

	actions := make([]any, 0, numDocs+5)

	// Step 1: define the collection without any inline @index directive so that
	// the index can be added later via NewIndex, triggering the backfill path.
	actions = append(actions, &action.AddCollection{
		SDL: `
			type User {
				name: String
				age:  Int
			}
		`,
	})

	// Step 2: insert 250 documents, each with a unique ~50 KB name value.
	// Using individual AddDoc actions ensures every insert runs in its own
	// transaction — only the subsequent NewIndex call is put under pressure.
	for i := 0; i < numDocs; i++ {
		name := fmt.Sprintf("%04d", i) + strings.Repeat("a", valuePadLen)
		doc := fmt.Sprintf(`{"name": %q, "age": %d}`, name, i)
		actions = append(actions, &action.AddDoc{
			Doc: doc,
		})
	}

	// Step 3: create the secondary index AFTER all documents exist.
	// On unpatched code this overflows the single badger transaction that
	// backfill currently uses and returns an ErrTxnTooBig error.
	actions = append(actions, &action.NewIndex{
		IndexName: "User_name",
		FieldName: "name",
	})

	// Step 4: confirm the index is registered correctly.
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

	// Step 5: query a specific document by its large name value to prove the
	// index was backfilled with the correct content.
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

	// Step 6: prove the planner actually uses the index (not a full scan).
	actions = append(actions, &action.Request{
		Request:  makeExplainQuery(req),
		Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
	})

	test := testUtils.TestCase{
		Actions: actions,
	}

	testUtils.ExecuteTestCase(t, test)
}
