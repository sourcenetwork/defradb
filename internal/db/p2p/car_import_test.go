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
	car "github.com/ipld/go-car/v2"
	"github.com/ipld/go-car/v2/storage"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/immutable"

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

// rootstoreDB satisfies DB for the import path, which reaches no other method. Any other
// call panics on the nil embedded interface, which is the intent: it keeps the stub from
// silently growing a second responsibility.
type rootstoreDB struct {
	DB
	store corekv.TxnStore
}

func (d rootstoreDB) Rootstore() corekv.TxnStore { return d.store }

// carWith returns a single-block CAR rooted at that block.
func carWith(t *testing.T, block blocks.Block) []byte {
	t.Helper()
	var buf bytes.Buffer
	w, err := storage.NewWritable(&buf, []cid.Cid{block.Cid()}, car.WriteAsCarV1(true))
	require.NoError(t, err)
	require.NoError(t, w.Put(context.Background(), block.Cid().KeyString(), block.RawData()))
	require.NoError(t, w.Finalize())
	return buf.Bytes()
}

// A CAR carries blocks the node already holds, because field blocks are content-addressed
// and recur across documents sharing a field value. Importing one must not re-stamp a
// to-merge marker over a block that already merged: the marker would read as an abandoned
// fetch, and the orphan sweep repairs or removes on that basis.
func TestImportCAR_DoesNotUnmergeABlockTheNodeAlreadyHas(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	p := &P2P{db: rootstoreDB{store: rootstore}}

	coreBlock := coreblock.Block{Delta: crdt.CRDT{DocCompositeDelta: &crdt.DocCompositeDelta{Status: 1}}}
	raw, err := coreBlock.Marshal()
	require.NoError(t, err)
	link, err := coreblock.GetLinkFromNode(coreBlock.GenerateNode())
	require.NoError(t, err)
	block, err := blocks.NewBlockWithCid(raw, link.Cid)
	require.NoError(t, err)

	// The node already holds it, merged: stored, with no to-merge marker.
	bstore := datastore.BlockstoreFrom(rootstore, immutable.None[int]())
	require.NoError(t, bstore.Put(ctx, block))
	require.NoError(t, bstore.MarkAsMerged(ctx, block.Cid()))
	merged, err := bstore.IsMerged(ctx, block.Cid())
	require.NoError(t, err)
	require.True(t, merged, "precondition")

	imported, err := p.importCAR(ctx, carWith(t, block))
	require.NoError(t, err)
	require.NotNil(t, imported)

	merged, err = bstore.IsMerged(ctx, block.Cid())
	require.NoError(t, err)
	require.True(t, merged, "importing a block the node already holds must not unmerge it")
}
