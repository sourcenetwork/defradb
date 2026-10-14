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
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	badgerds "github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/badger"

	"github.com/sourcenetwork/defradb/acp/dac"
	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
	intOpts "github.com/sourcenetwork/defradb/internal/options"
)

var errInjectedDrain = errors.New("injected drain failure")

// setGate installs the drain gate for the rest of the test.
func setGate(t *testing.T, fn func(ctx context.Context, collectionID string, indexID uint32) error) {
	t.Helper()
	vectorMaintenanceGate = fn
	t.Cleanup(func() { vectorMaintenanceGate = nil })
}

// A failed drain must keep the index dirty and leave the entries in place, or the work would be lost
// until the next write to the index.
func TestVectorMaintenance_Drain_FailureKeepsIndexDirtyAndEntriesPending(t *testing.T) {
	f := newDeferredFixture(t, "", true)
	for i := range 3 {
		f.add("d", []float32{1, float32(i), 0})
	}
	require.Equal(t, 3, f.pending())

	var fail atomic.Bool
	fail.Store(true)
	setGate(t, func(context.Context, string, uint32) error {
		if fail.Load() {
			return errInjectedDrain
		}
		return nil
	})

	err := f.db.vectorMaintenance.drainSync(f.ctx)
	require.ErrorIs(t, err, errInjectedDrain)
	assert.Equal(t, 1, f.dirty(), "a failed index must stay dirty")
	assert.Equal(t, 3, f.pending())

	fail.Store(false)
	f.drain()
	assert.Equal(t, 0, f.pending())
	assert.Equal(t, 0, f.dirty())
}

// With no new write, only the retry timer can make the writer try a failed index again.
func TestVectorMaintenance_RunLoop_RetriesFailedDrainWithoutNewWrites(t *testing.T) {
	setForTest(t, &vectorMaintenanceRetryDelay, 10*time.Millisecond)
	setForTest(t, &vectorMaintenanceMaxRetryDelay, 50*time.Millisecond)

	const failures = 3
	var calls atomic.Int32
	setGate(t, func(context.Context, string, uint32) error {
		if calls.Add(1) <= failures {
			return errInjectedDrain
		}
		return nil
	})
	f := newDeferredFixture(t, "", false)

	f.add("a", []float32{1, 0, 0})
	f.add("b", []float32{0, 1, 0})

	require.Eventually(t, func() bool { return f.pending() == 0 }, 10*time.Second, 5*time.Millisecond)
	assert.Greater(t, calls.Load(), int32(failures))
}

// Conflicts that outlast the transaction retries halve the batch for this drain only. Unlike a
// too-big transaction the smaller size is not remembered.
func TestVectorMaintenance_Drain_ConflictsHalveBatchButAreNotRemembered(t *testing.T) {
	f := newDeferredFixture(t, "", true)
	doc := f.add("a", []float32{1, 0, 0})
	f.add("b", []float32{0, 1, 0})

	// Overwrite the entry after every read of the first full set of attempts, so the batch commit
	// conflicts each time and the retries run out.
	exhaust := f.db.MaxTxnRetries()
	f.onPendingRead(func(call int) {
		if call <= exhaust {
			f.update(doc, []float32{float32(call), 1, 0})
		}
	})

	f.drain()

	assert.Greater(t, f.pendingReadCalls, exhaust, "the retries must have run out first")
	assert.Equal(t, 0, f.pending())
	assert.Equal(t, deferredBatchSize, f.db.vectorMaintenance.batchLimit(f.ref()))
	assert.False(t, f.hasBatchSize(f.ref()))
}

// The index can be dropped between the writer finding it and applying a batch. The writer must
// treat that as nothing to do.
func TestVectorMaintenance_Drain_IndexDroppedMidDrainIsForgotten(t *testing.T) {
	f := newLifecycleFixture(t)
	for i := range 3 {
		f.add("d", []float32{1, float32(i), 0})
	}
	collectionID := f.col.Version().CollectionID
	shortID := getCollectionShortID(t, f.ctx, f.db, collectionID)
	idx := f.col.Version().Indexes[0]
	ref := f.ref()

	dropped := false
	setGate(t, func(context.Context, string, uint32) error {
		if !dropped {
			dropped = true
			deleteIndexSync(t, f.ctx, f.db, f.col, idx.Name)
		}
		return nil
	})

	require.NoError(t, f.db.vectorMaintenance.drainSync(f.ctx))

	require.True(t, dropped)
	assert.Equal(t, 0, f.dirty())
	// The epoch sequence outlives a drop, so the batch sees an empty list rather than a missing
	// index and the writer keeps a zero counter instead of forgetting it.
	assert.Equal(t, int64(0), f.db.vectorMaintenance.pendingCounter(ref).Load())
	assert.Equal(t, 0, countIndexEntries(t, f.ctx, f.db, shortID, idx.ID))
}

// Only DEFERRED vector indexes belong to the writer. SYNC vector indexes and ordinary indexes in
// the same collection must be left alone.
func TestVectorMaintenance_MarkAllDeferredDirty_SkipsSyncAndOrdinaryIndexes(t *testing.T) {
	f := newDeferredFixtureWith(t, badgerds.DefaultOptions("").WithInMemory(true), mixedIndexesSDL, true)
	w := f.db.vectorMaintenance
	require.Len(t, f.col.Version().Indexes, 3)
	require.Equal(t, 0, f.dirty())

	require.NoError(t, w.markAllDeferredDirty(f.ctx))

	collectionID := f.col.Version().CollectionID
	var deferred []vectorIndexRef
	for _, idx := range f.col.Version().Indexes {
		ref := vectorIndexRef{collectionID, idx.ID}
		if vd, ok := idx.GetVector(); ok && vd.Maintenance == client.VectorIndexMaintenanceDeferred {
			deferred = append(deferred, ref)
			continue
		}
		assert.False(t, f.hasCounter(ref), "index %s must not be tracked", idx.Name)
	}
	require.Len(t, deferred, 1)
	assert.True(t, f.hasCounter(deferred[0]))
	assert.Equal(t, 1, f.dirty())
}

const mixedIndexesSDL = `
	type Users {
		name: String @index
		syncEmbedding: [Float32!] @index(vector: {dimensions: 3, maintenance: SYNC, hnsw: {metric: COSINE}})
		embedding: [Float32!] @index(vector: {dimensions: 3, maintenance: DEFERRED, hnsw: {metric: COSINE}})
	}
`

const usersACPPolicy = `
description: a test policy which marks a collection in a database as a resource
name: test
resources:
- name: users
  permissions:
  - name: delete
  - expr: reader
    name: read
  - name: update
  relations:
  - manages:
    - reader
    name: admin
    types:
    - actor
  - name: reader
    types:
    - actor
`

// The integration ACP test cannot prove this for pending entries, because the background writer
// drains them before the query runs. Here the writer is off, so the private document is still
// pending when it is searched.
func TestVectorMaintenance_PendingUpsert_IsFilteredByDocumentACP(t *testing.T) {
	ctx := context.Background()
	owner, err := identity.Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)
	other, err := identity.Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	rootstore, err := badger.NewDatastore("", badgerds.DefaultOptions("").WithInMemory(true))
	require.NoError(t, err)
	adminInfo, err := acpDB.NewNACInfo(ctx, "", false)
	require.NoError(t, err)
	localACP, err := dac.NewLocalDocumentACP("")
	require.NoError(t, err)

	setForTest(t, &suppressVectorMaintenanceRun, true)
	db, err := newDB(ctx, rootstore, adminInfo, intOpts.DB().SetDocumentACP(localACP))
	require.NoError(t, err)
	t.Cleanup(db.Close)

	policy, err := db.AddDACPolicy(ctx, usersACPPolicy, options.AddDACPolicy().SetIdentity(owner))
	require.NoError(t, err)
	_, err = db.AddCollection(ctx, fmt.Sprintf(`
		type Users @policy(id: %q, resource: "users") {
			name: String
			embedding: [Float32!] @index(vector: {dimensions: 3, maintenance: DEFERRED, hnsw: {metric: COSINE}})
		}
	`, policy.PolicyID))
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "Users")
	require.NoError(t, err)

	add := func(name string, vec []float32, id identity.Identity) {
		data, err := json.Marshal(map[string]any{"name": name, "embedding": vec})
		require.NoError(t, err)
		doc, err := client.NewDocFromJSON(ctx, data, col.Version())
		require.NoError(t, err)
		opts := options.AddDocument()
		if id != nil {
			opts.SetIdentity(id)
		}
		require.NoError(t, col.AddDocument(ctx, doc, opts))
	}
	add("public", []float32{0, 1, 0}, nil)
	add("private", []float32{1, 0, 0}, owner)

	f := &deferredIndexFixture{t: t, ctx: ctx, db: db, col: col}
	require.Equal(t, 2, f.pending(), "both entries must still be pending")

	search := func(id identity.Identity) []string {
		opts := options.ExecRequest()
		if id != nil {
			opts.SetIdentity(id)
		}
		res := db.ExecRequest(ctx, `query {
			Users(order: {_alias: {sim: DESC}}, limit: 3){
				name
				sim: SIMILARITY(embedding: {vector: [1, 0, 0]})
			}
		}`, opts)
		require.Empty(t, res.GQL.Errors)
		var names []string
		for _, row := range res.GQL.Data.(map[string]any)["Users"].([]map[string]any) {
			names = append(names, row["name"].(string))
		}
		return names
	}

	assert.Equal(t, []string{"public"}, search(nil))
	assert.Equal(t, []string{"public"}, search(other))
	assert.Equal(t, []string{"private", "public"}, search(owner))
}
