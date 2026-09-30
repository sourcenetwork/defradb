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
	"sync"
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

// goClientTypes restricts a gated test to the Go client, the only one running the DB in the test
// process where db.IndexBuildGate takes effect.
//
// inProcessDBType pins to a single backend: the gated behaviour is storage-agnostic, so running it
// across backends only adds cost, and leveldb must be avoided since a build parked at the gate would
// hold no txn but its concurrent-transaction limit makes the held-build scenarios fragile
// (https://github.com/sourcenetwork/defradb/issues/4959).
var (
	goClientTypes   = immutable.Some([]state.ClientType{state.GoClientType})
	inProcessDBType = immutable.Some([]state.DatabaseType{testUtils.BadgerIMType})
)

// installBuildGate sets db.IndexBuildGate to block the backfill at each batch boundary until release
// is called, holding the index at the building state. cleanup releases the gate and restores it.
func installBuildGate(t *testing.T) (release func(), cleanup func()) {
	t.Helper()
	orig := db.IndexBuildGate
	releaseCh := make(chan struct{})
	var releaseOnce, restoreOnce sync.Once

	db.IndexBuildGate = func(ctx context.Context, _ string, _ uint32) {
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
	return release, cleanup
}

// TestAsyncBackfill_WhileGated_ReportsBuildingThenReady creates an index on a populated collection:
// it is reported as building while the gate holds it, then ready once released, and the index is
// used by a filtered query.
func TestAsyncBackfill_WhileGated_ReportsBuildingThenReady(t *testing.T) {
	release, cleanup := installBuildGate(t)
	defer cleanup()

	req := `query { User(filter: {name: {_eq: "Fred"}}) { age } }`

	test := testUtils.TestCase{
		SupportedClientTypes:   goClientTypes,
		SupportedDatabaseTypes: inProcessDBType,
		Actions: []any{
			&action.AddCollection{
				SDL: `type User { name: String  age: Int }`,
			},
			&action.AddDoc{Doc: `{"name": "Fred", "age": 33}`},
			&action.AddDoc{Doc: `{"name": "John", "age": 41}`},
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "name",
				Async:        true,
			},
			// While the gate holds the build, the index is listed with an in-progress status.
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
					"User_name_ASC": {
						Status: client.InProgressActionStatus,
						Action: client.BackfillIndexAction,
					},
				},
			},
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
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{{"age": int64(33)}},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestAsyncBackfill_QueryWhileBuilding_FullScans checks that while an index is held at building, a
// filtered query returns correct results via a full scan (zero index fetches), then the build is
// released to ready.
func TestAsyncBackfill_QueryWhileBuilding_FullScans(t *testing.T) {
	release, cleanup := installBuildGate(t)
	defer cleanup()

	req := `query { User(filter: {name: {_eq: "Fred"}}) { age } }`

	test := testUtils.TestCase{
		SupportedClientTypes:   goClientTypes,
		SupportedDatabaseTypes: inProcessDBType,
		Actions: []any{
			&action.AddCollection{
				SDL: `type User { name: String  age: Int }`,
			},
			&action.AddDoc{Doc: `{"name": "Fred", "age": 33}`},
			&action.AddDoc{Doc: `{"name": "John", "age": 41}`},
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "name",
				Async:        true,
			},
			// While building, the planner must not use the partial index: full scan, 0 fetches.
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{{"age": int64(33)}},
				},
			},
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(0),
			},
			action.NewRunFunc(release),
			&action.WaitForIndexReady{CollectionID: 0},
			// Once ready, the planner uses the index.
			&action.Request{
				Request:  makeExplainQuery(req),
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestAsyncBackfill_RestartMidBuild_ResumesToReady checks that a build interrupted by a restart
// resumes to ready. The gate holds it at building so it is provably unfinished at restart.
//
// The gate is process-global and survives the restart. On Close the first instance's gated build
// returns via ctx (leaving the building record); after restart the second instance's worker hits
// the gate again, and releasing it lets that build finish. The watermark-resume off-by-one is
// covered by the unit tests in index_recovery_test.go.
func TestAsyncBackfill_RestartMidBuild_ResumesToReady(t *testing.T) {
	release, cleanup := installBuildGate(t)
	defer cleanup()

	req := `query { User(filter: {name: {_eq: "Fred"}}) { age } }`

	test := testUtils.TestCase{
		// In-process only: the restart-resume path is a node-local worker behaviour, and the gate
		// is process-global.
		SupportedClientTypes:   goClientTypes,
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{testUtils.BadgerFileType}),
		Actions: []any{
			&action.AddCollection{
				SDL: `type User { name: String  age: Int }`,
			},
			&action.AddDoc{Doc: `{"name": "Fred", "age": 33}`},
			&action.AddDoc{Doc: `{"name": "John", "age": 41}`},
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "name",
				Async:        true,
			},
			// Restart with the build in flight. The first instance's gated build returns on Close;
			// the second instance's worker re-drains the still-present building record.
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
			&action.Request{
				Request: req,
				Results: map[string]any{
					"User": []map[string]any{{"age": int64(33)}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestAsyncBackfill_WriteWhileBuilding_ReflectedInIndex checks that documents written while an index
// builds are reflected in the final index. The gate holds the build while the write happens, then
// the build is released and must converge on the final set.
func TestAsyncBackfill_WriteWhileBuilding_ReflectedInIndex(t *testing.T) {
	release, cleanup := installBuildGate(t)
	defer cleanup()

	test := testUtils.TestCase{
		SupportedClientTypes:   goClientTypes,
		SupportedDatabaseTypes: inProcessDBType,
		Actions: []any{
			&action.AddCollection{
				SDL: `type User { name: String  age: Int }`,
			},
			&action.AddDoc{Doc: `{"name": "Fred", "age": 33}`},
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "name",
				Async:        true,
			},
			// Write while gated: the live write maintains the building index, the backfill fills
			// the rest.
			&action.AddDoc{Doc: `{"name": "John", "age": 41}`},
			action.NewRunFunc(release),
			&action.WaitForIndexReady{CollectionID: 0},
			&action.Request{
				Request: `query { User(filter: {name: {_eq: "John"}}) { age } }`,
				Results: map[string]any{
					"User": []map[string]any{{"age": int64(41)}},
				},
			},
			&action.Request{
				Request: `query { User(filter: {name: {_eq: "Fred"}}) { age } }`,
				Results: map[string]any{
					"User": []map[string]any{{"age": int64(33)}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
