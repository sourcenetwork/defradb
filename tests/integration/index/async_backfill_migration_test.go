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

package index

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/lenses"
	"github.com/sourcenetwork/defradb/tests/state"
)

// onDiskDBType pins a gated test to the on-disk badger store, matching the restart tests in
// async_backfill_test.go so the gate and worker run against the production storage backend.
var onDiskDBType = immutable.Some([]state.DatabaseType{testUtils.BadgerFileType})

// installMultiEntryBuildGate is like installBuildGate but also counts how many distinct collectionIDs
// have entered the gate. waitEnteredAtLeast blocks until that many distinct collections are stopped at
// the gate, so a later assertion sees them building together instead of racing the worker.
func installMultiEntryBuildGate(t *testing.T) (
	release func(),
	cleanup func(),
	waitEnteredAtLeast func(n int),
) {
	t.Helper()
	orig := db.IndexBuildGate
	releaseCh := make(chan struct{})
	var releaseOnce, restoreOnce sync.Once

	var mu sync.Mutex
	entered := map[string]bool{}

	db.IndexBuildGate = func(ctx context.Context, collectionID string, _ uint32) {
		mu.Lock()
		entered[collectionID] = true
		mu.Unlock()
		// Honour ctx so the build does not block shutdown if the DB closes while gated.
		select {
		case <-releaseCh:
		case <-ctx.Done():
		}
	}

	release = func() {
		releaseOnce.Do(func() { close(releaseCh) })
	}
	cleanup = func() {
		// Release and restore even if the test fails early.
		release()
		restoreOnce.Do(func() { db.IndexBuildGate = orig })
	}
	waitEnteredAtLeast = func(n int) {
		require.Eventually(t, func() bool {
			mu.Lock()
			defer mu.Unlock()
			return len(entered) >= n
		}, indexBuildTimeoutForTest, time.Millisecond,
			"timed out waiting for %d distinct collections' builds to reach the gate", n)
	}
	return release, cleanup, waitEnteredAtLeast
}

const indexBuildTimeoutForTest = 10 * time.Second

// installFirstBatchThroughGate lets the first batch through so it commits a non-zero watermark, then
// blocks every later batch until release is called. waitBlocked returns once a later batch has hit the
// gate, which proves the first batch already committed.
//
// The count is never reset, including across a restart: a build resumed after a restart keeps calling
// the same gate, so its first call after the restart is already past the first batch and blocks.
func installFirstBatchThroughGate(t *testing.T) (
	release func(),
	cleanup func(),
	waitBlocked func(),
) {
	t.Helper()
	orig := db.IndexBuildGate
	releaseCh := make(chan struct{})
	var releaseOnce, restoreOnce sync.Once

	var mu sync.Mutex
	var count int
	blocked := make(chan struct{})
	var blockedOnce sync.Once

	db.IndexBuildGate = func(ctx context.Context, _ string, _ uint32) {
		mu.Lock()
		count++
		n := count
		mu.Unlock()

		if n == 1 {
			return
		}

		blockedOnce.Do(func() { close(blocked) })
		// Honour ctx so the build does not hold up shutdown if the DB closes while it is blocked.
		select {
		case <-releaseCh:
		case <-ctx.Done():
		}
	}

	release = func() {
		releaseOnce.Do(func() { close(releaseCh) })
	}
	cleanup = func() {
		release()
		restoreOnce.Do(func() { db.IndexBuildGate = orig })
	}
	waitBlocked = func() {
		select {
		case <-blocked:
		case <-time.After(indexBuildTimeoutForTest):
			t.Fatal("timed out waiting for the build to block after its first batch")
		}
	}
	return release, cleanup, waitBlocked
}

// TestAsyncBackfill_TwoCollectionsMigratedTogether_BothReindexConcurrentlyAndConverge patches two
// collections in one call, each with a migration that reindexes its indexed field. Both builds stage
// on the same commit, so the worker runs them at once: while gated, ListIndexes shows both in-progress
// together. After release, every index must converge and queries on both must use the migrated values.
func TestAsyncBackfill_TwoCollectionsMigratedTogether_BothReindexConcurrentlyAndConverge(t *testing.T) {
	release, cleanup, waitEnteredAtLeast := installMultiEntryBuildGate(t)
	defer cleanup()

	bookReq := `query { Book(filter: {rank: {_eq: 25}}) { title } }`
	authorReq := `query { Author(filter: {rank: {_eq: 15}}) { name } }`

	test := testUtils.TestCase{
		SupportedClientTypes:   goClientTypes,
		SupportedDatabaseTypes: onDiskDBType,
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						title: String
						rank: Int @index
					}
					type Author {
						name: String
						rank: Int @index
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"title": "Dune", "rank": 20}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc:          `{"name": "Herbert", "rank": 10}`,
			},
			// One patch, one commit: both collections stage a reindex on the same transaction. Patch-
			// Collection waits for the reindex, so run it async and await it below, after the in-progress
			// state is observed and the gate released.
			&action.Async{
				Child: &action.PatchCollection{
					Patch: `
						[
							{ "op": "add", "path": "/Book/Fields/-", "value": {"Name": "pages", "Kind": "Int"} },
							{ "op": "add", "path": "/Author/Fields/-", "value": {"Name": "born", "Kind": "Int"} }
						]
					`,
					Lens: immutable.Some(model.Lens{
						Lenses: []model.LensModule{
							{
								Path: lenses.IncrementModulePath,
								Arguments: map[string]any{
									"field": "rank",
									"value": 5,
								},
							},
						},
					}),
				},
			},
			// Wait until both rebuilds reach the gate, so they are provably in flight together before the
			// ListIndexes assertion below sees them both in-progress.
			action.NewRunFunc(func() { waitEnteredAtLeast(2) }),
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "Book_rank_ASC",
						ID:     1,
						Fields: []client.IndexedFieldDescription{{Name: "rank"}},
					},
				},
				ExpectedStatuses: map[string]client.ActionExecution{
					"Book_rank_ASC": {
						Status: client.InProgressActionStatus,
						Action: client.BackfillIndexAction,
					},
				},
			},
			&action.ListIndexes{
				CollectionID: 1,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "Author_rank_ASC",
						ID:     1,
						Fields: []client.IndexedFieldDescription{{Name: "rank"}},
					},
				},
				ExpectedStatuses: map[string]client.ActionExecution{
					"Author_rank_ASC": {
						Status: client.InProgressActionStatus,
						Action: client.BackfillIndexAction,
					},
				},
			},
			action.NewRunFunc(release),
			&action.Await{},
			&action.ListIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "Book_rank_ASC",
						ID:     1,
						Fields: []client.IndexedFieldDescription{{Name: "rank"}},
					},
				},
				ExpectedStatuses: map[string]client.ActionExecution{
					"Book_rank_ASC": {Status: client.CompletedActionStatus},
				},
			},
			&action.ListIndexes{
				CollectionID: 1,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name:   "Author_rank_ASC",
						ID:     1,
						Fields: []client.IndexedFieldDescription{{Name: "rank"}},
					},
				},
				ExpectedStatuses: map[string]client.ActionExecution{
					"Author_rank_ASC": {Status: client.CompletedActionStatus},
				},
			},
			&action.Request{
				Request: bookReq,
				Results: map[string]any{
					"Book": []map[string]any{{"title": "Dune"}},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(bookReq),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
			&action.Request{
				Request: authorReq,
				Results: map[string]any{
					"Author": []map[string]any{{"name": "Herbert"}},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(authorReq),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestAsyncBackfill_EpochAdvancedMidBuild_LiveEpochComplete is the integration version of the unit
// test of the same name in internal/db/index_rebuild_test.go. A build keeps its epoch for its whole
// run, so a migration that advances the epoch mid-build must not split the build across two epochs.
// The build is stopped at the gate before its first batch, a migration runs while it waits, then the
// gate is released. The live epoch must hold every document.
func TestAsyncBackfill_EpochAdvancedMidBuild_LiveEpochComplete(t *testing.T) {
	release, cleanup := installBuildGate(t)
	defer cleanup()

	const docCount = 150 // several times indexBackfillBatchSize (100) so the rebuild spans multiple batches

	actions := []any{
		&action.AddCollection{
			SDL: `
				type User {
					name: String
					level: Int
				}
			`,
		},
	}
	for i := range docCount {
		actions = append(actions, &action.AddDoc{
			Doc: fmt.Sprintf(`{"name": "n%03d", "level": 1}`, i),
		})
	}
	actions = append(actions,
		// This build is the one that gets stopped at the gate and then superseded below.
		&action.NewIndex{
			FieldName: "name",
			Async:     true,
		},
		// Run a migration while the build waits at the gate. It creates a new active version and
		// reindexes the index, advancing its epoch while the old build is still stopped.
		&action.Async{
			Child: &action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/User/Fields/-", "value": {"Name": "score", "Kind": "Int"} }
					]
				`,
				Lens: immutable.Some(model.Lens{
					Lenses: []model.LensModule{
						{
							Path: lenses.IncrementModulePath,
							Arguments: map[string]any{
								"field": "level",
								"value": 1,
							},
						},
					},
				}),
			},
		},
		action.NewRunFunc(release),
		&action.Await{},
		&action.WaitForIndexReady{CollectionID: 0},
		&action.ListIndexes{
			CollectionID: 0,
			ExpectedIndexes: []client.IndexDescription{
				{
					Name:   "User_name_ASC",
					ID:     1,
					Fields: []client.IndexedFieldDescription{{Name: "name"}},
				},
			},
			ExpectedStatuses: map[string]client.ActionExecution{
				"User_name_ASC": {Status: client.CompletedActionStatus},
			},
		},
	)
	// The live epoch must hold every document indexed before the epoch advance, not just the ones the
	// fresh rebuild would have picked up. The explain confirms the query uses the index, one _eq
	// confirms the migration transform applied (level 1 to 2), and the ordered scan confirms the full
	// set is present.
	allDocs := make([]map[string]any, docCount)
	for i := range docCount {
		allDocs[i] = map[string]any{"name": fmt.Sprintf("n%03d", i)}
	}
	actions = append(actions,
		&action.Request{
			Request:  makeExplainQuery(`query { User(filter: {name: {_eq: "n000"}}) { level } }`),
			Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
		},
		&action.Request{
			Request: `query { User(filter: {name: {_eq: "n000"}}) { level } }`,
			Results: map[string]any{
				"User": []map[string]any{{"level": int64(2)}},
			},
		},
		&action.Request{
			Request: `query { User(order: {name: ASC}) { name } }`,
			Results: map[string]any{
				"User": allDocs,
			},
		},
	)

	test := testUtils.TestCase{
		SupportedClientTypes:   goClientTypes,
		SupportedDatabaseTypes: onDiskDBType,
		Actions:                actions,
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestAsyncBackfill_RestartAfterNonZeroWatermark_ResumesToReady extends
// TestAsyncBackfill_RestartMidBuild_ResumesToReady, which only resumes from an empty watermark since
// its gate blocks before the first batch commits. Here the gate lets the first batch through so it
// commits and advances the watermark, then blocks the second batch. The restart happens with that
// non-zero watermark on disk, so the resumed build must continue from it rather than start over, and
// the final index must still hold every document with none duplicated or skipped.
func TestAsyncBackfill_RestartAfterNonZeroWatermark_ResumesToReady(t *testing.T) {
	release, cleanup, waitBlocked := installFirstBatchThroughGate(t)
	defer cleanup()

	// 250 docs is 3 batches at the default indexBackfillBatchSize of 100: the first batch commits
	// before the gate blocks the build, leaving two more for after the restart.
	const docCount = 250

	actions := []any{
		&action.AddCollection{
			SDL: `
				type User {
					name: String
				}
			`,
		},
	}
	for i := range docCount {
		actions = append(actions, &action.AddDoc{
			Doc: fmt.Sprintf(`{"name": "n%03d"}`, i),
		})
	}
	actions = append(actions,
		&action.NewIndex{
			FieldName: "name",
			Async:     true,
		},
		action.NewRunFunc(waitBlocked),
		// The watermark is now non-zero. Restart with the build still blocked, so the resumed worker
		// must continue from that watermark instead of rebuilding from scratch.
		testUtils.Restart{},
		action.NewRunFunc(release),
		&action.WaitForIndexReady{CollectionID: 0},
		&action.ListIndexes{
			CollectionID: 0,
			ExpectedIndexes: []client.IndexDescription{
				{
					Name:   "User_name_ASC",
					ID:     1,
					Fields: []client.IndexedFieldDescription{{Name: "name"}},
				},
			},
			ExpectedStatuses: map[string]client.ActionExecution{
				"User_name_ASC": {Status: client.CompletedActionStatus},
			},
		},
	)
	// The explain confirms the query uses the index, and the ordered scan confirms the full set is
	// present with none missing or duplicated: the resumed build must not skip documents before the
	// saved watermark nor re-emit ones already covered by it.
	allDocs := make([]map[string]any, docCount)
	for i := range docCount {
		allDocs[i] = map[string]any{"name": fmt.Sprintf("n%03d", i)}
	}
	actions = append(actions,
		&action.Request{
			Request:  makeExplainQuery(`query { User(filter: {name: {_eq: "n000"}}) { name } }`),
			Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
		},
		&action.Request{
			Request: `query { User(order: {name: ASC}) { name } }`,
			Results: map[string]any{
				"User": allDocs,
			},
		},
	)

	test := testUtils.TestCase{
		// In-process only: the restart-resume path is a node-local worker behaviour, and the gate
		// is process-global.
		SupportedClientTypes:   goClientTypes,
		SupportedDatabaseTypes: onDiskDBType,
		Actions:                actions,
	}

	testUtils.ExecuteTestCase(t, test)
}
