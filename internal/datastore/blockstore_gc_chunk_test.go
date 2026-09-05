// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package datastore

import (
	"context"
	"testing"
	"time"

	blocks "github.com/ipfs/go-block-format"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/internal/keys"
)

// A chunked store appends a suffix to every key it writes, so the CID has to be read off the front
// of the marker rather than from the whole remainder, and a delete has to cover every key the value
// was split across.
//
// A chunk size above the block size leaves each value in one key and exercises only the parse; one
// below it splits the value, which is what the delete has to cover. Both sit at or above the marker
// value's length, since neither half of a split marker decodes as a timestamp and the sweep would
// never nominate it.
//
// Deletions are asserted by prefix, because the unsuffixed key alone passes while the block is
// still stored under its chunked one. The block that survives is asserted by content, because a key
// with the right prefix can hold a value a delete only partly removed.
func TestReclaimOrphanBlocksThroughChunkedStore(t *testing.T) {
	for _, tc := range []struct {
		name  string
		chunk immutable.Option[int]
	}{
		{"unchunked", immutable.None[int]()},
		{"chunked", immutable.Some(1 << 20)},
		{"chunked across several keys", immutable.Some(toMergeValueLen)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			blockNS, systemNS, rootstore := newTestBlockstore(t)
			defer func() { require.NoError(t, rootstore.Close()) }()

			bs := P2PBlockstoreFrom(rootstore, tc.chunk)
			orphan := blocks.NewBlock([]byte("no document claims this"))
			owned := blocks.NewBlock([]byte("a document owns this"))
			require.NoError(t, bs.Put(ctx, orphan))
			require.NoError(t, bs.Put(ctx, owned))
			require.NoError(t, systemNS.Set(ctx,
				keys.NewBlockCIDToDocIDKey(owned.Cid().String(), "bae-owner").Bytes(), []byte{}))

			result, err := ReclaimOrphanBlocks(ctx, rootstore, time.Now().Add(time.Hour), nil, 100)
			require.NoError(t, err)
			require.Equal(t, 2, result.Scanned)
			require.Equal(t, 1, result.Reclaimed)
			require.Equal(t, 1, result.Repaired)

			requirePrefixEmpty(t, ctx, blockNS, orphan.Cid().Bytes(), "orphan block")
			requirePrefixEmpty(t, ctx, blockNS, newToMergeKey(orphan.Cid().Bytes()), "orphan marker")

			// The owned block keeps its data and loses only the marker.
			stored, err := bs.Get(ctx, owned.Cid())
			require.NoError(t, err, "owned block should still be readable")
			require.Equal(t, owned.RawData(), stored.RawData(), "owned block should be unchanged")
			requirePrefixEmpty(t, ctx, blockNS, newToMergeKey(owned.Cid().Bytes()), "owned marker")
		})
	}
}

func requirePrefixEmpty(t *testing.T, ctx context.Context, store corekv.ReaderWriter, prefix []byte, what string) {
	t.Helper()
	require.False(t, prefixHasKey(t, ctx, store, prefix), "%s should be gone, including any chunked key", what)
}

func prefixHasKey(t *testing.T, ctx context.Context, store corekv.ReaderWriter, prefix []byte) bool {
	t.Helper()
	iter, err := store.Iterator(ctx, corekv.IterOptions{Prefix: prefix, KeysOnly: true})
	require.NoError(t, err)
	hasNext, err := iter.Next()
	require.NoError(t, err)
	require.NoError(t, iter.Close())
	return hasNext
}
