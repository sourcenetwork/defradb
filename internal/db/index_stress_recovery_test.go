// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// These tests attack recovery itself: interrupting it partway, running it many times in a row, and
// pointing it at a deliberately tangled on-disk state (a partial build record sitting on top of
// several stale epochs). Recovery must be safe to interrupt and repeat — that is the whole contract
// that lets startup retry it after a crash.

// TestRecovery_Idempotent_MultipleConsecutiveRuns reproduces an interrupted rebuild, then runs
// recovery several times back to back. The first run repairs the index; every later run must be a
// no-op that leaves the single clean live epoch intact.
func TestRecovery_Idempotent_MultipleConsecutiveRuns(t *testing.T) {
	const docCount = 120
	ctx := context.Background()
	db, names, collectionID, shortID, desc := setupIndexedOnDisk(t, ctx, docCount)

	// Interrupted rebuild: epoch 2 partially built with a watermark, epoch 1 intact.
	partialFillNewEpoch(t, ctx, db.DB, freshUserVersion(t, ctx, db.DB), desc, 50)

	for i := range 4 {
		require.NoErrorf(t, db.recoverIndexStates(ctx), "recovery run %d failed", i+1)
		auditIndex(t, ctx, db.DB, collectionID, shortID, desc.ID, docCount)
		requireNoInflightState(t, ctx, db.DB)
	}
	requireQueryable(t, ctx, db.DB, names)
}

// TestRecovery_InterruptedThenResumed cancels recovery partway (modelling a crash during recovery),
// leaving the index still in its interrupted state, then runs recovery again to completion. The
// second recovery must finish the job. This is the "crash the recovery itself" case.
func TestRecovery_InterruptedThenResumed(t *testing.T) {
	const docCount = 140
	ctx := context.Background()
	db, names, collectionID, shortID, desc := setupIndexedOnDisk(t, ctx, docCount)

	// Interrupted rebuild: epoch 2 partially built, epoch 1 intact, build record present.
	partialFillNewEpoch(t, ctx, db.DB, freshUserVersion(t, ctx, db.DB), desc, 40)

	// First recovery is cancelled before it can do anything: recoverIndexStates checks ctx.Err()
	// and bails. The interrupted state must be unchanged.
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	require.NoError(t, db.recoverIndexStates(cancelled))

	st := readIndexState(t, ctx, db.DB, collectionID, desc.ID)
	require.True(t, st.isBuilding(), "cancelled recovery must leave the build record in place")
	require.Equal(t, docCount, countIndexEpochEntries(t, ctx, db.DB, shortID, desc.ID, 1),
		"cancelled recovery must not touch the old epoch")

	// Second recovery runs to completion and repairs everything.
	require.NoError(t, db.recoverIndexStates(ctx))
	auditIndex(t, ctx, db.DB, collectionID, shortID, desc.ID, docCount)
	requireNoInflightState(t, ctx, db.DB)
	requireQueryable(t, ctx, db.DB, names)
}

// TestRecovery_TangledState_BuildRecordOverStaleEpochs points recovery at the worst tangle a crash
// can produce: a partial build record for epoch N sitting on top of several fully-built stale
// epochs below it (the leftovers of earlier interrupted rebuilds that never got swept). Recovery
// must finish the build of the live epoch AND collect every stale epoch beneath it.
func TestRecovery_TangledState_BuildRecordOverStaleEpochs(t *testing.T) {
	const (
		docCount    = 100
		staleEpochs = 4
	)
	ctx := context.Background()
	db, names, collectionID, shortID, desc := setupIndexedOnDisk(t, ctx, docCount)

	// Pile up several fully-built stale epochs (earlier rebuilds whose stale GC never ran).
	stackStaleEpochs(t, ctx, db.DB, freshUserVersion(t, ctx, db.DB), desc, staleEpochs)

	// Now start one more rebuild and only partially build its epoch, leaving a build record. This is
	// the live epoch; everything below it is stale.
	partialFillNewEpoch(t, ctx, db.DB, freshUserVersion(t, ctx, db.DB), desc, 30)

	live := liveEpochOfIndex(t, ctx, db.DB, collectionID, desc.ID)
	require.Equal(t, uint32(1+staleEpochs+1), live)
	st := readIndexState(t, ctx, db.DB, collectionID, desc.ID)
	require.True(t, st.isBuilding(), "expected a building record over the stale stack")

	require.NoError(t, db.recoverIndexStates(ctx))

	auditIndex(t, ctx, db.DB, collectionID, shortID, desc.ID, docCount)
	requireNoInflightState(t, ctx, db.DB)
	requireQueryable(t, ctx, db.DB, names)
}

// TestRecovery_RepeatedReopen_Converges reopens the on-disk DB several times after an interrupted
// rebuild. Each reopen runs full startup recovery; the index must converge to one clean live epoch
// and stay there across reopens, never regressing or leaking.
func TestRecovery_RepeatedReopen_Converges(t *testing.T) {
	const docCount = 110
	ctx := context.Background()
	db, names, collectionID, shortID, desc := setupIndexedOnDisk(t, ctx, docCount)

	partialFillNewEpoch(t, ctx, db.DB, freshUserVersion(t, ctx, db.DB), desc, 45)

	for i := range 3 {
		db.reopen(t, ctx)
		auditIndex(t, ctx, db.DB, collectionID, shortID, desc.ID, docCount)
		requireNoInflightStatef(t, ctx, db.DB, "after reopen %d", i+1)
	}
	requireQueryable(t, ctx, db.DB, names)
}

// requireNoInflightStatef is requireNoInflightState with a contextual message.
func requireNoInflightStatef(t *testing.T, ctx context.Context, db *DB, format string, args ...any) {
	t.Helper()
	recs := scanAllIndexStateRecords(t, ctx, db)
	require.Emptyf(t, recs, format+": expected no in-flight index state records, found %+v", append(args, recs)...)
}
