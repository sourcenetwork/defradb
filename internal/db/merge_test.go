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
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	badgerds "github.com/dgraph-io/badger/v4"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/badger"
	"github.com/sourcenetwork/corekv/blockstore"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/datastore"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
	"github.com/sourcenetwork/defradb/internal/db/id"
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
	d, _ := newDagBuilder(ctx, col, initialDocState)
	compInfo, err := d.generateCompositeUpdate(&lsys, initialDocState, compositeInfo{})
	require.NoError(t, err)
	docID := client.NewDocIDV0(compInfo.link.Cid)
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

func TestMerge_ConcurrentNewDocuments(t *testing.T) {
	ctx := context.Background()
	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(blockstore.NewIPLDStore(datastore.BlockstoreFrom(db.rootstore, immutable.None[int]())))

	const docCount = 24
	events := make([]event.Merge, docCount)
	docIDs := make([]client.DocID, docCount)
	for i := range docCount {
		state := map[string]any{"name": fmt.Sprintf("user-%d", i), "age": i}
		builder, _ := newDagBuilder(ctx, col, state)
		composite, err := builder.generateCompositeUpdate(&lsys, state, compositeInfo{})
		require.NoError(t, err)
		docIDs[i] = client.NewDocIDV0(composite.link.Cid)
		events[i] = event.Merge{
			DocID:        docIDs[i].String(),
			Cid:          composite.link.Cid,
			CollectionID: col.CollectionID(),
		}
	}

	start := make(chan struct{})
	errs := make(chan error, docCount)
	var group sync.WaitGroup
	for _, mergeEvent := range events {
		group.Go(func() {
			<-start
			errs <- db.Merge(ctx, mergeEvent)
		})
	}
	close(start)
	group.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	for _, docID := range docIDs {
		_, err := col.GetDocument(ctx, docID)
		require.NoError(t, err)
	}
}

func TestMerge_ZeroMaxRetriesStillAttempts(t *testing.T) {
	ctx := context.Background()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	db.maxTxnRetries = immutable.Some(0)

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	// The block cannot be loaded, so an attempt fails there. Asserting that specific
	// failure is what distinguishes one attempt from none: skipping the loop entirely
	// reports exhausted retries and never touches the block.
	err = db.Merge(ctx, event.Merge{
		DocID:        "missing",
		Cid:          blocks.NewBlock(nil).Cid(),
		CollectionID: col.CollectionID(),
	})
	require.ErrorContains(t, err, "failed to load block for merge")
}

// The batch path has its own retry loop, so it needs the same floor as the single-document
// one or a zero budget drops every event in the batch without trying.
func TestMergeBatch_ZeroMaxRetriesStillAttempts(t *testing.T) {
	ctx := context.Background()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	db.maxTxnRetries = immutable.Some(0)

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	merged, err := db.MergeBatchWithTxn(ctx, []event.Merge{{
		DocID:        "missing",
		Cid:          blocks.NewBlock(nil).Cid(),
		CollectionID: col.CollectionID(),
	}})
	require.ErrorContains(t, err, "failed to load block for merge")
	require.Equal(t, []bool{false}, merged)
}

func TestMerge_GenesisWithEmptyDocID_ResolvesDocIDAndFieldMappings(t *testing.T) {
	ctx := context.Background()

	sourceDB, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer sourceDB.Close()
	targetDB, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer targetDB.Close()

	_, err = sourceDB.AddCollection(ctx, userSchema)
	require.NoError(t, err)
	_, err = targetDB.AddCollection(ctx, userSchema)
	require.NoError(t, err)

	sourceCol, err := sourceDB.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	targetCol, err := targetDB.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	setDocIDSequence(t, ctx, sourceDB, 100)
	setDocIDSequence(t, ctx, targetDB, 200)

	sourceDoc, err := client.NewDocFromJSON(ctx, []byte(`{"name":"John","age":30}`), sourceCol.Version())
	require.NoError(t, err)
	err = sourceCol.AddDocument(ctx, sourceDoc)
	require.NoError(t, err)

	copyDAGBlocks(t, ctx, sourceDB, targetDB, sourceDoc.Head())

	err = targetDB.executeMerge(ctx, targetCol.(*collection), event.Merge{
		DocID:        sourceDoc.ID().String(),
		Cid:          sourceDoc.Head(),
		CollectionID: targetCol.CollectionID(),
	})
	require.NoError(t, err)

	mergedDoc, err := targetCol.GetDocument(ctx, sourceDoc.ID())
	require.NoError(t, err)
	docMap, err := mergedDoc.ToMap()
	require.NoError(t, err)
	require.Equal(t, map[string]any{
		"_docID": sourceDoc.ID().String(),
		"name":   "John",
		"age":    int64(30),
	}, docMap)

	compositeBlock := loadTestBlock(t, ctx, sourceDB, sourceDoc.Head())
	require.NotEmpty(t, compositeBlock.Links)
	fieldCID := compositeBlock.Links[0].Cid

	txn, err := targetDB.NewTxn(true)
	require.NoError(t, err)
	defer txn.Discard()
	dbTxn, ok := txn.(*Txn)
	require.True(t, ok)
	txnCtx := InitContext(ctx, dbTxn)
	collectionShortID, err := id.GetCollectionShortID(txnCtx, targetCol.CollectionID())
	require.NoError(t, err)

	docShortID, found, err := id.GetDocShortID(txnCtx, collectionShortID, sourceDoc.ID().String())
	require.NoError(t, err)
	require.True(t, found)
	require.NotEqual(t, sourceDoc.ID().String(), docShortID)

	blockDocIDs, err := id.GetDocIDsForBlockFromStore(txnCtx, dbTxn.Systemstore(), fieldCID)
	require.NoError(t, err)
	require.Equal(t, []string{sourceDoc.ID().String()}, blockDocIDs)
}

// conflictingStore fails every write transaction with a conflict once armed. Arming it is the
// only way to drive a merge through its whole retry budget on demand: a real conflict needs a
// second writer committing inside the window between this transaction's first read and its commit.
type conflictingStore struct {
	corekv.TxnStore

	// failCommits is decremented unconditionally, so reading and spending the budget is one
	// step. A spent budget goes negative, which reads the same as zero.
	failCommits atomic.Int64
}

func (s *conflictingStore) NewTxn(readonly bool) corekv.Txn {
	txn := s.TxnStore.NewTxn(readonly)
	if readonly {
		return txn
	}
	return conflictingTxn{Txn: txn, store: s}
}

type conflictingTxn struct {
	corekv.Txn
	store *conflictingStore
}

func (t conflictingTxn) Commit() error {
	if t.store.failCommits.Add(-1) >= 0 {
		return corekv.ErrTxnConflict
	}
	return t.Txn.Commit()
}

// newConflictingBadgerDB returns a database whose write commits can be made to conflict, and the
// store holding the budget that does it. Setup runs with an empty budget so collections can still
// be created.
func newConflictingBadgerDB(ctx context.Context) (*DB, *conflictingStore, error) {
	rootstore, err := badger.NewDatastore("", badgerds.DefaultOptions("").WithInMemory(true))
	if err != nil {
		return nil, nil, err
	}
	adminInfo, err := acpDB.NewNACInfo(ctx, "", false)
	if err != nil {
		return nil, nil, err
	}
	store := &conflictingStore{TxnStore: rootstore}
	db, err := newDB(ctx, store, adminInfo)
	if err != nil {
		return nil, nil, err
	}
	return db, store, nil
}

// stageUnmergedDoc writes a document's blocks to the store without merging them, and returns the
// event that would merge it.
func stageUnmergedDoc(t *testing.T, ctx context.Context, db *DB, col client.Collection) (client.DocID, event.Merge) {
	t.Helper()

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(blockstore.NewIPLDStore(datastore.BlockstoreFrom(db.rootstore, immutable.None[int]())))

	state := map[string]any{"name": "John"}
	builder, _ := newDagBuilder(ctx, col, state)
	composite, err := builder.generateCompositeUpdate(&lsys, state, compositeInfo{})
	require.NoError(t, err)

	docID := client.NewDocIDV0(composite.link.Cid)
	return docID, event.Merge{
		DocID:        docID.String(),
		Cid:          composite.link.Cid,
		CollectionID: col.CollectionID(),
	}
}

// Every attempt conflicting means nothing was written. Returning nil here would tell the caller
// the document merged, and the caller relays and acknowledges on that basis.
func TestMerge_ExhaustedRetries_ReportsFailure(t *testing.T) {
	ctx := context.Background()

	db, conflict, err := newConflictingBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	docID, mergeEvent := stageUnmergedDoc(t, ctx, db, col)

	conflict.failCommits.Store(math.MaxInt64)
	err = db.Merge(ctx, mergeEvent)
	require.Error(t, err, "a merge that committed nothing must not report success")
	require.ErrorContains(t, err, "maximum transaction")

	_, err = col.GetDocument(ctx, docID)
	require.Error(t, err, "the error must be truthful: nothing was stored")
}

// The batch path has its own retry loop, so it needs its own case: the parallel slice is what
// callers read to decide which events to relay.
func TestMergeBatch_ExhaustedRetries_ReportsNothingMerged(t *testing.T) {
	ctx := context.Background()

	db, conflict, err := newConflictingBadgerDB(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	docID, mergeEvent := stageUnmergedDoc(t, ctx, db, col)

	conflict.failCommits.Store(math.MaxInt64)
	merged, err := db.MergeBatchWithTxn(ctx, []event.Merge{mergeEvent})
	require.Error(t, err)
	require.Equal(t, []bool{false}, merged, "an event that committed nothing must not be reported as merged")

	_, err = col.GetDocument(ctx, docID)
	require.Error(t, err, "the report must be truthful: nothing was stored")
}

func TestMergeResolveBlockDocID(t *testing.T) {
	ctx := context.Background()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	c, ok := col.(*collection)
	require.True(t, ok)

	txn, err := db.NewTxn(false)
	require.NoError(t, err)
	defer txn.Discard()
	dbTxn, ok := txn.(*Txn)
	require.True(t, ok)
	txnCtx := InitContext(ctx, dbTxn)

	collectionShortID, err := id.GetCollectionShortID(txnCtx, col.CollectionID())
	require.NoError(t, err)

	mp, err := db.newMergeProcessor(txnCtx, c, true)
	require.NoError(t, err)

	genesisCID := blocks.NewBlock([]byte("genesis composite")).Cid()
	genesisDocID := client.NewDocIDV0(genesisCID).String()
	genesisBlock := &coreblock.Block{
		Delta: crdt.NewCRDT(&crdt.DocCompositeDelta{
			CollectionVersionID: col.Version().VersionID,
			Status:              client.Active,
		}),
	}
	resolved, err := mp.resolveCompositeBlockDocRef(txnCtx, collectionShortID, genesisBlock, genesisCID)
	require.NoError(t, err)
	require.Equal(t, genesisDocID, resolved.docID)
	require.NotEmpty(t, resolved.docShortID)
	require.NotZero(t, resolved.docShortID)
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
	d, _ := newDagBuilder(ctx, col, initialDocState)
	compInfo, err := d.generateCompositeUpdate(&lsys, initialDocState, compositeInfo{})
	require.NoError(t, err)
	docID := client.NewDocIDV0(compInfo.link.Cid)
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

func copyDAGBlocks(t *testing.T, ctx context.Context, sourceDB *DB, targetDB *DB, root cid.Cid) {
	t.Helper()

	sourceStore := datastore.BlockstoreFrom(sourceDB.rootstore, sourceDB.blockStoreChunkSize)
	targetStore := datastore.BlockstoreFrom(targetDB.rootstore, targetDB.blockStoreChunkSize)
	seen := make(map[cid.Cid]struct{})

	var copyBlock func(cid.Cid)
	copyBlock = func(blockCID cid.Cid) {
		if _, ok := seen[blockCID]; ok {
			return
		}
		seen[blockCID] = struct{}{}

		rawBlock, err := sourceStore.Get(ctx, blockCID)
		require.NoError(t, err)
		err = targetStore.Put(ctx, rawBlock)
		require.NoError(t, err)

		block, err := coreblock.GetFromBytes(rawBlock.RawData())
		require.NoError(t, err)
		for _, link := range block.AllLinks() {
			copyBlock(link.Cid)
		}
	}

	copyBlock(root)
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
	d, _ := newDagBuilder(ctx, col, initialDocState)
	compInfo, err := d.generateCompositeUpdate(&lsys, initialDocState, compositeInfo{})
	require.NoError(t, err)
	docID := client.NewDocIDV0(compInfo.link.Cid)
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
	require.ErrorContains(t, err, "could not find "+someUnknownLink.Cid.String())

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

func TestMergeBatch_OneEventCannotMerge_OthersStillLand(t *testing.T) {
	ctx := context.Background()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(blockstore.NewIPLDStore(datastore.BlockstoreFrom(db.rootstore, immutable.None[int]())))

	goodState := map[string]any{"name": "John"}
	goodBuilder, _ := newDagBuilder(ctx, col, goodState)
	goodInfo, err := goodBuilder.generateCompositeUpdate(&lsys, goodState, compositeInfo{})
	require.NoError(t, err)
	goodDocID := client.NewDocIDV0(goodInfo.link.Cid)

	// A second document whose parent composite was never stored, so it cannot merge.
	badState := map[string]any{"name": "Jane"}
	badBuilder, _ := newDagBuilder(ctx, col, badState)
	badGenesis, err := badBuilder.generateCompositeUpdate(&lsys, badState, compositeInfo{})
	require.NoError(t, err)
	badDocID := client.NewDocIDV0(badGenesis.link.Cid)

	missingBlock := coreblock.Block{Delta: crdt.CRDT{DocCompositeDelta: &crdt.DocCompositeDelta{Status: 1}}}
	missingLink, err := coreblock.GetLinkFromNode(missingBlock.GenerateNode())
	require.NoError(t, err)

	badInfo, err := badBuilder.generateCompositeUpdate(
		&lsys,
		map[string]any{"name": "Janet"},
		compositeInfo{link: missingLink, height: 2},
	)
	require.NoError(t, err)

	// The unmergeable event is listed first so that an all-or-nothing batch would
	// abort before ever reaching the one that can merge.
	merged, err := db.MergeBatchWithTxn(ctx, []event.Merge{
		{DocID: badDocID.String(), Cid: badInfo.link.Cid, CollectionID: col.CollectionID()},
		{DocID: goodDocID.String(), Cid: goodInfo.link.Cid, CollectionID: col.CollectionID()},
	})
	require.ErrorContains(t, err, "could not find "+missingLink.Cid.String())
	require.ErrorContains(t, err, badDocID.String())
	require.Equal(t, []bool{false, true}, merged)

	doc, err := col.GetDocument(ctx, goodDocID)
	require.NoError(t, err)
	docMap, err := doc.ToMap()
	require.NoError(t, err)
	require.Equal(t, map[string]any{"_docID": goodDocID.String(), "name": "John"}, docMap)
}

// The result slice is indexed by the caller's event order, which chunking has to
// preserve on both the committed and the isolated path. Only a chunk at a non-zero
// offset distinguishes that from an index taken relative to the chunk, so the batch
// spans three: one clean, one holding the bad event, and one clean after it.
func TestMergeBatch_FailureInLaterChunk_ReportsFailureAgainstTheRightEvent(t *testing.T) {
	ctx := context.Background()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)

	_, err = db.AddCollection(ctx, userSchema)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(blockstore.NewIPLDStore(datastore.BlockstoreFrom(db.rootstore, immutable.None[int]())))

	// Sized off the chunk size so the bad event lands in the second chunk and a third
	// chunk still commits whole.
	count := 2*mergeChunkSize + 2
	badIndex := mergeChunkSize + 1

	names := make([]string, count)
	events := make([]event.Merge, count)
	docIDs := make([]client.DocID, count)
	var missingLink cidlink.Link

	for i := range names {
		name := fmt.Sprintf("user%d", i)
		names[i] = name
		state := map[string]any{"name": name}
		builder, _ := newDagBuilder(ctx, col, state)
		genesis, err := builder.generateCompositeUpdate(&lsys, state, compositeInfo{})
		require.NoError(t, err)
		docIDs[i] = client.NewDocIDV0(genesis.link.Cid)

		info := genesis
		if i == badIndex {
			// Parent it on a composite that was never stored, so it cannot merge.
			missingBlock := coreblock.Block{Delta: crdt.CRDT{DocCompositeDelta: &crdt.DocCompositeDelta{Status: 1}}}
			missingLink, err = coreblock.GetLinkFromNode(missingBlock.GenerateNode())
			require.NoError(t, err)

			info, err = builder.generateCompositeUpdate(
				&lsys,
				map[string]any{"name": name + "ita"},
				compositeInfo{link: missingLink, height: 2},
			)
			require.NoError(t, err)
		}
		events[i] = event.Merge{
			DocID:        docIDs[i].String(),
			Cid:          info.link.Cid,
			CollectionID: col.CollectionID(),
		}
	}

	merged, err := db.MergeBatchWithTxn(ctx, events)
	require.ErrorContains(t, err, "could not find "+missingLink.Cid.String())

	expected := make([]bool, count)
	for i := range expected {
		expected[i] = i != badIndex
	}
	require.Equal(t, expected, merged)

	// The result slice has to agree with what is actually readable from the store.
	for i, docID := range docIDs {
		doc, err := col.GetDocument(ctx, docID)
		if i == badIndex {
			require.Error(t, err)
			continue
		}
		require.NoError(t, err)
		docMap, err := doc.ToMap()
		require.NoError(t, err)
		require.Equal(t, map[string]any{"_docID": docID.String(), "name": names[i]}, docMap)
	}
}

type dagBuilder struct {
	fieldsHeight map[string]uint64
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
	builder, _ := newDagBuilder(ctx, col, initialDocState)
	compInfo, err := builder.generateCompositeUpdate(&lsys, initialDocState, compositeInfo{})
	require.NoError(t, err)
	docID := client.NewDocIDV0(compInfo.link.Cid)

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
	d, _ := newDagBuilder(ctx, col, initialDocState)
	compInfo, err := d.generateCompositeUpdate(&lsys, initialDocState, compositeInfo{})
	require.NoError(t, err)
	docID := client.NewDocIDV0(compInfo.link.Cid)

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
	d, _ := newDagBuilder(ctx, col, initialDocState)
	compInfo, err := d.generateCompositeUpdate(&lsys, initialDocState, compositeInfo{})
	require.NoError(t, err)
	docID := client.NewDocIDV0(compInfo.link.Cid)

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
	d, _ := newDagBuilder(ctx, col, initialDocState)
	compInfo, err := d.generateCompositeUpdate(&lsys, initialDocState, compositeInfo{})
	require.NoError(t, err)
	docID := client.NewDocIDV0(compInfo.link.Cid)

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
	d, _ := newDagBuilder(ctx, col, initialDocState)
	compInfo, err := d.generateCompositeUpdate(&lsys, initialDocState, compositeInfo{})
	require.NoError(t, err)
	docID := client.NewDocIDV0(compInfo.link.Cid)

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
	d, _ := newDagBuilder(ctx, col, initialDocState)

	// Initial block: use LWW for name, counter for points
	// We need to create the initial block with mixed field types.
	// The initial create uses generateCompositeUpdate for name (LWW)
	// and counter for points.
	compInfo, err := d.generateCompositeUpdate(&lsys, map[string]any{"name": "John"}, compositeInfo{})
	require.NoError(t, err)
	docID := client.NewDocIDV0(compInfo.link.Cid)

	// Add the initial counter field as a follow-up composite on the same DAG.
	d.fieldsHeight["points"] = 0
	compInfoInit, err := d.generateCounterCompositeUpdate(&lsys, map[string]any{"points": int64(0)}, compInfo)
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
