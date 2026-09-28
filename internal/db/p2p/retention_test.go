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
	"math"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/blockstore"
	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/lock"
	"github.com/sourcenetwork/defradb/internal/db/p2p/protocol"
)

// recordingBus records the names of published messages.
type recordingBus struct {
	event.Bus
	published []event.Name
}

func (b *recordingBus) Publish(msg event.Message) { b.published = append(b.published, msg.Name) }

// gateRig is a P2P receiving pushes. network holds the blocks peers serve over bitswap, and cars the
// CARs they serve.
type gateRig struct {
	p       *P2P
	db      *ingestDB
	bus     *recordingBus
	network datastore.Blockstore
	cars    *fakeCARChannel
}

// newGateRig returns a rig whose rule judges blockNumber against floor, or nothing when floor is nil.
func newGateRig(t *testing.T, floor *int64) *gateRig {
	t.Helper()
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	bus := &recordingBus{}
	db := &ingestDB{
		multistoreDB: multistoreDB{stores: datastore.NewMultistore(store, lock.NewLockSet(), immutable.None[int]())},
		store:        store,
		bus:          bus,
	}
	network := datastore.BlockstoreFrom(memory.NewDatastore(ctx), immutable.None[int]())
	cars := &fakeCARChannel{replies: map[string][]protocol.CARReply{}}
	p := withReasonMaps(&P2P{
		host:                 ipldHost{store: blockstore.NewIPLDStore(network)},
		carProtocol:          cars,
		db:                   db,
		processQueue:         newProcessQueue(),
		syncBlockLinkTimeout: time.Second,
		retentionFloor: func(context.Context, string) (string, int64, bool) {
			if floor == nil {
				return "", 0, false
			}
			return "blockNumber", *floor, true
		},
	})
	t.Cleanup(p.processQueue.close)
	return &gateRig{p: p, db: db, bus: bus, network: network, cars: cars}
}

// gateDoc is a document's create head and the height block it links under blockNumber.
type gateDoc struct {
	head   blocks.Block
	height blocks.Block
}

// newGateDoc returns a document holding value in blockNumber. Documents at one height share their
// height block; index keeps them distinct.
func newGateDoc(t *testing.T, value client.NormalValue, priority uint64, index int64) gateDoc {
	t.Helper()
	_, height := fieldBlock(t, "blockNumber", value)
	_, position := fieldBlock(t, "logIndex", client.NewNormalInt(index))
	head := &coreblock.Block{
		Delta: crdt.CRDT{DocCompositeDelta: &crdt.DocCompositeDelta{Priority: priority}},
		Links: []coreblock.DAGLink{
			coreblock.NewDAGLink("blockNumber", cidlink.Link{Cid: height.Cid()}),
			coreblock.NewDAGLink("logIndex", cidlink.Link{Cid: position.Cid()}),
		},
	}
	return gateDoc{head: encodeBlock(t, head), height: height}
}

func (d gateDoc) car(t *testing.T) []byte {
	return carWith(t, d.head.Cid(), d.head, d.height)
}

func (d gateDoc) info(t *testing.T, withCAR bool) protocol.DocumentInfo {
	info := protocol.DocumentInfo{DocID: d.head.Cid().String(), CID: d.head.Cid().Bytes(), Block: d.head.RawData()}
	if withCAR {
		info.CAR = d.car(t)
	}
	return info
}

// push delivers docs as one message: a single-document push when single is set, a batch otherwise.
// Only a single push returns its document's error.
func (r *gateRig) push(t *testing.T, single bool, docs ...protocol.DocumentInfo) error {
	t.Helper()
	req := &protocol.PushLogRequest{CollectionID: "col", Creator: "creator", Documents: docs}
	if single {
		require.Len(t, docs, 1)
		d := docs[0]
		req = &protocol.PushLogRequest{
			DocID: d.DocID, CollectionID: "col", Creator: "creator", CID: d.CID, Block: d.Block, CAR: d.CAR,
		}
	}
	return r.p.processPushlogRequest(context.Background(), req, true)
}

func floorOf(h int64) *int64 { return &h }

func pathName(single bool) string {
	if single {
		return "single"
	}
	return "batch"
}

// The gate refuses a document at the first source holding its height, and fetches or writes nothing
// after it.
func TestRetentionGateRefusesOnTheFirstHeightItCanRead(t *testing.T) {
	for _, tc := range []struct {
		name string
		// value is the document's blockNumber; nil means 100, the floor.
		value client.NormalValue
		// setup prepares the rig and returns what to push.
		setup          func(t *testing.T, r *gateRig, doc gateDoc) []protocol.DocumentInfo
		wantReason     string
		wantCARAsks    int
		wantHeightAsks int64
	}{
		{
			name:       "in the CAR sent with it",
			setup:      sendCARs,
			wantReason: "retentionBeforeFetch",
		},
		{
			name: "in the local store",
			setup: func(t *testing.T, r *gateRig, doc gateDoc) []protocol.DocumentInfo {
				require.NoError(t, r.db.Multistore().Blockstore().Put(context.Background(), doc.height))
				return []protocol.DocumentInfo{doc.info(t, false)}
			},
			wantReason: "retentionBeforeFetch",
		},
		{
			name: "in its fetched CAR",
			setup: func(t *testing.T, r *gateRig, doc gateDoc) []protocol.DocumentInfo {
				r.cars.replies["creator"] = []protocol.CARReply{{CARs: [][]byte{doc.car(t)}}}
				return []protocol.DocumentInfo{doc.info(t, false)}
			},
			wantReason:  "retentionAfterCARFetch",
			wantCARAsks: 1,
		},
		{
			name: "in its height block, fetched on its own",
			setup: func(t *testing.T, r *gateRig, doc gateDoc) []protocol.DocumentInfo {
				require.NoError(t, r.network.Put(context.Background(), doc.height))
				return []protocol.DocumentInfo{doc.info(t, false)}
			},
			wantReason:     "retentionAfterHeightFetch",
			wantCARAsks:    1,
			wantHeightAsks: 1,
		},
		{
			name:       "a value that is not a height",
			value:      client.NewNormalString("100"),
			setup:      sendCARs,
			wantReason: "retentionNoHeight",
		},
	} {
		for _, single := range []bool{true, false} {
			t.Run(tc.name+"/"+pathName(single), func(t *testing.T) {
				r := newGateRig(t, floorOf(100))
				value := tc.value
				if value == nil {
					value = client.NewNormalInt(100)
				}
				doc := newGateDoc(t, value, 1, 0)

				require.NoError(t, r.push(t, single, tc.setup(t, r, doc)...))

				require.Equal(t, map[string]int64{tc.wantReason: 1}, reasonMap(r.p.docSkipReason.drain()))
				require.Len(t, r.cars.requested, tc.wantCARAsks)
				require.Equal(t, tc.wantHeightAsks, r.p.statHeightFetches.Load())
				require.Zero(t, r.p.statSyncDAGCalls.Load(), "a refused document is not walked")
				held, err := r.db.Multistore().Blockstore().Has(context.Background(), doc.head.Cid())
				require.NoError(t, err)
				require.False(t, held, "a refused document writes nothing")
				require.Empty(t, r.db.merged)
			})
		}
	}
}

// sendCARs pushes doc with its CAR.
func sendCARs(t *testing.T, _ *gateRig, doc gateDoc) []protocol.DocumentInfo {
	return []protocol.DocumentInfo{doc.info(t, true)}
}

// A CAR fetched for one document supplies the shared height block to a sibling judged before it.
func TestRetentionGateReadsTheHeightFromASiblingsCAR(t *testing.T) {
	r := newGateRig(t, floorOf(100))
	unserved := newGateDoc(t, client.NewNormalInt(100), 1, 0)
	served := newGateDoc(t, client.NewNormalInt(100), 1, 1)
	require.Equal(t, served.height.Cid(), unserved.height.Cid())
	r.cars.replies["creator"] = []protocol.CARReply{{CARs: [][]byte{nil, served.car(t)}}}

	require.NoError(t, r.push(t, false, unserved.info(t, false), served.info(t, false)))

	require.Equal(t, map[string]int64{"retentionAfterCARFetch": 2}, reasonMap(r.p.docSkipReason.drain()))
	require.Zero(t, r.p.statHeightFetches.Load())
	require.Empty(t, r.db.merged)
}

// A document refused in a batch leaves the rest of the batch to go on with their own CARs.
func TestRetentionGateKeepsTheRestOfABatch(t *testing.T) {
	r := newGateRig(t, floorOf(100))
	old := newGateDoc(t, client.NewNormalInt(100), 1, 0)
	current := newGateDoc(t, client.NewNormalInt(101), 1, 1)
	require.NoError(t, r.db.Multistore().Blockstore().Put(context.Background(), old.height))
	r.cars.replies["creator"] = []protocol.CARReply{{CARs: [][]byte{current.car(t)}}}

	require.NoError(t, r.push(t, false, old.info(t, false), current.info(t, false)))

	require.Equal(t, map[string]int64{"retentionBeforeFetch": 1}, reasonMap(r.p.docSkipReason.drain()))
	require.Equal(t, []carCall{{peer: "creator", heads: 1}}, r.cars.requested)
	require.Len(t, r.db.merged, 1)
	require.Equal(t, current.head.Cid(), r.db.merged[0].Cid)
	require.Zero(t, r.p.statSyncDAGCalls.Load())
}

// The gate lets through what it cannot refuse.
func TestRetentionGateLeavesWhatItCannotRefuse(t *testing.T) {
	for _, tc := range []struct {
		name  string
		floor *int64
		doc   func(t *testing.T) gateDoc
		// setup prepares the rig and returns what to push.
		setup           func(t *testing.T, r *gateRig, doc gateDoc) []protocol.DocumentInfo
		wantMerged      bool
		wantWalked      bool
		wantFetchMissed int64
	}{
		{
			name:       "above the floor",
			floor:      floorOf(100),
			doc:        func(t *testing.T) gateDoc { return newGateDoc(t, client.NewNormalInt(101), 1, 0) },
			setup:      sendCARs,
			wantMerged: true,
		},
		{
			name:       "not a create",
			floor:      floorOf(100),
			doc:        func(t *testing.T) gateDoc { return newGateDoc(t, client.NewNormalInt(100), 2, 0) },
			setup:      sendCARs,
			wantMerged: true,
		},
		{
			name:       "collection not judged",
			doc:        func(t *testing.T) gateDoc { return newGateDoc(t, client.NewNormalInt(100), 1, 0) },
			setup:      sendCARs,
			wantMerged: true,
		},
		{
			name:  "height block no peer serves",
			floor: floorOf(100),
			doc:   func(t *testing.T) gateDoc { return newGateDoc(t, client.NewNormalInt(100), 1, 0) },
			setup: func(t *testing.T, _ *gateRig, doc gateDoc) []protocol.DocumentInfo {
				return []protocol.DocumentInfo{doc.info(t, false)}
			},
			wantWalked:      true,
			wantFetchMissed: 1,
		},
	} {
		for _, single := range []bool{true, false} {
			t.Run(tc.name+"/"+pathName(single), func(t *testing.T) {
				r := newGateRig(t, tc.floor)
				doc := tc.doc(t)

				err := r.push(t, single, tc.setup(t, r, doc)...)

				require.Empty(t, reasonMap(r.p.docSkipReason.drain()))
				require.Equal(t, tc.wantMerged, len(r.db.merged) == 1)
				require.Equal(t, tc.wantWalked, r.p.statSyncDAGCalls.Load() == 1)
				require.Equal(t, tc.wantFetchMissed, r.p.statHeightFetchMissed.Load())
				if tc.wantWalked {
					// The walk fails on the same missing block.
					require.Equal(t, int64(1), r.p.statDroppedDocs.Load())
				} else {
					require.NoError(t, err)
				}
			})
		}
	}
}

// Documents sharing an unavailable height block fetch it once.
func TestRetentionGateAsksForASharedHeightBlockOnce(t *testing.T) {
	r := newGateRig(t, floorOf(100))
	first := newGateDoc(t, client.NewNormalInt(100), 1, 0)
	second := newGateDoc(t, client.NewNormalInt(100), 1, 1)

	require.NoError(t, r.push(t, false, first.info(t, false), second.info(t, false)))

	require.Equal(t, int64(1), r.p.statHeightFetches.Load())
	require.Equal(t, int64(1), r.p.statHeightFetchMissed.Load())
	require.Equal(t, int64(2), r.p.statSyncDAGCalls.Load())
}

// A refusal at merge is counted as a skip, and the document is not relayed.
func TestRetentionRejectionAtMergeIsASkip(t *testing.T) {
	for _, single := range []bool{true, false} {
		t.Run(pathName(single), func(t *testing.T) {
			r := newGateRig(t, nil)
			r.db.mergeErr = client.ErrBelowRetentionFloor
			doc := newGateDoc(t, client.NewNormalInt(100), 1, 0)

			require.NoError(t, r.push(t, single, doc.info(t, true)))

			require.Equal(t, map[string]int64{"retentionAtMerge": 1}, reasonMap(r.p.docSkipReason.drain()))
			require.Zero(t, r.p.statDroppedDocs.Load(), "a refusal is not a loss")
			require.Zero(t, r.p.statBatchesWithDrops.Load())
			require.Zero(t, r.p.statMergedDocs.Load())
			require.NotContains(t, r.bus.published, event.UpdateName, "a refused document is not relayed")
		})
	}
}

func TestHeightIn(t *testing.T) {
	lww := func(data []byte, encryption *cidlink.Link) blocks.Block {
		return encodeBlock(t, &coreblock.Block{
			Delta:      crdt.CRDT{LWWDelta: &crdt.LWWDelta{FieldName: "blockNumber", Priority: 1, Data: data}},
			Encryption: encryption,
		})
	}
	sevenData, err := client.NewFieldValue(client.LWW_REGISTER, client.NewNormalInt(7)).Bytes()
	require.NoError(t, err)
	beyondInt64Data, err := cbor.Marshal(uint64(math.MaxInt64) + 1)
	require.NoError(t, err)

	_, height := fieldBlock(t, "blockNumber", client.NewNormalInt(7))
	_, eight := fieldBlock(t, "blockNumber", client.NewNormalInt(8))
	_, other := fieldBlock(t, "logIndex", client.NewNormalInt(7))
	_, negative := fieldBlock(t, "blockNumber", client.NewNormalInt(-1))
	_, text := fieldBlock(t, "blockNumber", client.NewNormalString("7"))
	beyondInt64 := lww(beyondInt64Data, nil)
	undecodable := lww([]byte{0xff}, nil)
	encrypted := lww(sevenData, &cidlink.Link{Cid: height.Cid()})
	counter := encodeBlock(t, &coreblock.Block{
		Delta: crdt.CRDT{CounterDelta: &crdt.CounterDelta{FieldName: "blockNumber", Priority: 1, Data: sevenData}},
	})
	notABlock := undecodableBlock(t, "not a block")

	for _, tc := range []struct {
		name      string
		raw       []byte
		link      cid.Cid
		wantH     int64
		wantValid bool
		wantOK    bool
	}{
		{name: "a height", raw: height.RawData(), link: height.Cid(), wantH: 7, wantValid: true, wantOK: true},
		{name: "a negative value", raw: negative.RawData(), link: negative.Cid(), wantOK: true},
		{name: "a value that is not an integer", raw: text.RawData(), link: text.Cid(), wantOK: true},
		{name: "a value beyond int64", raw: beyondInt64.RawData(), link: beyondInt64.Cid(), wantOK: true},
		{name: "a value that does not decode", raw: undecodable.RawData(), link: undecodable.Cid()},
		{name: "an encrypted value", raw: encrypted.RawData(), link: encrypted.Cid()},
		{name: "a counter", raw: counter.RawData(), link: counter.Cid()},
		{name: "bytes that are not a block", raw: notABlock.RawData(), link: notABlock.Cid()},
		{name: "another field", raw: other.RawData(), link: other.Cid()},
		{name: "bytes that do not match the link", raw: eight.RawData(), link: height.Cid()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, valid, ok := heightIn(tc.raw, tc.link, "blockNumber")
			require.Equal(t, tc.wantOK, ok)
			require.Equal(t, tc.wantValid, valid)
			require.Equal(t, tc.wantH, h)
		})
	}
}
