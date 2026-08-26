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
	"fmt"
	"os"
	"testing"

	badgerds "github.com/dgraph-io/badger/v4"
	badgeropts "github.com/dgraph-io/badger/v4/options"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/badger"

	"github.com/sourcenetwork/defradb/client"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
)

// newOnDiskDB opens a DB on a badger directory with the options a node uses, so the value
// threshold and compression match a deployment. It returns a close func rather than
// registering cleanup so a caller can release it per iteration. An in-memory store has no
// value log and different write amplification, so it cannot stand in for a node when the
// cost of a write path is what is being measured.
func newOnDiskDB(b *testing.B, ctx context.Context) (*DB, func()) {
	b.Helper()

	dir, err := os.MkdirTemp("", "purgebench")
	require.NoError(b, err)

	// Mirrors node/store_badger.go. ZSTDCompressionLevel is left alone because badger
	// already defaults it to the value a node sets.
	opts := badgerds.DefaultOptions(dir)
	opts.ValueThreshold = 1 << 8
	opts.Compression = badgeropts.ZSTD

	rootstore, err := badger.NewDatastore(dir, opts)
	require.NoError(b, err)

	adminInfo, err := acpDB.NewNACInfo(ctx, "", false)
	require.NoError(b, err)

	db, err := newDB(ctx, rootstore, adminInfo)
	require.NoError(b, err)

	return db, func() {
		db.Close()
		_ = os.RemoveAll(dir)
	}
}

// setupIndexedCollection builds a collection with three secondary indexes, one unique. The
// index count decides how many writes each purged document adds to its transaction, which
// is what the pending-write sort scales with.
func setupIndexedCollection(b *testing.B, ctx context.Context, db *DB) client.Collection {
	b.Helper()

	_, err := db.AddCollection(ctx, `type Record {
		hash: String
		blockNumber: Int
		groupID: String
		payload: String
	}`)
	require.NoError(b, err)

	col, err := db.GetCollectionByName(ctx, "Record")
	require.NoError(b, err)

	for _, req := range []client.NewIndexRequest{
		{Fields: []client.IndexedFieldDescription{{Name: "hash"}}, Unique: true},
		{Fields: []client.IndexedFieldDescription{{Name: "blockNumber"}}},
		{Fields: []client.IndexedFieldDescription{{Name: "groupID"}}},
	} {
		_, err := col.NewIndex(ctx, req)
		require.NoError(b, err)
	}

	return col
}

// addRecords writes n documents and returns their IDs. The payload field is unindexed and
// exists to carry each document past the value threshold, so writes reach the value log the
// way they do on a node.
func addRecords(b *testing.B, ctx context.Context, col client.Collection, n int) []client.DocID {
	b.Helper()

	docIDs := make([]client.DocID, 0, n)
	for i := range n {
		doc, err := client.NewDocFromJSON(ctx, fmt.Appendf(nil,
			`{"hash":"0x%064x","blockNumber":%d,"groupID":"g-%d","payload":%q}`,
			i, i/200, i/200, fmt.Sprintf("%0512d", i)), col.Version())
		require.NoError(b, err)
		require.NoError(b, col.AddDocument(ctx, doc))
		docIDs = append(docIDs, doc.ID())
	}

	return docIDs
}

// BenchmarkPurgeByDocIDsChunkSize measures how long it takes to purge a fixed set of
// documents as the number of them sharing a transaction changes.
//
// Both pruneHistory branches are swept. The CLI and HTTP defaults are false and true adds
// the per-document DAG walk, so a size measured under one branch is not automatically right
// for the other. Documents are written locally, so their DAGs are one commit deep; a
// document built up over many merges walks further and costs more than this measures.
//
// The sweep runs below purgeChunkSize so the low end is measured rather than assumed to be
// the endpoint.
//
// This purges single-threaded against an idle store, so it shows how cost scales with chunk
// size and not how a size holds up under concurrent merge traffic, which is what the choice
// of constant turns on.
func BenchmarkPurgeByDocIDsChunkSize(b *testing.B) {
	const docs = 2000

	for _, pruneHistory := range []bool{true, false} {
		for _, chunkSize := range []int{1, 2, 4, 8, 25, 50, 100, 200} {
			b.Run(fmt.Sprintf("prune=%v/chunk=%d", pruneHistory, chunkSize), func(b *testing.B) {
				ctx := context.Background()

				for b.Loop() {
					b.StopTimer()
					db, closeDB := newOnDiskDB(b, ctx)
					col := setupIndexedCollection(b, ctx, db)
					docIDs := addRecords(b, ctx, col, docs)
					concrete, ok := col.(*collection)
					require.True(b, ok)
					b.StartTimer()

					for i := 0; i < len(docIDs); i += chunkSize {
						end := min(i+chunkSize, len(docIDs))
						require.NoError(b, concrete.purgeChunk(ctx, docIDs[i:end], pruneHistory))
					}

					b.StopTimer()
					closeDB()
					b.StartTimer()
				}
			})
		}
	}
}
