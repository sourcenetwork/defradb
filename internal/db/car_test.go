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
	"bytes"
	"context"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-car/v2"
	"github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/require"
)

// makeDummyCID creates a deterministic CIDv1 from a short seed byte slice.
func makeDummyCID(t *testing.T, seed []byte) cid.Cid {
	t.Helper()
	mh, err := multihash.Sum(seed, multihash.SHA2_256, -1)
	require.NoError(t, err)
	return cid.NewCidV1(cid.DagCBOR, mh)
}

func TestGenerateCAR_SingleBlock(t *testing.T) {
	ctx := context.Background()

	data := []byte("hello, CAR")
	blockCID := makeDummyCID(t, data)

	blocks := []savedBlock{
		{cid: blockCID, data: data},
	}

	carBytes, err := generateCAR(ctx, blockCID, blocks)
	require.NoError(t, err)
	require.NotEmpty(t, carBytes)

	// Verify the output is valid CARv1 by reading it back.
	reader, err := car.NewBlockReader(bytes.NewReader(carBytes))
	require.NoError(t, err)

	require.Len(t, reader.Roots, 1)
	require.Equal(t, blockCID, reader.Roots[0])

	blk, err := reader.Next()
	require.NoError(t, err)
	require.Equal(t, blockCID, blk.Cid())
	require.Equal(t, data, blk.RawData())
}

func TestGenerateCAR_MultipleBlocks(t *testing.T) {
	ctx := context.Background()

	payloads := [][]byte{
		[]byte("block-alpha"),
		[]byte("block-beta"),
		[]byte("block-gamma"),
	}

	blocks := make([]savedBlock, len(payloads))
	for i, p := range payloads {
		blocks[i] = savedBlock{cid: makeDummyCID(t, p), data: p}
	}

	rootCID := blocks[0].cid

	carBytes, err := generateCAR(ctx, rootCID, blocks)
	require.NoError(t, err)
	require.NotEmpty(t, carBytes)

	reader, err := car.NewBlockReader(bytes.NewReader(carBytes))
	require.NoError(t, err)

	require.Len(t, reader.Roots, 1)
	require.Equal(t, rootCID, reader.Roots[0])

	// Collect all blocks from the CAR and verify all three are present.
	seen := make(map[cid.Cid][]byte)
	for {
		blk, err := reader.Next()
		if err != nil {
			break
		}
		seen[blk.Cid()] = blk.RawData()
	}

	require.Len(t, seen, len(payloads))
	for _, b := range blocks {
		data, ok := seen[b.cid]
		require.True(t, ok, "block %s not found in CAR", b.cid)
		require.Equal(t, b.data, data)
	}
}

func TestGenerateCAR_EmptyBlocks(t *testing.T) {
	ctx := context.Background()

	// Use a CID derived from some seed as the declared root.
	rootData := []byte("root-only")
	rootCID := makeDummyCID(t, rootData)

	carBytes, err := generateCAR(ctx, rootCID, nil)
	require.NoError(t, err)
	require.NotEmpty(t, carBytes)

	// The CAR must be parseable and carry the correct root CID even with no blocks.
	reader, err := car.NewBlockReader(bytes.NewReader(carBytes))
	require.NoError(t, err)

	require.Len(t, reader.Roots, 1)
	require.Equal(t, rootCID, reader.Roots[0])
}
