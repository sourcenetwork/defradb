// Copyright 2024 Democratized Data Foundation
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
	"testing"

	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/blockstore"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

const counterSchema = `
type Tally {
	hits: Int @crdt(type: pcounter)
}
`

func tallyHits(ctx context.Context, t *testing.T, col client.Collection, docID client.DocID) int64 {
	doc, err := col.GetDocument(ctx, docID)
	require.NoError(t, err)
	m, err := doc.ToMap()
	require.NoError(t, err)
	v, _ := m["hits"].(int64)
	return v
}

// Merging from a *field-block* CID as the head must be idempotent: a counter
// field delta must not be re-applied when its block arrives as a merge head.
//
// Repro for the counter multiply-apply bug: `coreblock.ProcessBlock` applies the
// delta unconditionally (no per-block `IsMerged` guard), so when `executeMerge`
// is invoked with a field-block CID (as happens when a peer pushes the DAG
// block-by-block — every block as its own pushLogRequest), `loadComposites`
// collects that field block into the processing list and `processBlock` applies
// its counter delta a SECOND time (the first application happened via the field
// block's owning composite). The materialized counter is then double-counted.
//
// This test merges a single +5 increment via its composite head (value -> 5),
// then merges from the field-block head and asserts the value is still 5. On
// current code it becomes 10.
func TestMerge_CounterFieldBlockAsHead_IsNotReapplied(t *testing.T) {
	ctx := context.Background()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, counterSchema)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "Tally")
	require.NoError(t, err)

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(blockstore.NewIPLDStore(datastore.BlockstoreFrom(db.rootstore, immutable.None[int]())))

	doc, err := client.NewDocFromMap(ctx, map[string]any{"hits": int64(5)}, col.Version())
	require.NoError(t, err)
	docID := []byte(doc.ID().String())

	// One counter increment of +5: a field block (priority 1) linked by a
	// composite block (priority 1).
	fieldBlock := coreblock.Block{
		Delta: crdt.CRDT{CounterDelta: &crdt.CounterDelta{
			DocID:               docID,
			FieldName:           "hits",
			Priority:            1,
			CollectionVersionID: col.Version().VersionID,
			Data:                encodeValue(int64(5)),
		}},
	}
	fieldLink, err := lsys.Store(ipld.LinkContext{}, coreblock.GetLinkPrototype(), fieldBlock.GenerateNode())
	require.NoError(t, err)

	compositeBlock := coreblock.New(
		crdt.NewCRDT(&crdt.DocCompositeDelta{
			DocID:               docID,
			Priority:            1,
			CollectionVersionID: col.Version().VersionID,
			Status:              1,
		}),
		[]coreblock.DAGLink{{Name: "hits", Link: fieldLink.(cidlink.Link)}},
	)
	compositeLink, err := lsys.Store(ipld.LinkContext{}, coreblock.GetLinkPrototype(), compositeBlock.GenerateNode())
	require.NoError(t, err)

	// Normal merge from the composite head: the counter accumulates to 5.
	err = db.executeMerge(ctx, col.(*collection), event.Merge{
		DocID:        doc.ID().String(),
		Cid:          compositeLink.(cidlink.Link).Cid,
		CollectionID: col.CollectionID(),
	})
	require.NoError(t, err)
	require.Equal(t, int64(5), tallyHits(ctx, t, col, doc.ID()))

	// A peer that pushes the document DAG block-by-block delivers the field block
	// as its own pushLogRequest head. Merging from that field-block head must be
	// idempotent (the increment was already applied via the composite above).
	err = db.executeMerge(ctx, col.(*collection), event.Merge{
		DocID:        doc.ID().String(),
		Cid:          fieldLink.(cidlink.Link).Cid,
		CollectionID: col.CollectionID(),
	})
	require.NoError(t, err)

	require.Equal(t, int64(5), tallyHits(ctx, t, col, doc.ID()),
		"counter field delta was re-applied when its block arrived as a merge head (double-count)")
}
