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
	"container/list"
	"context"
	"fmt"
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/blockstore"

	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/utils"
)

func (db *DB) Merge(ctx context.Context, evt event.Merge) error {
	col, err := getCollectionFromCollectionID(ctx, db, evt.CollectionID)
	if err != nil {
		return err
	}

	if col.Version().IsBranchable {
		// As collection commits link to document composite commits, all events
		// recieved for branchable collections must be processed serially else
		// they may otherwise cause a transaction conflict.
		db.colMergeQueue.add(evt.CollectionID)
		defer db.colMergeQueue.done(evt.CollectionID)
	} else {
		// ensure only one merge per docID
		db.docMergeQueue.add(evt.DocID)
		defer db.docMergeQueue.done(evt.DocID)
	}

	// Conflicts occur when a user updates a document while a merge is in progress.
	max := db.MaxTxnRetries()
	if max < 1 {
		max = 1
	}
	for i := 0; i < max; i++ {
		err = db.executeMerge(ctx, col, evt)
		if errors.Is(err, corekv.ErrTxnConflict) {
			continue
		}
		if err != nil {
			return err
		}
		return nil
	}
	return client.NewErrMaxTxnRetries(err)
}

func (db *DB) executeMerge(ctx context.Context, col *collection, dagMerge event.Merge) error {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return NewErrCreateMergeTxn(err, dagMerge.DocID, dagMerge.Cid.String())
	}

	defer txn.Discard()

	key, exists, err := getDocHeadstoreKey(ctx, col, dagMerge.DocID)
	if err != nil {
		return err
	}

	mt := newMergeTarget()
	if exists {
		mt, err = getHeadsAsMergeTarget(ctx, key)
		if err != nil {
			return NewErrGetMergeTargetHeads(err, dagMerge.DocID, string(key.Bytes()))
		}
	}

	mp, err := db.newMergeProcessor(ctx, col, len(mt.heads) == 0)
	if err != nil {
		return err
	}

	err = mp.loadComposites(ctx, dagMerge.Cid, mt)
	if err != nil {
		return NewErrLoadComposites(err, dagMerge.Cid.String(), dagMerge.DocID)
	}

	err = mp.mergeComposites(ctx)
	if err != nil {
		return NewErrMergeComposites(err, dagMerge.DocID)
	}

	for docID, oldDoc := range mp.docIDs {
		err = syncIndexedDoc(ctx, docID, mp.col, oldDoc)
		if err != nil {
			return NewErrSyncIndexedDoc(err, docID.String())
		}
	}

	err = txn.Commit()
	if err != nil {
		return err
	}

	// send a complete event so we can track merges in the integration tests
	db.events.Publish(event.NewMessage(event.MergeCompleteName, event.MergeComplete{
		Merge: dagMerge,
	}))
	return nil
}

// mergeQueue is synchronization source to ensure that concurrent
// document merges do not cause transaction conflicts.
type mergeQueue struct {
	keys  map[string]chan struct{}
	mutex sync.Mutex
}

func newMergeQueue() *mergeQueue {
	return &mergeQueue{
		keys: make(map[string]chan struct{}),
	}
}

// add adds a key to the queue. If the key is already in the queue, it will
// wait for the key to be removed from the queue. For every add call, done must
// be called to remove the key from the queue. Otherwise, subsequent add calls will
// block forever.
func (m *mergeQueue) add(key string) {
	m.mutex.Lock()
	done, ok := m.keys[key]
	if !ok {
		m.keys[key] = make(chan struct{})
	}
	m.mutex.Unlock()
	if ok {
		<-done
		m.add(key)
	}
}

func (m *mergeQueue) done(key string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	done, ok := m.keys[key]
	if ok {
		delete(m.keys, key)
		close(done)
	}
}

type mergeProcessor struct {
	blockLS    linking.LinkSystem
	encBlockLS linking.LinkSystem
	col        *collection
	db         *DB

	// docIDs contains all docIDs and their original values
	// that have been merged so far by the mergeProcessor
	// the original values are used to update indexes
	docIDs map[client.DocID]*client.Document

	// composites is a list of composites that need to be merged.
	composites *list.List

	blockDocRefs           map[string]resolvedDocRef
	currentCompositeDocRef *resolvedDocRef
	newDocCreateMode       bool

	// Heads are only saved when the merge finishes, so within one merge they cannot tell us
	// what has already been applied.  This covers that gap.
	appliedFieldBlocks map[string]struct{}

	// Ancestors already walked back from each field's heads, keyed by document and field.  Every
	// field block of the same field asks the same question, so the walk is done once and then
	// extended only when a deeper block needs it.
	ancestorCache map[string]*ancestorSet
}

type resolvedDocRef struct {
	docID      string
	docShortID uint64
}

func (mp *mergeProcessor) resolveOrAllocateDocShortID(
	ctx context.Context,
	collectionShortID uint32,
	docID string,
) (uint64, error) {
	docShortID, found, err := id.GetDocShortID(ctx, collectionShortID, docID)
	if err != nil {
		return 0, err
	}
	if found {
		return docShortID, nil
	}

	docShortID, err = mp.db.reserveDocShortID(ctx)
	if err != nil {
		return 0, err
	}
	if err := id.SetDocIDMapping(ctx, collectionShortID, docShortID, docID); err != nil {
		return 0, err
	}
	return docShortID, nil
}

// getDocHeadstoreKey returns the headstore key under which the given document's composite heads are
// stored. The returned exists is false when the document does not yet exist locally (the merge is
// creating it), in which case it has no heads and the caller must treat the merge target as empty.
func getDocHeadstoreKey(ctx context.Context, col *collection, docID string) (keys.HeadstoreKey, bool, error) {
	collectionShortID, err := id.GetCollectionShortID(ctx, col.Version().CollectionID)
	if err != nil {
		return nil, false, err
	}

	if docID != "" {
		docShortID, found, err := id.GetDocShortID(ctx, collectionShortID, docID)
		if err != nil {
			return nil, false, err
		}
		if !found {
			return nil, false, nil
		}
		return keys.HeadstoreDocKey{
			DocShortID: docShortID,
			FieldID:    core.COMPOSITE_NAMESPACE,
		}, true, nil
	}

	return keys.NewHeadstoreColKey(collectionShortID), true, nil
}

func (db *DB) newMergeProcessor(
	ctx context.Context,
	col *collection,
	newDocCreateMode bool,
) (*mergeProcessor, error) {
	txn := datastore.CtxMustGetTxn(ctx)

	blockLS := cidlink.DefaultLinkSystem()
	blockLS.SetReadStorage(blockstore.NewIPLDStore(txn.Blockstore()))

	encBlockLS := cidlink.DefaultLinkSystem()
	encBlockLS.SetReadStorage(blockstore.NewIPLDStore(txn.Encstore()))

	return &mergeProcessor{
		blockLS:            blockLS,
		encBlockLS:         encBlockLS,
		col:                col,
		db:                 db,
		docIDs:             make(map[client.DocID]*client.Document),
		composites:         list.New(),
		blockDocRefs:       make(map[string]resolvedDocRef),
		newDocCreateMode:   newDocCreateMode,
		appliedFieldBlocks: make(map[string]struct{}),
		ancestorCache:      make(map[string]*ancestorSet),
	}, nil
}

type mergeTarget struct {
	heads map[cid.Cid]*coreblock.Block
}

// minHeight is the lowest height among the heads being compared against.
//
// Heads sit at different heights whenever the document has concurrent branches, so no single
// height describes them all.  Taking the lowest keeps every branch in scope: a block at or above
// it may still be new to one of the branches, and is collected rather than skipped.
func (mt mergeTarget) minHeight() uint64 {
	var min uint64
	first := true
	for _, b := range mt.heads {
		h := b.Delta.GetPriority()
		if first || h < min {
			min, first = h, false
		}
	}
	return min
}

func newMergeTarget() mergeTarget {
	return mergeTarget{
		heads: make(map[cid.Cid]*coreblock.Block),
	}
}

// loadComposites retrieves and stores into the merge processor the composite blocks for the given
// CID until it reaches a block that has already been merged or until we reach the genesis block.
func (mp *mergeProcessor) loadComposites(
	ctx context.Context,
	blockCid cid.Cid,
	mt mergeTarget,
) error {
	if _, ok := mt.heads[blockCid]; ok {
		// We've already processed this block.
		return nil
	}

	nd, err := mp.blockLS.Load(linking.LinkContext{Ctx: ctx}, cidlink.Link{Cid: blockCid}, coreblock.BlockSchemaPrototype)
	if err != nil {
		return NewErrLoadBlockForMerge(err, blockCid.String())
	}

	block, err := coreblock.GetFromNode(nd)
	if err != nil {
		return NewErrDecodeBlockForMerge(err, blockCid.String())
	}

	// In the simplest case, the new block or its children will link to the current head/heads (merge target)
	// of the composite DAG. However, the new block and its children might have branched off from an older block.
	// In this case, we also need to walk back the merge target's DAG until we reach a common block.
	if block.Delta.GetPriority() >= mt.minHeight() {
		mp.composites.PushFront(block)
		for _, head := range block.Heads {
			err := mp.loadComposites(ctx, head.Cid, mt)
			if err != nil {
				return NewErrLoadParentComposite(err, head.Cid.String())
			}
		}
	} else {
		newMT := newMergeTarget()
		for _, b := range mt.heads {
			for _, link := range b.Heads {
				nd, err := mp.blockLS.Load(linking.LinkContext{Ctx: ctx}, link, coreblock.BlockSchemaPrototype)
				if err != nil {
					return NewErrLoadMergeTargetBlock(err, link.String())
				}

				childBlock, err := coreblock.GetFromNode(nd)
				if err != nil {
					return NewErrDecodeMergeTargetBlock(err, link.String())
				}

				newMT.heads[link.Cid] = childBlock
			}
		}
		return mp.loadComposites(ctx, blockCid, newMT)
	}
	return nil
}

func (mp *mergeProcessor) mergeComposites(ctx context.Context) error {
	for e := mp.composites.Front(); e != nil; e = e.Next() {
		block := e.Value.(*coreblock.Block)
		link, err := block.GenerateLink()
		if err != nil {
			return NewErrGenerateMergeLink(err)
		}
		err = mp.processBlock(ctx, block, link)
		if err != nil {
			return NewErrProcessBlockMerge(err, link.String())
		}
	}

	return nil
}

// processBlock merges the block and its children to the datastore and sets the head accordingly.
func (mp *mergeProcessor) processBlock(
	ctx context.Context,
	dagBlock *coreblock.Block,
	blockLink cidlink.Link,
) error {
	block, canRead, err := coreblock.ProcessEncryptedBlock(ctx, mp.encBlockLS, dagBlock)
	if err != nil {
		return NewErrProcessEncryptedBlock(err, blockLink.String())
	}

	if canRead {
		alreadyApplied, err := mp.isAlreadyApplied(ctx, block, blockLink)
		if err != nil {
			return err
		}
		if alreadyApplied {
			return nil
		}

		shouldProcess, headstorePrefix, docRef, err := mp.mergeBlock(ctx, block, blockLink)
		if err != nil {
			return NewErrInitCRDTForMerge(err, blockLink.String())
		}

		// The field may not be known to this node - it may belong to a collection version that does not exist
		// locally.  In this case, we must ignore it as we cannot merge when we do not know what
		// kind of CRDT the field is.
		if !shouldProcess {
			return nil
		}

		var previousCompositeDocRef *resolvedDocRef
		if block.Delta.IsComposite() && docRef.docID != "" {
			previousCompositeDocRef = mp.currentCompositeDocRef
			resolved := docRef
			mp.currentCompositeDocRef = &resolved
			defer func() {
				mp.currentCompositeDocRef = previousCompositeDocRef
			}()
		}

		err = coreblock.UpdateHeads(ctx, headstorePrefix, block, blockLink)
		if err != nil {
			return NewErrProcessCRDTBlock(err, blockLink.String())
		}

		if docRef.docID != "" {
			if err := mp.setBlockDocIDMapping(ctx, docRef.docID, blockLink.Cid); err != nil {
				return err
			}
			if dagBlock.Encryption != nil {
				if err := mp.setBlockDocIDMapping(ctx, docRef.docID, dagBlock.Encryption.Cid); err != nil {
					return err
				}
			}
		}
		if block.Delta.IsComposite() && docRef.docID != "" {
			if err := mp.setLinkedBlockDocIDMappings(ctx, docRef.docID, dagBlock.Links); err != nil {
				return err
			}
		}
	}

	for _, link := range dagBlock.Links {
		nd, err := mp.blockLS.Load(linking.LinkContext{Ctx: ctx}, link.Link, coreblock.BlockSchemaPrototype)
		if err != nil {
			return NewErrLoadChildBlock(err, link.Link.String())
		}

		childBlock, err := coreblock.GetFromNode(nd)
		if err != nil {
			return NewErrDecodeChildBlock(err, link.Link.String())
		}

		if err := mp.processBlock(ctx, childBlock, link.Link); err != nil {
			return NewErrProcessChildBlock(err, link.Link.String())
		}
	}

	return nil
}

// ancestorSet holds the blocks reachable from one field's heads, down to floor.
//
// floor records how deep the walk went: a block at or below it may exist in the history without
// having been visited, so a query about such a block needs the walk extending first.
type ancestorSet struct {
	seen  map[cid.Cid]struct{}
	floor uint64
}

// isAlreadyApplied reports whether this field block's value was already counted.
//
// Such a block is reached because loadComposites can collect a composite it has already applied.
// Where the incoming branch is older than the document's heads, it walks the merge target back
// through their parents; that drops a head from the set it compares against and lowers the height
// beneath it, so the head passes both of its checks again.  Walking down from the block then
// reaches a field block applied by an earlier merge, or by the local write that created it.
//
// Applying twice is invisible for a last-write-wins field, and adds the increment twice for a
// counter.
//
// A head is only saved after its block has been applied, so anything reachable by walking back
// from a head is already counted.  That covers earlier merges; appliedFieldBlocks covers repeats
// inside the current one, whose heads are not saved yet.
//
// The blockstore's IsMerged flag looks like it would work here but does not: UpdateHeads clears
// that flag for a block's links before this recursion applies them, so a block appears merged on
// its first, correct application.  It is also keyed by block alone, while application is per
// document.
func (mp *mergeProcessor) isAlreadyApplied(
	ctx context.Context,
	block *coreblock.Block,
	blockLink cidlink.Link,
) (bool, error) {
	// Composites hold no value to double-count, and walking them is what tells the field blocks
	// below which document they belong to.
	if block.Delta.IsComposite() || block.Delta.IsCollection() {
		return false, nil
	}

	// Without knowing the document we cannot look up its heads, so let mergeBlock decide - it
	// rejects this case with a proper error.
	if mp.currentCompositeDocRef == nil {
		return false, nil
	}

	collectionShortID, err := id.GetCollectionShortID(ctx, mp.col.Version().CollectionID)
	if err != nil {
		return false, NewErrGetCollectionShortIDForMerge(err, mp.col.Version().CollectionID)
	}

	fieldName := block.Delta.GetFieldName()
	fd, ok := mp.col.Version().GetFieldByName(fieldName)
	if !ok {
		// mergeBlock already handles fields this node does not know about.
		return false, nil
	}

	fieldShortID, err := id.GetShortFieldID(ctx, collectionShortID, fd.FieldID)
	if err != nil {
		return false, NewErrGetShortFieldIDMerge(err, fd.FieldID, fieldName)
	}

	prefix := keys.DataStoreKey{
		CollectionShortID: collectionShortID,
		DocShortID:        mp.currentCompositeDocRef.docShortID,
	}.WithFieldID(fmt.Sprint(fieldShortID)).ToHeadStoreKey()

	// Catches repeats within this merge, before any head has been saved.
	appliedKey := mp.currentCompositeDocRef.docID + "/" + blockLink.Cid.String()
	if _, ok := mp.appliedFieldBlocks[appliedKey]; ok {
		return true, nil
	}

	// Comparing against the heads themselves is not enough: once later updates arrive, an
	// already-applied block sits behind the head rather than being one.
	ancestors, err := mp.fieldAncestors(ctx, prefix, block.Delta.GetPriority())
	if err != nil {
		return false, err
	}
	if _, ok := ancestors.seen[blockLink.Cid]; ok {
		return true, nil
	}

	mp.appliedFieldBlocks[appliedKey] = struct{}{}
	return false, nil
}

// fieldAncestors returns the blocks reachable from this field's heads, walked down to floor.
func (mp *mergeProcessor) fieldAncestors(
	ctx context.Context,
	prefix keys.HeadstoreKey,
	floor uint64,
) (*ancestorSet, error) {
	cacheKey := string(prefix.Bytes())
	cached, ok := mp.ancestorCache[cacheKey]
	if ok && cached.floor <= floor {
		return cached, nil
	}

	txn := datastore.CtxMustGetTxn(ctx)
	headset := coreblock.NewHeadSet(txn.Headstore(), prefix)

	heads, _, err := headset.List(ctx)
	if err != nil {
		return nil, err
	}

	seen, err := mp.walkAncestors(ctx, heads, floor)
	if err != nil {
		return nil, err
	}

	set := &ancestorSet{seen: seen, floor: floor}
	mp.ancestorCache[cacheKey] = set
	return set, nil
}

// walkAncestors collects every block reachable back from the given heads, stopping at floor.
//
// Links only ever point backwards, so a block below floor cannot lead to one at floor or above.
// Not descending past it is what keeps the walk off the whole history.
func (mp *mergeProcessor) walkAncestors(
	ctx context.Context,
	heads []cid.Cid,
	floor uint64,
) (map[cid.Cid]struct{}, error) {
	seen := make(map[cid.Cid]struct{}, len(heads))
	stack := make([]cid.Cid, 0, len(heads))
	stack = append(stack, heads...)

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if _, ok := seen[current]; ok {
			continue
		}
		seen[current] = struct{}{}

		nd, err := mp.blockLS.Load(
			linking.LinkContext{Ctx: ctx},
			cidlink.Link{Cid: current},
			coreblock.BlockSchemaPrototype,
		)
		if err != nil {
			// Missing locally, so it cannot show that anything below it was applied.  Leaving it
			// out risks a repeat, not a lost update.
			continue
		}
		currentBlock, err := coreblock.GetFromNode(nd)
		if err != nil {
			return nil, err
		}
		if currentBlock.Delta.GetPriority() <= floor {
			continue
		}
		for _, l := range currentBlock.AllLinks() {
			stack = append(stack, l.Cid)
		}
	}

	return seen, nil
}

func (mp *mergeProcessor) setBlockDocIDMapping(
	ctx context.Context,
	docID string,
	blockCID cid.Cid,
) error {
	if docID == "" || !blockCID.Defined() {
		return nil
	}

	return id.SetBlockDocIDMapping(ctx, blockCID, docID)
}

func (mp *mergeProcessor) setLinkedBlockDocIDMappings(
	ctx context.Context,
	docID string,
	links []coreblock.DAGLink,
) error {
	if docID == "" || len(links) == 0 {
		return nil
	}

	for _, link := range links {
		if err := id.SetBlockDocIDMapping(ctx, link.Cid, docID); err != nil {
			return err
		}
	}
	return nil
}

func (mp *mergeProcessor) mergeBlock(
	ctx context.Context,
	block *coreblock.Block,
	blockLink cidlink.Link,
) (bool, keys.HeadstoreKey, resolvedDocRef, error) {
	txn := datastore.CtxMustGetTxn(ctx)

	crdtUnion := block.Delta

	collectionShortID, err := id.GetCollectionShortID(ctx, mp.col.Version().CollectionID)
	if err != nil {
		return false, nil, resolvedDocRef{}, NewErrGetCollectionShortIDForMerge(err, mp.col.Version().CollectionID)
	}

	switch {
	case crdtUnion.IsComposite():
		docRef, err := mp.resolveCompositeBlockDocRef(
			ctx,
			collectionShortID,
			block,
			blockLink.Cid,
		)
		if err != nil {
			return false, nil, resolvedDocRef{}, NewErrParseDocIDMerge(err, blockLink.Cid.String())
		}
		docID, err := client.NewDocIDFromString(docRef.docID)
		if err != nil {
			return false, nil, resolvedDocRef{}, err
		}
		err = mp.trackMergedDocument(ctx, docID)
		if err != nil {
			return false, nil, resolvedDocRef{}, err
		}
		c := crdt.NewDocComposite()

		err = c.Merge(
			ctx,
			txn.Datastore(),
			keys.PrimaryDataStoreKey{
				CollectionShortID: collectionShortID,
				DocShortID:        docRef.docShortID,
			},
			block.Delta.GetDelta(),
		)
		if err != nil {
			return false,
				nil,
				resolvedDocRef{},
				NewErrProcessCRDTBlock(coreblock.NewErrMergingDelta(blockLink.Cid, err), blockLink.String())
		}

		return true,
			keys.DataStoreKey{
				CollectionShortID: collectionShortID,
				DocShortID:        docRef.docShortID,
			}.WithFieldID(core.COMPOSITE_NAMESPACE).ToHeadStoreKey(),
			docRef,
			nil

	case crdtUnion.IsCollection():
		// no-op: collection value blocks are not merged

		return true, keys.NewHeadstoreColKey(collectionShortID), resolvedDocRef{}, nil

	default:
		// A field block is always processed as a child of its composite block, which records
		// the owning document in currentCompositeDocRef. A field block's delta must be merged
		// into that document - never one resolved from the block-CID owner index, since a field
		// block can be shared across documents.
		if mp.currentCompositeDocRef == nil {
			return false, nil, resolvedDocRef{}, NewErrParseDocIDMerge(client.ErrMalformedDocID, blockLink.Cid.String())
		}
		docRef := *mp.currentCompositeDocRef
		docID, err := client.NewDocIDFromString(docRef.docID)
		if err != nil {
			return false, nil, resolvedDocRef{}, err
		}
		err = mp.trackMergedDocument(ctx, docID)
		if err != nil {
			return false, nil, resolvedDocRef{}, err
		}

		field := crdtUnion.GetFieldName()
		fd, ok := mp.col.Version().GetFieldByName(field)
		if !ok {
			// The field may not be known to this node - it may belong to a collection version that does not exist
			// locally.  In this case, return nil and have the calling code ignore it.  We cannot merge when we do
			// not know what kind of CRDT the field is.
			return false, nil, resolvedDocRef{}, nil
		}

		fieldShortID, err := id.GetShortFieldID(ctx, collectionShortID, fd.FieldID)
		if err != nil {
			return false, nil, resolvedDocRef{}, NewErrGetShortFieldIDMerge(err, fd.FieldID, field)
		}

		fieldCRDT, ok := crdt.TryGetFieldCRDT(fd.Typ)
		if !ok {
			return false, nil, resolvedDocRef{}, client.NewErrUnknownCRDT(fd.Typ)
		}

		err = fieldCRDT.Merge(
			ctx,
			txn.Datastore(),
			keys.DataStoreKey{
				CollectionShortID: collectionShortID,
				DocShortID:        docRef.docShortID,
			}.WithFieldID(fmt.Sprint(fieldShortID)),
			fd.Kind,
			block.Delta.GetDelta(),
		)
		if err != nil {
			return false,
				nil,
				resolvedDocRef{},
				NewErrProcessCRDTBlock(coreblock.NewErrMergingDelta(blockLink.Cid, err), blockLink.String())
		}

		return true, keys.DataStoreKey{
			CollectionShortID: collectionShortID,
			DocShortID:        docRef.docShortID,
		}.WithFieldID(fmt.Sprint(fieldShortID)).ToHeadStoreKey(), docRef, nil
	}
}

func (mp *mergeProcessor) resolveCompositeBlockDocRef(
	ctx context.Context,
	collectionShortID uint32,
	block *coreblock.Block,
	blockCID cid.Cid,
) (resolvedDocRef, error) {
	if resolved, ok := mp.blockDocRefs[blockCID.String()]; ok {
		return resolved, nil
	}

	// A composite block is owned by exactly one document. Use the recorded owner as a fast
	// path only when it is unambiguous; otherwise determine the DocID from the block itself:
	// a genesis composite's CID is the DocID, an update inherits it from the genesis reached
	// through its heads.
	owners, err := id.GetDocIDsForBlockFromStore(
		ctx,
		datastore.CtxMustGetTxn(ctx).Systemstore(),
		blockCID,
	)
	if err != nil {
		return resolvedDocRef{}, err
	}
	if len(owners) == 1 {
		return mp.resolveAndCacheBlockDocRef(ctx, collectionShortID, blockCID, owners[0])
	}

	if len(block.Heads) == 0 {
		return mp.resolveAndCacheBlockDocRef(ctx, collectionShortID, blockCID, client.NewDocIDV0(blockCID).String())
	}

	for _, head := range block.Heads {
		resolved, err := mp.resolveDocRefForCompositeCID(ctx, collectionShortID, head.Cid)
		if err != nil {
			return resolvedDocRef{}, err
		}
		if resolved.docID != "" {
			mp.blockDocRefs[blockCID.String()] = resolved
			return resolved, nil
		}
	}

	return resolvedDocRef{}, client.ErrMalformedDocID
}

func (mp *mergeProcessor) resolveDocRefForCompositeCID(
	ctx context.Context,
	collectionShortID uint32,
	blockCID cid.Cid,
) (resolvedDocRef, error) {
	if resolved, ok := mp.blockDocRefs[blockCID.String()]; ok {
		return resolved, nil
	}

	// A composite block is owned by exactly one document. Use the recorded owner as a fast
	// path only when it is unambiguous; otherwise load the block and determine the DocID from
	// the composite itself.
	owners, err := id.GetDocIDsForBlockFromStore(
		ctx,
		datastore.CtxMustGetTxn(ctx).Systemstore(),
		blockCID,
	)
	if err != nil {
		return resolvedDocRef{}, err
	}
	if len(owners) == 1 {
		return mp.resolveAndCacheBlockDocRef(ctx, collectionShortID, blockCID, owners[0])
	}

	nd, err := mp.blockLS.Load(linking.LinkContext{Ctx: ctx}, cidlink.Link{Cid: blockCID}, coreblock.BlockSchemaPrototype)
	if err != nil {
		return resolvedDocRef{}, err
	}
	block, err := coreblock.GetFromNode(nd)
	if err != nil {
		return resolvedDocRef{}, err
	}
	if !block.Delta.IsComposite() {
		return resolvedDocRef{}, client.ErrMalformedDocID
	}
	return mp.resolveCompositeBlockDocRef(ctx, collectionShortID, block, blockCID)
}

func (mp *mergeProcessor) resolveAndCacheBlockDocRef(
	ctx context.Context,
	collectionShortID uint32,
	blockCID cid.Cid,
	docID string,
) (resolvedDocRef, error) {
	docShortID, err := mp.resolveOrAllocateDocShortID(ctx, collectionShortID, docID)
	if err != nil {
		return resolvedDocRef{}, err
	}
	resolved := resolvedDocRef{docID: docID, docShortID: docShortID}
	mp.blockDocRefs[blockCID.String()] = resolved
	return resolved, nil
}

// trackMergedDocument tracks the current version of the document so we
// can correctly sync indexes after a merge.
func (mp *mergeProcessor) trackMergedDocument(ctx context.Context, docID client.DocID) error {
	if len(mp.col.indexes) == 0 {
		mp.docIDs[docID] = nil
		return nil
	}
	_, exists := mp.docIDs[docID]
	if exists {
		return nil
	}
	if mp.newDocCreateMode {
		mp.docIDs[docID] = nil
		return nil
	}
	doc, err := getDocForMerge(ctx, mp.col, docID)
	if err != nil && !errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized) {
		return err
	}
	mp.docIDs[docID] = doc
	return nil
}

// getDocForMerge fetches a doc during inbound merge without the ACP read filter.
// The merge ctx has no caller identity, so GetDocument would deny access and
// return nil, silently skipping the secondary-index Save. Access was already
// gated at the P2P boundary, so we read directly.
func getDocForMerge(
	ctx context.Context,
	col *collection,
	docID client.DocID,
) (*client.Document, error) {
	primaryKey, err := col.getPrimaryKeyFromDocID(ctx, docID)
	if err != nil {
		return nil, err
	}
	return col.getInternal(ctx, primaryKey, nil, false)
}

func getCollectionFromCollectionID(ctx context.Context, db *DB, collectionID string) (*collection, error) {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return nil, err
	}
	defer txn.Discard()

	cols, err := db.getCollections(
		ctx,
		utils.NewOptions(options.GetCollections().SetCollectionID(collectionID)),
		true,
	)
	if err != nil {
		return nil, err
	}
	if len(cols) == 0 {
		return nil, client.NewErrCollectionNotFoundForRoot(collectionID)
	}
	// We currently only support one active collection per collection root
	// so it is safe to return the first one.
	return cols[0].(*collection), nil
}

// getHeadsAsMergeTarget retrieves the heads of the composite DAG for the given document
// and returns them as a merge target.
func getHeadsAsMergeTarget(ctx context.Context, key keys.HeadstoreKey) (mergeTarget, error) {
	cids, err := getHeads(ctx, key)

	if err != nil {
		return mergeTarget{}, NewErrGetHeadsForMerge(err, string(key.Bytes()))
	}

	mt := newMergeTarget()
	for _, cid := range cids {
		block, err := loadBlockFromBlockStore(ctx, cid)
		if err != nil {
			return mergeTarget{}, err
		}

		mt.heads[cid] = block
	}
	return mt, nil
}

// getHeads retrieves the heads associated with the given datastore key.
func getHeads(ctx context.Context, key keys.HeadstoreKey) ([]cid.Cid, error) {
	txn := datastore.CtxMustGetTxn(ctx)
	headset := coreblock.NewHeadSet(txn.Headstore(), key)

	cids, _, err := headset.List(ctx)
	if err != nil {
		return nil, err
	}

	return cids, nil
}

// loadBlockFromBlockStore loads a block from the blockstore.
func loadBlockFromBlockStore(ctx context.Context, cid cid.Cid) (*coreblock.Block, error) {
	txn := datastore.CtxMustGetTxn(ctx)
	b, err := txn.Blockstore().Get(ctx, cid)
	if err != nil {
		return nil, NewErrLoadBlockFromStore(err, cid.String())
	}

	block, err := coreblock.GetFromBytes(b.RawData())
	if err != nil {
		return nil, NewErrDecodeBlockFromStore(err, cid.String())
	}

	return block, nil
}

func syncIndexedDoc(
	ctx context.Context,
	docID client.DocID,
	col *collection,
	oldDoc *client.Document,
) error {
	newDoc, err := getDocForMerge(ctx, col, docID)
	if err != nil && !errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized) {
		return err
	}
	// Both can be nil during concurrent P2P operations (e.g. delete + update)
	// where the document was already deleted and no prior indexed state exists.
	if oldDoc == nil && newDoc == nil {
		log.InfoContext(ctx, "skipping index update: no document found", corelog.String("docID", docID.String()))
		return nil
	}
	if oldDoc != nil && newDoc != nil {
		return col.updateDocIndex(ctx, oldDoc, newDoc)
	} else if oldDoc == nil {
		return col.addDocToIndex(ctx, newDoc)
	} else {
		return col.deleteIndexedDoc(ctx, oldDoc)
	}
}
