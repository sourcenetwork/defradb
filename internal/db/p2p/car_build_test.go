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
	"context"
	"testing"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/immutable"

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

// carFixture returns a P2P whose store holds the given blocks and nothing else.
func carFixture(t *testing.T, held ...blocks.Block) *P2P {
	t.Helper()
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	bstore := datastore.BlockstoreFrom(rootstore, immutable.None[int]())
	for _, block := range held {
		require.NoError(t, bstore.Put(ctx, block))
	}
	return withReasonMaps(&P2P{db: rootstoreDB{store: rootstore}})
}

// compositeLinking returns a composite block linking to children, in the decoded form the CAR
// build takes and the encoded form the store holds.
func compositeLinking(t *testing.T, children ...cid.Cid) (*coreblock.Block, blocks.Block) {
	t.Helper()
	core := &coreblock.Block{
		Delta: crdt.CRDT{DocCompositeDelta: &crdt.DocCompositeDelta{Priority: 1}},
	}
	for i, child := range children {
		name := string(rune('a' + i))
		core.Links = append(core.Links, coreblock.NewDAGLink(name, cidlink.Link{Cid: child}))
	}
	raw, err := core.Marshal()
	require.NoError(t, err)
	link, err := core.GenerateLink()
	require.NoError(t, err)
	encoded, err := blocks.NewBlockWithCid(raw, link.Cid)
	require.NoError(t, err)
	return core, encoded
}

// undecodableBlock returns a block whose bytes are not a Block encoding, so the store reads it
// back but the walk cannot decode it.
func undecodableBlock(t *testing.T, raw string) blocks.Block {
	t.Helper()
	blockCID, err := coreblock.GetLinkPrototype().Sum([]byte(raw))
	require.NoError(t, err)
	block, err := blocks.NewBlockWithCid([]byte(raw), blockCID)
	require.NoError(t, err)
	return block
}

// A link the walk could not load is usually a block the write loop cannot read either, so the
// CAR is abandoned. Counting the link then reports a gap in a CAR nobody received.
func TestGenerateCAR_CountsNoMissingLinkWhenTheCARIsAbandoned(t *testing.T) {
	absent := compositeBlock(t, 2)
	root, encoded := compositeLinking(t, absent.Cid())
	p := carFixture(t, encoded)

	_, err := p.generateCAR(context.Background(), root)
	require.Error(t, err)
	require.Equal(t, map[string]int64{reasonBlockRead: 1}, reasonMap(p.carFailureReason.drain()))

	require.Equal(t, int64(1), p.statCARFailed.Load())
	require.Zero(t, p.statCARBuilt.Load())
	require.Zero(t, p.statCARMissing.Load())
}

// A block the store reads but the walk cannot decode ships in the CAR, and the counter records
// one per link the walk could not follow.
func TestGenerateCAR_CountsEveryMissingLinkWhenTheCARShips(t *testing.T) {
	first := undecodableBlock(t, "not a block")
	second := undecodableBlock(t, "also not a block")
	root, encoded := compositeLinking(t, first.Cid(), second.Cid())
	p := carFixture(t, encoded, first, second)

	_, err := p.generateCAR(context.Background(), root)
	require.NoError(t, err)

	require.Equal(t, int64(1), p.statCARBuilt.Load())
	require.Equal(t, int64(2), p.statCARMissing.Load())
}
