// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package p2p

import (
	"bytes"
	"context"
	"testing"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/corekv/memory"

	"github.com/sourcenetwork/defradb/internal/datastore"
)

func TestBlockSyncCAR_RoundTripRoutesEncBlocksAndSetsToMergeMarker(t *testing.T) {
	ctx := context.Background()

	b1 := blocks.NewBlock([]byte("block-one"))
	b2 := blocks.NewBlock([]byte("block-two"))
	encB := blocks.NewBlock([]byte("encryption-envelope"))

	roots := []cid.Cid{b1.Cid()}

	var buf bytes.Buffer
	err := writeCAR(ctx, &buf, roots, []blocks.Block{b1, b2, encB})
	require.NoError(t, err)

	rootstore := memory.NewDatastore(ctx)
	blockDst := datastore.P2PBlockstoreFrom(rootstore, immutable.None[int]())
	encDst := datastore.EncstoreFrom(rootstore)
	encCIDs := map[cid.Cid]struct{}{encB.Cid(): {}}

	gotRoots, err := ingestCAR(ctx, blockDst, encDst, encCIDs, &buf)
	require.NoError(t, err)
	require.Equal(t, roots, gotRoots)

	// Regular blocks land in the block store and are flagged as not-yet-merged (the to-merge
	// marker is set), so the merge step will pick them up.
	for _, b := range []blocks.Block{b1, b2} {
		has, err := blockDst.Has(ctx, b.Cid())
		require.NoError(t, err)
		require.True(t, has)

		merged, err := blockDst.IsMerged(ctx, b.Cid())
		require.NoError(t, err)
		require.False(t, merged)
	}

	// The encryption block is routed to the encryption store, not the block store.
	hasEncInBlocks, err := blockDst.Has(ctx, encB.Cid())
	require.NoError(t, err)
	require.False(t, hasEncInBlocks)

	hasEnc, err := encDst.Has(ctx, encB.Cid())
	require.NoError(t, err)
	require.True(t, hasEnc)
}
