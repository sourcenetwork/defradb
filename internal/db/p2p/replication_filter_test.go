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
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"
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

// fieldBlock returns a field block holding value under name, decoded and encoded. Its delta is
// encoded as the write path encodes it.
func fieldBlock(t *testing.T, name string, value client.NormalValue) (*coreblock.Block, blocks.Block) {
	t.Helper()
	data, err := client.NewFieldValue(client.LWW_REGISTER, value).Bytes()
	require.NoError(t, err)

	core := &coreblock.Block{
		Delta: crdt.CRDT{LWWDelta: &crdt.LWWDelta{FieldName: name, Priority: 1, Data: data}},
	}
	return core, encodeBlock(t, core)
}

// encodeBlock returns the encoded form of a block, as a CAR carries it.
func encodeBlock(t *testing.T, block *coreblock.Block) blocks.Block {
	t.Helper()
	raw, err := block.Marshal()
	require.NoError(t, err)
	link, err := block.GenerateLink()
	require.NoError(t, err)
	encoded, err := blocks.NewBlockWithCid(raw, link.Cid)
	require.NoError(t, err)
	return encoded
}

// The values of a composite block's fields are read from the field blocks it links.
func TestExtractFieldsFromBlock_ReadsValuesFromTheLinkedFieldBlocks(t *testing.T) {
	_, number := fieldBlock(t, "blockNumber", client.NewNormalInt(25951718))
	_, address := fieldBlock(t, "address", client.NewNormalString("0xabc"))
	composite, root := compositeLinking(t, number.Cid(), address.Cid())

	fields := extractFieldsFromBlock(composite, carWith(t, root.Cid(), number, address))

	require.Equal(t, map[string]any{"blockNumber": uint64(25951718), "address": "0xabc"}, fields)
}

// An encrypted field block holds ciphertext, so its field is left out.
func TestExtractFieldsFromBlock_SkipsAnEncryptedFieldBlock(t *testing.T) {
	_, plain := fieldBlock(t, "address", client.NewNormalString("0xabc"))
	secret, _ := fieldBlock(t, "blockNumber", client.NewNormalInt(25951718))
	secret.Encryption = &cidlink.Link{Cid: plain.Cid()}
	encrypted := encodeBlock(t, secret)

	composite, root := compositeLinking(t, encrypted.Cid(), plain.Cid())

	fields := extractFieldsFromBlock(composite, carWith(t, root.Cid(), encrypted, plain))

	require.Equal(t, map[string]any{"address": "0xabc"}, fields)
}

// A field block whose value does not decode is left out.
func TestExtractFieldsFromBlock_SkipsAValueThatDoesNotDecode(t *testing.T) {
	_, plain := fieldBlock(t, "address", client.NewNormalString("0xabc"))
	// 0xff on its own is a CBOR break code with nothing to close, so it does not decode.
	broken := encodeBlock(t, &coreblock.Block{
		Delta: crdt.CRDT{LWWDelta: &crdt.LWWDelta{FieldName: "blockNumber", Priority: 1, Data: []byte{0xff}}},
	})

	composite, root := compositeLinking(t, broken.Cid(), plain.Cid())

	fields := extractFieldsFromBlock(composite, carWith(t, root.Cid(), broken, plain))

	require.Equal(t, map[string]any{"address": "0xabc"}, fields)
}

// heightFilter rejects a document whose blockNumber is at or below cutoff, and records the
// fields of every call.
type heightFilter struct {
	cutoff uint64
	calls  []map[string]any
}

func (f *heightFilter) AllowReplication(_ context.Context, _, _ string, fields map[string]any) bool {
	f.calls = append(f.calls, fields)
	height, ok := fields["blockNumber"].(uint64)
	return !ok || height > f.cutoff
}

// ingestDB provides the stores, merge and event bus the pushlog path uses.
type ingestDB struct {
	multistoreDB
	store corekv.TxnStore
	bus   event.Bus
}

func (d ingestDB) Rootstore() corekv.TxnStore { return d.store }

func (d ingestDB) Merge(context.Context, event.Merge) error { return nil }

func (d ingestDB) Events() event.Bus { return d.bus }

// A document received without its CAR is filtered again on the values in the fetched CAR, and one
// received with its CAR is filtered once. A rejected document writes nothing.
func TestArrivalIsFilteredOnTheFieldValuesInItsCAR(t *testing.T) {
	for _, tc := range []struct {
		name       string
		inline     bool
		cutoff     uint64
		wantCalls  []map[string]any
		wantSkips  map[string]int64
		wantStored bool
	}{
		{
			name:      "rejected on a fetched value",
			cutoff:    100,
			wantCalls: []map[string]any{{}, {"blockNumber": uint64(100)}},
			wantSkips: map[string]int64{"filteredAfterFetch": 1},
		},
		{
			name:       "allowed on a fetched value",
			cutoff:     99,
			wantCalls:  []map[string]any{{}, {"blockNumber": uint64(100)}},
			wantSkips:  map[string]int64{},
			wantStored: true,
		},
		{
			name:      "rejected on a value sent with its CAR",
			inline:    true,
			cutoff:    100,
			wantCalls: []map[string]any{{"blockNumber": uint64(100)}},
			wantSkips: map[string]int64{"filtered": 1},
		},
		{
			name:       "allowed on a value sent with its CAR",
			inline:     true,
			cutoff:     99,
			wantCalls:  []map[string]any{{"blockNumber": uint64(100)}},
			wantSkips:  map[string]int64{},
			wantStored: true,
		},
	} {
		for _, single := range []bool{true, false} {
			path := "batch"
			if single {
				path = "single"
			}
			t.Run(path+": "+tc.name, func(t *testing.T) {
				ctx := context.Background()
				_, number := fieldBlock(t, "blockNumber", client.NewNormalInt(100))
				_, root := compositeLinking(t, number.Cid())
				head := root.Cid()
				carData := carWith(t, head, root, number)

				p := fetchFixture(&fakeCARChannel{replies: map[string][]protocol.CARReply{
					"creator": {{CARs: [][]byte{carData}}},
				}})
				store := memory.NewDatastore(ctx)
				stores := datastore.NewMultistore(store, lock.NewLockSet(), immutable.None[int]())
				bus := event.NewChannelBus(0, 0)
				t.Cleanup(bus.Close)
				p.db = ingestDB{multistoreDB: multistoreDB{stores: stores}, store: store, bus: bus}
				filter := &heightFilter{cutoff: tc.cutoff}
				p.replicationFilter = filter

				doc := protocol.DocumentInfo{DocID: "d", CID: head.Bytes(), Block: root.RawData()}
				if tc.inline {
					doc.CAR = carData
				}
				if single {
					p.processQueue = newProcessQueue()
					t.Cleanup(p.processQueue.close)
					req := &protocol.PushLogRequest{
						DocID:        doc.DocID,
						CollectionID: "col",
						Creator:      "creator",
						CID:          doc.CID,
						Block:        doc.Block,
						CAR:          doc.CAR,
					}
					require.NoError(t, p.processPushlogRequest(ctx, req, true))
				} else {
					req := &protocol.PushLogRequest{
						CollectionID: "col", Creator: "creator", Documents: []protocol.DocumentInfo{doc},
					}
					_, _, err := p.processBatchedDocuments(ctx, req, true)
					require.NoError(t, err)
				}

				require.Equal(t, tc.wantCalls, filter.calls)
				require.Equal(t, tc.wantSkips, reasonMap(p.docSkipReason.drain()))
				stored, err := stores.Blockstore().Has(ctx, head)
				require.NoError(t, err)
				require.Equal(t, tc.wantStored, stored)
			})
		}
	}
}
