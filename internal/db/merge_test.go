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

	"github.com/sourcenetwork/corekv/blockstore"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

const userSchema = `
type User {
	name: String
	age: Int
}
`

const userSchemaWithCounter = `
type User {
	name: String
	points: Int @crdt(type: pncounter)
}
`

func TestMerge_SingleBranch_NoError(t *testing.T) {
	ctx := context.Background()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(blockstore.NewIPLDStore(datastore.BlockstoreFrom(db.rootstore, immutable.None[int]())))

	initialDocState := map[string]any{
		"name": "John",
	}
	d, docID := newDagBuilder(ctx, col, initialDocState)
	compInfo, err := d.generateCompositeUpdate(&lsys, initialDocState, compositeInfo{})
	require.NoError(t, err)
	compInfo2, err := d.generateCompositeUpdate(&lsys, map[string]any{"name": "Johny"}, compInfo)
	require.NoError(t, err)

	err = db.executeMerge(ctx, col.(*collection), event.Merge{
		DocID:        docID.String(),
		Cid:          compInfo2.link.Cid,
		CollectionID: col.CollectionID(),
	})
	require.NoError(t, err)

	// Verify the document was added with the expected values
	doc, err := col.GetDocument(ctx, docID)
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

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(blockstore.NewIPLDStore(datastore.BlockstoreFrom(db.rootstore, immutable.None[int]())))

	initialDocState := map[string]any{
		"name": "John",
	}
	d, docID := newDagBuilder(ctx, col, initialDocState)
	compInfo, err := d.generateCompositeUpdate(&lsys, initialDocState, compositeInfo{})
	require.NoError(t, err)
	compInfo2, err := d.generateCompositeUpdate(&lsys, map[string]any{"name": "Johny"}, compInfo)
	require.NoError(t, err)

	err = db.executeMerge(ctx, col.(*collection), event.Merge{
		DocID:        docID.String(),
		Cid:          compInfo2.link.Cid,
		CollectionID: col.CollectionID(),
	})
	require.NoError(t, err)

	compInfo3, err := d.generateCompositeUpdate(&lsys, map[string]any{"age": 30}, compInfo)
	require.NoError(t, err)

	err = db.executeMerge(ctx, col.(*collection), event.Merge{
		DocID:        docID.String(),
		Cid:          compInfo3.link.Cid,
		CollectionID: col.CollectionID(),
	})
	require.NoError(t, err)

	// Verify the document was added with the expected values
	doc, err := col.GetDocument(ctx, docID)
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

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(blockstore.NewIPLDStore(datastore.BlockstoreFrom(db.rootstore, immutable.None[int]())))

	initialDocState := map[string]any{
		"name": "John",
	}
	d, docID := newDagBuilder(ctx, col, initialDocState)
	compInfo, err := d.generateCompositeUpdate(&lsys, initialDocState, compositeInfo{})
	require.NoError(t, err)
	compInfo2, err := d.generateCompositeUpdate(&lsys, map[string]any{"name": "Johny"}, compInfo)
	require.NoError(t, err)

	err = db.executeMerge(ctx, col.(*collection), event.Merge{
		DocID:        docID.String(),
		Cid:          compInfo2.link.Cid,
		CollectionID: col.CollectionID(),
	})
	require.NoError(t, err)

	someUnknownBlock := coreblock.Block{Delta: crdt.CRDT{DocCompositeDelta: &crdt.DocCompositeDelta{Status: 1}}}
	someUnknownLink, err := coreblock.GetLinkFromNode(someUnknownBlock.GenerateNode())
	require.NoError(t, err)

	compInfoUnkown := compositeInfo{
		link:   someUnknownLink,
		height: 2,
	}

	compInfo3, err := d.generateCompositeUpdate(&lsys, map[string]any{"name": "Johny"}, compInfoUnkown)
	require.NoError(t, err)

	err = db.executeMerge(ctx, col.(*collection), event.Merge{
		DocID:        docID.String(),
		Cid:          compInfo3.link.Cid,
		CollectionID: col.CollectionID(),
	})
	require.ErrorContains(t, err, "could not find bafyreihs5kx5u6k6mc3m6st3ytam4e3mmk3sd6p4jn3hh5o63wpf4holoq")

	// Verify the document was added with the expected values
	doc, err := col.GetDocument(ctx, docID)
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

func newDagBuilder(ctx context.Context, col client.Collection, initalDocState map[string]any) (*dagBuilder, client.DocID) {
	doc, err := client.NewDocFromMap(
		ctx,
		initalDocState,
		col.Version(),
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
				LWWDelta: &crdt.LWWDelta{
					DocID:               d.docID,
					FieldName:           field,
					Priority:            d.fieldsHeight[field],
					CollectionVersionID: d.col.Version().VersionID,
					Data:                encodeValue(val),
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
		crdt.NewCRDT(&crdt.DocCompositeDelta{
			DocID:               d.docID,
			Priority:            newPriority,
			CollectionVersionID: d.col.Version().VersionID,
			Status:              1,
		}),
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

// generateCompositeUpdateFromHeads creates a composite block with multiple parents.
// This produces a merge/reconvergence point in the DAG.
func (d *dagBuilder) generateCompositeUpdateFromHeads(
	lsys *linking.LinkSystem,
	fields map[string]any,
	parents []compositeInfo,
) (compositeInfo, error) {
	var maxHeight uint64
	heads := []cid.Cid{}
	for _, p := range parents {
		if p.link.ByteLen() != 0 {
			heads = append(heads, p.link.Cid)
		}
		if p.height > maxHeight {
			maxHeight = p.height
		}
	}
	newPriority := maxHeight + 1

	links := []coreblock.DAGLink{}
	for field, val := range fields {
		d.fieldsHeight[field]++
		fieldBlock := coreblock.Block{
			Delta: crdt.CRDT{
				LWWDelta: &crdt.LWWDelta{
					DocID:               d.docID,
					FieldName:           field,
					Priority:            d.fieldsHeight[field],
					CollectionVersionID: d.col.Version().VersionID,
					Data:                encodeValue(val),
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
		crdt.NewCRDT(&crdt.DocCompositeDelta{
			DocID:               d.docID,
			Priority:            newPriority,
			CollectionVersionID: d.col.Version().VersionID,
			Status:              client.Active,
		}),
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

// generateCompositeDelete creates a composite block that marks the document as deleted.
func (d *dagBuilder) generateCompositeDelete(lsys *linking.LinkSystem, from compositeInfo) (compositeInfo, error) {
	heads := []cid.Cid{}
	newPriority := from.height + 1
	if from.link.ByteLen() != 0 {
		heads = append(heads, from.link.Cid)
	}

	compositeBlock := coreblock.New(
		crdt.NewCRDT(&crdt.DocCompositeDelta{
			DocID:               d.docID,
			Priority:            newPriority,
			CollectionVersionID: d.col.Version().VersionID,
			Status:              client.Deleted,
		}),
		nil,
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

// generateCounterCompositeUpdate creates a composite block with counter field deltas
// instead of LWW deltas.
func (d *dagBuilder) generateCounterCompositeUpdate(
	lsys *linking.LinkSystem,
	fields map[string]any,
	from compositeInfo,
) (compositeInfo, error) {
	heads := []cid.Cid{}
	newPriority := from.height + 1
	if from.link.ByteLen() != 0 {
		heads = append(heads, from.link.Cid)
	}

	links := []coreblock.DAGLink{}
	for field, val := range fields {
		d.fieldsHeight[field]++
		fieldBlock := coreblock.Block{
			Delta: crdt.CRDT{
				CounterDelta: &crdt.CounterDelta{
					DocID:               d.docID,
					FieldName:           field,
					Priority:            d.fieldsHeight[field],
					CollectionVersionID: d.col.Version().VersionID,
					Data:                encodeValue(val),
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
		crdt.NewCRDT(&crdt.DocCompositeDelta{
			DocID:               d.docID,
			Priority:            newPriority,
			CollectionVersionID: d.col.Version().VersionID,
			Status:              client.Active,
		}),
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
	require.Len(t, q.keys, 1)
	q.done(testDocID)
	// give time for the goroutine to add the docID
	time.Sleep(10 * time.Millisecond)
	q.mutex.Lock()
	require.Len(t, q.keys, 1)
	q.mutex.Unlock()
	q.done(testDocID)
	q.mutex.Lock()
	require.Len(t, q.keys, 0)
	q.mutex.Unlock()
}

// TestMerge_ThreeWayFork_NoError tests merging three concurrent branches
// that each update a different field from the same parent.
//
// DAG structure:
//
//	   A (create: name="John")
//	 / | \
//	B  C  D  (B: name="Johny", C: age=30, D: name="Jane")
func TestMerge_ThreeWayFork_NoError(t *testing.T) {
	ctx := context.Background()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(blockstore.NewIPLDStore(datastore.BlockstoreFrom(db.rootstore, immutable.None[int]())))

	initialDocState := map[string]any{
		"name": "John",
	}
	builder, docID := newDagBuilder(ctx, col, initialDocState)
	compInfo, err := builder.generateCompositeUpdate(&lsys, initialDocState, compositeInfo{})
	require.NoError(t, err)

	// Branch B: update name
	compInfoB, err := builder.generateCompositeUpdate(&lsys, map[string]any{"name": "Johny"}, compInfo)
	require.NoError(t, err)

	err = db.executeMerge(ctx, col.(*collection), event.Merge{
		DocID:        docID.String(),
		Cid:          compInfoB.link.Cid,
		CollectionID: col.CollectionID(),
	})
	require.NoError(t, err)

	// Branch C: update age (from same parent A)
	compInfoC, err := builder.generateCompositeUpdate(&lsys, map[string]any{"age": 30}, compInfo)
	require.NoError(t, err)

	err = db.executeMerge(ctx, col.(*collection), event.Merge{
		DocID:        docID.String(),
		Cid:          compInfoC.link.Cid,
		CollectionID: col.CollectionID(),
	})
	require.NoError(t, err)

	// Reset name height so D has same field priority as B.
	builder.fieldsHeight["name"] = 1

	// Branch D: update name again (from same parent A)
	compInfoD, err := builder.generateCompositeUpdate(&lsys, map[string]any{"name": "Jane"}, compInfo)
	require.NoError(t, err)

	err = db.executeMerge(ctx, col.(*collection), event.Merge{
		DocID:        docID.String(),
		Cid:          compInfoD.link.Cid,
		CollectionID: col.CollectionID(),
	})
	require.NoError(t, err)

	doc, err := col.GetDocument(ctx, docID)
	require.NoError(t, err)
	docMap, err := doc.ToMap()
	require.NoError(t, err)

	// "Johny" vs "Jane" at same priority: "Johny" > "Jane" lexicographically (CBOR).
	expectedDocMap := map[string]any{
		"_docID": docID.String(),
		"age":    int64(30),
		"name":   "Johny",
	}

	require.Equal(t, expectedDocMap, docMap)
}

// TestMerge_DiamondMerge_NoError tests a diamond DAG shape where two branches
// fork from the same parent and then reconverge into a single multi-parent block.
//
// DAG structure:
//
//	  A (create: name="John")
//	 / \
//	B   C  (B: name="Johny", C: age=30)
//	 \ /
//	  D  (reconverge — D has heads=[B,C])
func TestMerge_DiamondMerge_NoError(t *testing.T) {
	ctx := context.Background()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(blockstore.NewIPLDStore(datastore.BlockstoreFrom(db.rootstore, immutable.None[int]())))

	initialDocState := map[string]any{
		"name": "John",
	}
	d, docID := newDagBuilder(ctx, col, initialDocState)
	compInfo, err := d.generateCompositeUpdate(&lsys, initialDocState, compositeInfo{})
	require.NoError(t, err)

	// Branch B: update name
	compInfoB, err := d.generateCompositeUpdate(&lsys, map[string]any{"name": "Johny"}, compInfo)
	require.NoError(t, err)

	// Branch C: update age (from same parent A)
	compInfoC, err := d.generateCompositeUpdate(&lsys, map[string]any{"age": 30}, compInfo)
	require.NoError(t, err)

	// D: reconverge from B and C
	compInfoD, err := d.generateCompositeUpdateFromHeads(&lsys, map[string]any{"name": "Final"}, []compositeInfo{compInfoB, compInfoC})
	require.NoError(t, err)

	err = db.executeMerge(ctx, col.(*collection), event.Merge{
		DocID:        docID.String(),
		Cid:          compInfoD.link.Cid,
		CollectionID: col.CollectionID(),
	})
	require.NoError(t, err)

	doc, err := col.GetDocument(ctx, docID)
	require.NoError(t, err)
	docMap, err := doc.ToMap()
	require.NoError(t, err)

	expectedDocMap := map[string]any{
		"_docID": docID.String(),
		"age":    int64(30),
		"name":   "Final",
	}

	require.Equal(t, expectedDocMap, docMap)
}

// TestMerge_AsymmetricBranches_NoError tests an asymmetric DAG where one
// branch is much deeper than the other. This exercises the loadComposites
// backward walk when the incoming block's priority is less than the merge
// target's head height.
//
// DAG structure:
//
//	  A (create: name="John")
//	 / \
//	B   E  (E: age=40, at height=2)
//	|
//	C
//	|
//	D  (D: name at height=4)
func TestMerge_AsymmetricBranches_NoError(t *testing.T) {
	ctx := context.Background()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(blockstore.NewIPLDStore(datastore.BlockstoreFrom(db.rootstore, immutable.None[int]())))

	initialDocState := map[string]any{
		"name": "John",
	}
	d, docID := newDagBuilder(ctx, col, initialDocState)
	compInfo, err := d.generateCompositeUpdate(&lsys, initialDocState, compositeInfo{})
	require.NoError(t, err)

	// Deep branch: A → B → C → D
	compInfoB, err := d.generateCompositeUpdate(&lsys, map[string]any{"name": "B"}, compInfo)
	require.NoError(t, err)
	compInfoC, err := d.generateCompositeUpdate(&lsys, map[string]any{"name": "C"}, compInfoB)
	require.NoError(t, err)
	compInfoD, err := d.generateCompositeUpdate(&lsys, map[string]any{"name": "D"}, compInfoC)
	require.NoError(t, err)

	// Merge the deep branch first — heads now at D (height=4)
	err = db.executeMerge(ctx, col.(*collection), event.Merge{
		DocID:        docID.String(),
		Cid:          compInfoD.link.Cid,
		CollectionID: col.CollectionID(),
	})
	require.NoError(t, err)

	// Shallow branch: A → E (height=2, less than current head height=4)
	compInfoE, err := d.generateCompositeUpdate(&lsys, map[string]any{"age": 40}, compInfo)
	require.NoError(t, err)

	// This merge must walk the target DAG backward to find the common ancestor.
	err = db.executeMerge(ctx, col.(*collection), event.Merge{
		DocID:        docID.String(),
		Cid:          compInfoE.link.Cid,
		CollectionID: col.CollectionID(),
	})
	require.NoError(t, err)

	doc, err := col.GetDocument(ctx, docID)
	require.NoError(t, err)
	docMap, err := doc.ToMap()
	require.NoError(t, err)

	expectedDocMap := map[string]any{
		"_docID": docID.String(),
		"age":    int64(40),
		"name":   "D",
	}

	require.Equal(t, expectedDocMap, docMap)
}

// TestMerge_DeleteVsUpdate_DeleteWins tests that when one branch updates a
// field and the other branch deletes the document, the delete wins.
//
// DAG structure:
//
//	  A (create: name="John")
//	 / \
//	B   C  (B: name="Jane" [update], C: delete)
//
// Merge order: B first, then C.
func TestMerge_DeleteVsUpdate_DeleteWins(t *testing.T) {
	ctx := context.Background()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(blockstore.NewIPLDStore(datastore.BlockstoreFrom(db.rootstore, immutable.None[int]())))

	initialDocState := map[string]any{
		"name": "John",
	}
	d, docID := newDagBuilder(ctx, col, initialDocState)
	compInfo, err := d.generateCompositeUpdate(&lsys, initialDocState, compositeInfo{})
	require.NoError(t, err)

	// Branch B: update name
	compInfoB, err := d.generateCompositeUpdate(&lsys, map[string]any{"name": "Jane"}, compInfo)
	require.NoError(t, err)

	err = db.executeMerge(ctx, col.(*collection), event.Merge{
		DocID:        docID.String(),
		Cid:          compInfoB.link.Cid,
		CollectionID: col.CollectionID(),
	})
	require.NoError(t, err)

	// Branch C: delete
	compInfoC, err := d.generateCompositeDelete(&lsys, compInfo)
	require.NoError(t, err)

	err = db.executeMerge(ctx, col.(*collection), event.Merge{
		DocID:        docID.String(),
		Cid:          compInfoC.link.Cid,
		CollectionID: col.CollectionID(),
	})
	require.NoError(t, err)

	// Document should not be accessible via normal Get (deleted).
	_, err = col.GetDocument(ctx, docID)
	require.ErrorIs(t, err, client.ErrDocumentNotFoundOrNotAuthorized)
}

// TestMerge_UpdateVsDelete_DeleteStillWins tests the same scenario as
// TestMerge_DeleteVsUpdate_DeleteWins but with merge order reversed.
// The delete should still win because P2P updates don't undelete.
//
// DAG structure:
//
//	  A (create: name="John")
//	 / \
//	B   C  (B: delete, C: name="Jane" [update])
//
// Merge order: B (delete) first, then C (update).
func TestMerge_UpdateVsDelete_DeleteStillWins(t *testing.T) {
	ctx := context.Background()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(blockstore.NewIPLDStore(datastore.BlockstoreFrom(db.rootstore, immutable.None[int]())))

	initialDocState := map[string]any{
		"name": "John",
	}
	d, docID := newDagBuilder(ctx, col, initialDocState)
	compInfo, err := d.generateCompositeUpdate(&lsys, initialDocState, compositeInfo{})
	require.NoError(t, err)

	// Branch B: delete first
	compInfoB, err := d.generateCompositeDelete(&lsys, compInfo)
	require.NoError(t, err)

	err = db.executeMerge(ctx, col.(*collection), event.Merge{
		DocID:        docID.String(),
		Cid:          compInfoB.link.Cid,
		CollectionID: col.CollectionID(),
	})
	require.NoError(t, err)

	// Branch C: update (from same parent A)
	compInfoC, err := d.generateCompositeUpdate(&lsys, map[string]any{"name": "Jane"}, compInfo)
	require.NoError(t, err)

	err = db.executeMerge(ctx, col.(*collection), event.Merge{
		DocID:        docID.String(),
		Cid:          compInfoC.link.Cid,
		CollectionID: col.CollectionID(),
	})
	require.NoError(t, err)

	// Document should still be deleted — P2P updates don't undelete.
	_, err = col.GetDocument(ctx, docID)
	require.ErrorIs(t, err, client.ErrDocumentNotFoundOrNotAuthorized)
}

// TestMerge_CounterThreeWayFork_Accumulates tests that three concurrent
// counter increments from the same parent all accumulate correctly.
//
// DAG structure:
//
//	   A (create: points=0)
//	 / | \
//	B  C  D  (B: +10, C: +20, D: +30)
//
// Final value should be 0+10+20+30 = 60.
func TestMerge_CounterThreeWayFork_Accumulates(t *testing.T) {
	ctx := context.Background()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)

	_, err = db.AddCollection(ctx, userSchemaWithCounter)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(blockstore.NewIPLDStore(datastore.BlockstoreFrom(db.rootstore, immutable.None[int]())))

	initialDocState := map[string]any{
		"name":   "John",
		"points": 0,
	}
	d, docID := newDagBuilder(ctx, col, initialDocState)

	// Initial block: use LWW for name, counter for points
	// We need to create the initial block with mixed field types.
	// The initial create uses generateCompositeUpdate for name (LWW)
	// and counter for points.
	compInfo, err := d.generateCompositeUpdate(&lsys, map[string]any{"name": "John"}, compositeInfo{})
	require.NoError(t, err)

	// Also create initial counter block at same parent
	d.fieldsHeight["points"] = 0
	compInfoInit, err := d.generateCounterCompositeUpdate(&lsys, map[string]any{"points": int64(0)}, compositeInfo{})
	require.NoError(t, err)

	// Merge both initial blocks
	err = db.executeMerge(ctx, col.(*collection), event.Merge{
		DocID:        docID.String(),
		Cid:          compInfo.link.Cid,
		CollectionID: col.CollectionID(),
	})
	require.NoError(t, err)

	err = db.executeMerge(ctx, col.(*collection), event.Merge{
		DocID:        docID.String(),
		Cid:          compInfoInit.link.Cid,
		CollectionID: col.CollectionID(),
	})
	require.NoError(t, err)

	// Branch B: +10
	compInfoB, err := d.generateCounterCompositeUpdate(&lsys, map[string]any{"points": int64(10)}, compInfoInit)
	require.NoError(t, err)

	err = db.executeMerge(ctx, col.(*collection), event.Merge{
		DocID:        docID.String(),
		Cid:          compInfoB.link.Cid,
		CollectionID: col.CollectionID(),
	})
	require.NoError(t, err)

	// Reset counter field height so C gets same field priority as B.
	d.fieldsHeight["points"] = 1

	// Branch C: +20 (from same parent)
	compInfoC, err := d.generateCounterCompositeUpdate(&lsys, map[string]any{"points": int64(20)}, compInfoInit)
	require.NoError(t, err)

	err = db.executeMerge(ctx, col.(*collection), event.Merge{
		DocID:        docID.String(),
		Cid:          compInfoC.link.Cid,
		CollectionID: col.CollectionID(),
	})
	require.NoError(t, err)

	// Reset counter field height so D gets same field priority as B and C.
	d.fieldsHeight["points"] = 1

	// Branch D: +30 (from same parent)
	compInfoD, err := d.generateCounterCompositeUpdate(&lsys, map[string]any{"points": int64(30)}, compInfoInit)
	require.NoError(t, err)

	err = db.executeMerge(ctx, col.(*collection), event.Merge{
		DocID:        docID.String(),
		Cid:          compInfoD.link.Cid,
		CollectionID: col.CollectionID(),
	})
	require.NoError(t, err)

	doc, err := col.GetDocument(ctx, docID)
	require.NoError(t, err)
	docMap, err := doc.ToMap()
	require.NoError(t, err)

	// 0 + 10 + 20 + 30 = 60
	expectedDocMap := map[string]any{
		"_docID": docID.String(),
		"name":   "John",
		"points": int64(60),
	}

	require.Equal(t, expectedDocMap, docMap)
}
