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
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
)

const userSchema = `
type User {
	name: String
	age: Int
}
`

func TestMerge_SingleBranch_NoError(t *testing.T) {
	ctx := context.Background()

	db, err := newDefraMemoryDB(ctx)
	require.NoError(t, err)

	_, err = db.AddSchema(ctx, userSchema)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(db.multistore.Blockstore().AsIPLDStorage())

	initialDocState := map[string]any{
		"name": "John",
	}
	d, docID := newDagBuilder(col, initialDocState)
	compInfo, err := d.generateCompositeUpdate(&lsys, initialDocState, compositeInfo{})
	require.NoError(t, err)
	compInfo2, err := d.generateCompositeUpdate(&lsys, map[string]any{"name": "Johny"}, compInfo)
	require.NoError(t, err)

	err = db.executeMerge(ctx, event.Merge{
		DocID:      docID.String(),
		Cid:        compInfo2.link.Cid,
		SchemaRoot: col.SchemaRoot(),
	})
	require.NoError(t, err)

	// Verify the document was created with the expected values
	doc, err := col.Get(ctx, docID, false)
	require.NoError(t, err)
	docMap, err := doc.ToMap()
	require.NoError(t, err)

	expectedDocMap := map[string]any{
		"_docID": docID.String(),
		"name":   "Johny",
	}

	require.Equal(t, expectedDocMap, docMap)
}

func TestMerge_DualBranch_NoError(t *testing.T) {
	ctx := context.Background()

	db, err := newDefraMemoryDB(ctx)
	require.NoError(t, err)

	_, err = db.AddSchema(ctx, userSchema)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(db.multistore.Blockstore().AsIPLDStorage())

	initialDocState := map[string]any{
		"name": "John",
	}
	d, docID := newDagBuilder(col, initialDocState)
	compInfo, err := d.generateCompositeUpdate(&lsys, initialDocState, compositeInfo{})
	require.NoError(t, err)
	compInfo2, err := d.generateCompositeUpdate(&lsys, map[string]any{"name": "Johny"}, compInfo)
	require.NoError(t, err)

	err = db.executeMerge(ctx, event.Merge{
		DocID:      docID.String(),
		Cid:        compInfo2.link.Cid,
		SchemaRoot: col.SchemaRoot(),
	})
	require.NoError(t, err)

	compInfo3, err := d.generateCompositeUpdate(&lsys, map[string]any{"age": 30}, compInfo)
	require.NoError(t, err)

	err = db.executeMerge(ctx, event.Merge{
		DocID:      docID.String(),
		Cid:        compInfo3.link.Cid,
		SchemaRoot: col.SchemaRoot(),
	})
	require.NoError(t, err)

	// Verify the document was created with the expected values
	doc, err := col.Get(ctx, docID, false)
	require.NoError(t, err)
	docMap, err := doc.ToMap()
	require.NoError(t, err)

	expectedDocMap := map[string]any{
		"_docID": docID.String(),
		"age":    int64(30),
		"name":   "Johny",
	}

	require.Equal(t, expectedDocMap, docMap)
}

// This test is not something we can reproduce in with integration tests.
// Until we introduce partial dag syncs to integration tests, this should not be removed.
func TestMerge_DualBranchWithOneIncomplete_CouldNotFindCID(t *testing.T) {
	ctx := context.Background()

	db, err := newDefraMemoryDB(ctx)
	require.NoError(t, err)

	_, err = db.AddSchema(ctx, userSchema)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(db.multistore.Blockstore().AsIPLDStorage())

	initialDocState := map[string]any{
		"name": "John",
	}
	d, docID := newDagBuilder(col, initialDocState)
	compInfo, err := d.generateCompositeUpdate(&lsys, initialDocState, compositeInfo{})
	require.NoError(t, err)
	compInfo2, err := d.generateCompositeUpdate(&lsys, map[string]any{"name": "Johny"}, compInfo)
	require.NoError(t, err)

	err = db.executeMerge(ctx, event.Merge{
		DocID:      docID.String(),
		Cid:        compInfo2.link.Cid,
		SchemaRoot: col.SchemaRoot(),
	})
	require.NoError(t, err)

	someUnknownBlock := coreblock.Block{Delta: crdt.CRDT{CompositeDAGDelta: &crdt.CompositeDAGDelta{Status: 1}}}
	someUnknownLink, err := coreblock.GetLinkFromNode(someUnknownBlock.GenerateNode())
	require.NoError(t, err)

	compInfoUnkown := compositeInfo{
		link:   someUnknownLink,
		height: 2,
	}

	compInfo3, err := d.generateCompositeUpdate(&lsys, map[string]any{"name": "Johny"}, compInfoUnkown)
	require.NoError(t, err)

	err = db.executeMerge(ctx, event.Merge{
		DocID:      docID.String(),
		Cid:        compInfo3.link.Cid,
		SchemaRoot: col.SchemaRoot(),
	})
	require.ErrorContains(t, err, "could not find bafyreifi4sa4auy4uk6psoljwuzqepgwqzsjk3h6p2xjdtsty7bdjz4uzm")

	// Verify the document was created with the expected values
	doc, err := col.Get(ctx, docID, false)
	require.NoError(t, err)
	docMap, err := doc.ToMap()
	require.NoError(t, err)

	expectedDocMap := map[string]any{
		"_docID": docID.String(),
		"name":   "Johny",
	}

	require.Equal(t, expectedDocMap, docMap)
}

type dagBuilder struct {
	fieldsHeight map[string]uint64
	docID        []byte
	col          client.Collection
}

func newDagBuilder(col client.Collection, initalDocState map[string]any) (*dagBuilder, client.DocID) {
	doc, err := client.NewDocFromMap(
		initalDocState,
		col.Definition(),
	)
	if err != nil {
		panic(err)
	}
	return &dagBuilder{
		fieldsHeight: make(map[string]uint64),
		docID:        []byte(doc.ID().String()),
		col:          col,
	}, doc.ID()
}

type compositeInfo struct {
	link   cidlink.Link
	height uint64
}

func (d *dagBuilder) generateCompositeUpdate(lsys *linking.LinkSystem, fields map[string]any, from compositeInfo) (compositeInfo, error) {
	heads := []cid.Cid{}
	newPriority := from.height + 1
	if from.link.ByteLen() != 0 {
		heads = append(heads, from.link.Cid)
	}

	links := []coreblock.DAGLink{}
	for field, val := range fields {
		d.fieldsHeight[field]++
		// Generate new Block and save to lsys
		fieldBlock := coreblock.Block{
			Delta: crdt.CRDT{
				LWWRegDelta: &crdt.LWWRegDelta{
					DocID:           d.docID,
					FieldName:       field,
					Priority:        d.fieldsHeight[field],
					SchemaVersionID: d.col.Schema().VersionID,
					Data:            encodeValue(val),
				},
			},
		}
		fieldBlockLink, err := lsys.Store(ipld.LinkContext{}, coreblock.GetLinkPrototype(), fieldBlock.GenerateNode())
		if err != nil {
			return compositeInfo{}, err
		}
		links = append(links, coreblock.DAGLink{
			Name: field,
			Link: fieldBlockLink.(cidlink.Link),
		})
	}

	compositeBlock := coreblock.New(
		&crdt.CompositeDAGDelta{
			DocID:           d.docID,
			Priority:        newPriority,
			SchemaVersionID: d.col.Schema().VersionID,
			Status:          1,
		},
		links,
		heads...,
	)

	compositeBlockLink, err := lsys.Store(ipld.LinkContext{}, coreblock.GetLinkPrototype(), compositeBlock.GenerateNode())
	if err != nil {
		return compositeInfo{}, err
	}

	return compositeInfo{
		link:   compositeBlockLink.(cidlink.Link),
		height: newPriority,
	}, nil
}

func encodeValue(val any) []byte {
	em, err := client.CborEncodingOptions().EncMode()
	if err != nil {
		// safe to panic here as this is a test
		panic(err)
	}
	b, err := em.Marshal(val)
	if err != nil {
		// safe to panic here as this is a test
		panic(err)
	}
	return b
}

func TestMergeQueue(t *testing.T) {
	q := newMergeQueue()

	testDocID := "test"

	q.add(testDocID)
	go q.add(testDocID)
	// give time for the goroutine to block
	time.Sleep(10 * time.Millisecond)
	require.Len(t, q.docs, 1)
	q.done(testDocID)
	// give time for the goroutine to add the docID
	time.Sleep(10 * time.Millisecond)
	q.mutex.Lock()
	require.Len(t, q.docs, 1)
	q.mutex.Unlock()
	q.done(testDocID)
	q.mutex.Lock()
	require.Len(t, q.docs, 0)
	q.mutex.Unlock()
}
