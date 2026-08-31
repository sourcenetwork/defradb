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
	"sync"
	"sync/atomic"
	"testing"
	"time"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"
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

// blockingStore holds reads open once armed, so a test can keep one arrival inside the
// handler while further arrivals are made.
type blockingStore struct {
	corekv.TxnStore

	armed   atomic.Bool
	once    sync.Once
	reached chan struct{}
	release chan struct{}
}

func (s *blockingStore) Has(ctx context.Context, key []byte) (bool, error) {
	if s.armed.Load() {
		s.once.Do(func() { close(s.reached) })
		<-s.release
	}
	return s.TxnStore.Has(ctx, key)
}

// pushRequest returns a single-document request for head whose block does not decode, so an
// arrival that reaches the handler ends as a drop.
func pushRequest(docID string, head cid.Cid) *protocol.PushLogRequest {
	return &protocol.PushLogRequest{
		DocID:        docID,
		CollectionID: "col",
		CID:          head.Bytes(),
		Block:        []byte("not a block"),
	}
}

// awaitClose fails the test rather than hanging the package, so a change that makes an
// arrival wait on another reports which wait it was.
func awaitClose(t *testing.T, c <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-c:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for %s", what)
	}
}

// The counters record arrivals, not documents. An arrival for a document already in flight is
// deduplicated on its head, so it is a skip while the arrival that is running reports its own
// outcome later. An arrival for a different head is not deduplicated.
func TestArrivalForAHeadAlreadyInFlightIsSkipped(t *testing.T) {
	ctx := context.Background()
	store := &blockingStore{
		TxnStore: memory.NewDatastore(ctx),
		reached:  make(chan struct{}),
		release:  make(chan struct{}),
	}
	stores := datastore.NewMultistore(store, lock.NewLockSet(), immutable.None[int]())
	p := &P2P{db: multistoreDB{stores: stores}, processQueue: newProcessQueue()}
	t.Cleanup(p.processQueue.close)

	// Released on any exit, so a failed assertion cannot leave an arrival parked.
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(store.release) }) }
	defer release()

	// Two heads of one document, so the queue key is what decides whether the second is
	// deduplicated against the first.
	held := pushRequest("d", blocks.NewBlock([]byte("held head")).Cid())
	other := pushRequest("d", blocks.NewBlock([]byte("other head")).Cid())

	store.armed.Store(true)
	running := make(chan error, 2)
	go func() { running <- p.processPushlogRequest(ctx, held, false) }()
	awaitClose(t, store.reached, "the first arrival to reach the store")

	require.NoError(t, p.processPushlogRequest(ctx, held, false), "a deduplicated arrival is not an error")
	go func() { running <- p.processPushlogRequest(ctx, other, false) }()

	require.Equal(t, map[string]int64{"inFlight": 1}, reasonMap(p.docSkipReason.drain()))
	require.Equal(t, int64(0), p.statDroppedDocs.Load(), "no arrival has reported an outcome yet")

	release()
	require.Error(t, <-running)
	require.Error(t, <-running)
	require.Equal(t, map[string]int64{"blockDecode": 2}, reasonMap(p.docDropReason.drain()),
		"the arrival for a different head was not deduplicated")
	require.Empty(t, reasonMap(p.docSkipReason.drain()), "only the repeated head was skipped")
}

// A document the sync queue refuses is one this node will not hold, so it counts as a loss.
func TestArrivalRefusedByAFullSyncQueueIsDropped(t *testing.T) {
	p := &P2P{
		// An unserved queue with no capacity refuses on arrival.
		processQueue: &processQueue{inFlight: map[string]struct{}{}, queue: make(chan syncRequest)},
	}
	req := pushRequest("d", blocks.NewBlock([]byte("a head")).Cid())

	// Refused twice, because a document the queue turned away is not left marked in flight.
	require.ErrorIs(t, p.processPushlogRequest(context.Background(), req, false), ErrSyncQueueFull)
	require.ErrorIs(t, p.processPushlogRequest(context.Background(), req, false), ErrSyncQueueFull)

	require.Equal(t, map[string]int64{"syncQueueFull": 2}, reasonMap(p.docDropReason.drain()))
	require.Empty(t, reasonMap(p.docSkipReason.drain()), "a refused document is not a skip")
}
