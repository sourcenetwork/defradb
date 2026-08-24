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
	"strconv"
	"testing"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/blockstore"
	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/lock"
	"github.com/sourcenetwork/defradb/internal/db/p2p/protocol"
)

// heldDocuments returns a P2P whose blockstore already holds n merged blocks, and the
// batch entries that carry them.
func heldDocuments(t *testing.T, n int) (*P2P, []protocol.DocumentInfo) {
	t.Helper()
	ctx := context.Background()
	stores := datastore.NewMultistore(memory.NewDatastore(ctx), lock.NewLockSet(), immutable.None[int]())

	docs := make([]protocol.DocumentInfo, 0, n)
	for i := range n {
		cb := coreblock.Block{Delta: crdt.CRDT{DocCompositeDelta: &crdt.DocCompositeDelta{
			Priority: uint64(i + 1),
		}}}
		raw, err := cb.Marshal()
		require.NoError(t, err)
		link, err := coreblock.GetLinkFromNode(cb.GenerateNode())
		require.NoError(t, err)
		block, err := blocks.NewBlockWithCid(raw, link.Cid)
		require.NoError(t, err)

		require.NoError(t, stores.Blockstore().Put(ctx, block))
		require.NoError(t, stores.Blockstore().MarkAsMerged(ctx, block.Cid()))

		docs = append(docs, protocol.DocumentInfo{DocID: "held-" + strconv.Itoa(i), CID: block.Cid().Bytes()})
	}
	return &P2P{db: multistoreDB{stores: stores}}, docs
}

// badDocuments returns batch entries whose CIDs do not parse. That is the first check a
// document meets, and it needs no store.
func badDocuments(n int) []protocol.DocumentInfo {
	docs := make([]protocol.DocumentInfo, 0, n)
	for i := range n {
		docs = append(docs, protocol.DocumentInfo{
			DocID: "bad-" + strconv.Itoa(i),
			CID:   []byte{0xff, 0xff, 0xff},
		})
	}
	return docs
}

// The drop counts are read against the batch count, so a batch that loses every document
// it carries still has to be counted.
func TestBatchIsCountedWhenNoDocumentSurvives(t *testing.T) {
	p := &P2P{}
	req := &protocol.PushLogRequest{CollectionID: "col", Documents: badDocuments(40)}

	require.NoError(t, p.processPushlogRequest(context.Background(), req, false))

	require.Equal(t, int64(40), p.statDroppedDocs.Load(), "drops")
	require.Equal(t, int64(1), p.statBatches.Load(), "the batch counts even though nothing survived it")
	require.Equal(t, int64(1), p.statBatchesWithDrops.Load(), "a drop before the merge still marks the batch")
}

// Documents the node already holds are skipped, not lost. Marking such a batch as one that
// dropped documents would report ordinary duplicate delivery as data loss.
func TestBatchOfHeldDocumentsCountsNoDrops(t *testing.T) {
	p, docs := heldDocuments(t, 5)
	req := &protocol.PushLogRequest{CollectionID: "col", Documents: docs}

	require.NoError(t, p.processPushlogRequest(context.Background(), req, false))

	require.Equal(t, int64(5), p.statSkippedDocs.Load(), "skips")
	require.Equal(t, int64(0), p.statDroppedDocs.Load(), "a held document is not a loss")
	require.Equal(t, int64(1), p.statBatches.Load(), "the batch counts")
	require.Equal(t, int64(0), p.statBatchesWithDrops.Load(), "no document dropped")
}

// Every document a batch carries ends merged, skipped or dropped. If the three do not add
// up to what arrived, an outcome went unrecorded and the drop count understates the loss.
func TestBatchAccountsForEveryDocument(t *testing.T) {
	const held, bad = 5, 3
	p, docs := heldDocuments(t, held)
	req := &protocol.PushLogRequest{CollectionID: "col", Documents: append(docs, badDocuments(bad)...)}

	require.NoError(t, p.processPushlogRequest(context.Background(), req, false))

	accounted := p.statMergedDocs.Load() + p.statDroppedDocs.Load() + p.statSkippedDocs.Load()
	require.Equal(t, int64(len(req.Documents)), accounted, "merged + dropped + skipped must equal received")
	require.Equal(t, int64(bad), p.statDroppedDocs.Load())
	require.Equal(t, int64(held), p.statSkippedDocs.Load())
	require.Equal(t, int64(1), p.statBatchesWithDrops.Load(), "the batch carried a drop")
}

// A single-document push ends in the same outcomes as one carried in a batch, and is
// counted under the same reasons.
func TestSingleDocumentOutcomesAreCounted(t *testing.T) {
	t.Run("a CID that does not parse is a drop", func(t *testing.T) {
		p := &P2P{}
		req := &protocol.PushLogRequest{DocID: "d", CollectionID: "col", CID: []byte{0xff, 0xff, 0xff}}

		require.Error(t, p.processPushlogRequest(context.Background(), req, false))
		require.Equal(t, int64(1), p.statDroppedDocs.Load())
	})

	t.Run("a document already held is a skip", func(t *testing.T) {
		p, docs := heldDocuments(t, 1)
		p.processQueue = newProcessQueue()
		defer p.processQueue.close()
		req := &protocol.PushLogRequest{DocID: docs[0].DocID, CollectionID: "col", CID: docs[0].CID}

		require.NoError(t, p.processPushlogRequest(context.Background(), req, false))
		require.Equal(t, int64(1), p.statSkippedDocs.Load())
		require.Equal(t, int64(0), p.statDroppedDocs.Load(), "a held document is not a loss")
	})
}

// emptyIPLDHost satisfies Host for the doc-sync path, backed by a store holding no blocks
// so a DAG walk fails to load its root.
type emptyIPLDHost struct {
	client.Host
	store blockstore.IPLDStore
}

func (h emptyIPLDHost) ID() string                      { return "peerID" }
func (h emptyIPLDHost) IPLDStore() blockstore.IPLDStore { return h.store }

func newEmptyIPLDHost(ctx context.Context) emptyIPLDHost {
	bs := datastore.BlockstoreFrom(memory.NewDatastore(ctx), immutable.None[int]())
	return emptyIPLDHost{store: blockstore.NewIPLDStore(bs)}
}

// Documents pulled from a peer reach the store by a different route than a pushed batch,
// but end in the same outcomes and belong in the same counters.
func TestDocSyncOutcomesAreCounted(t *testing.T) {
	ctx := context.Background()

	t.Run("a head that does not parse is a drop", func(t *testing.T) {
		p := &P2P{}
		item := docSyncItem{DocID: "d", Heads: [][]byte{{0xff, 0xff, 0xff}}}

		p.handleDocSyncItem(ctx, item, "sender", "col", map[string][]cid.Cid{})

		require.Equal(t, int64(1), p.statDroppedDocs.Load())
		require.Equal(t, int64(0), p.statSkippedDocs.Load())
	})

	t.Run("a head already seen in the response is a skip", func(t *testing.T) {
		p := &P2P{}
		_, head, err := cid.CidFromBytes(blocks.NewBlock([]byte("a block")).Cid().Bytes())
		require.NoError(t, err)
		item := docSyncItem{DocID: "d", Heads: [][]byte{head.Bytes()}}

		p.handleDocSyncItem(ctx, item, "sender", "col", map[string][]cid.Cid{"d": {head}})

		require.Equal(t, int64(1), p.statSkippedDocs.Load(), "a repeated head is not a loss")
		require.Equal(t, int64(0), p.statDroppedDocs.Load())
	})

	t.Run("a DAG that cannot be walked is a drop", func(t *testing.T) {
		p := &P2P{host: newEmptyIPLDHost(ctx)}
		head := blocks.NewBlock([]byte("a block")).Cid()
		item := docSyncItem{DocID: "d", Heads: [][]byte{head.Bytes()}}

		p.handleDocSyncItem(ctx, item, "sender", "col", map[string][]cid.Cid{})

		require.Equal(t, int64(1), p.statDroppedDocs.Load(), "a document that could not be fetched is a loss")
		require.Equal(t, int64(0), p.statMergedDocs.Load())
	})
}
