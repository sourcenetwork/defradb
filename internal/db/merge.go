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
	"sort"
	"strings"
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
	"github.com/sourcenetwork/defradb/internal/db/blockowner"
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
	for i := 0; i < db.txnAttempts(); i++ {
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

// mergeChunkSize bounds how many events share a transaction. Badger re-sorts the
// transaction's pending writes on every iterator open, and each merged document opens
// several, so a bigger chunk sorts a bigger set more times.
const mergeChunkSize = 8

// Phases a merge chunk can fail in, reported on the retry-exhaustion log line.
const (
	phaseRead   = "read"
	phaseCommit = "commit"
)

type mergeEntry struct {
	evt event.Merge
	col *collection
	// index of the event in the caller's slice, so its outcome can be reported back.
	index int
}

// MergeBatchWithTxn merges events in chunks of at most mergeChunkSize. All per-key
// locks are acquired upfront and held for the lifetime of the call, so callers must
// ensure the merges are independent (different docIDs and collectionIDs).
//
// A chunk that fails is re-run one event at a time. A deterministic failure then drops
// only the event that caused it, and a chunk that exhausted its retry budget is retried
// over a smaller write set.
//
// The returned slice is parallel to merges and reports which events committed. An event
// that did not commit is not stored, so callers must not relay it onward as merged. The
// error names every dropped event and is nil when all of them committed.
func (db *DB) MergeBatchWithTxn(ctx context.Context, merges []event.Merge) ([]bool, error) {
	merged := make([]bool, len(merges))
	if len(merges) == 0 {
		return merged, nil
	}

	var errs []error
	entries := make([]mergeEntry, 0, len(merges))
	for i, evt := range merges {
		col, err := getCollectionFromCollectionID(ctx, db, evt.CollectionID)
		if err != nil {
			errs = append(errs, NewErrMergeEventDropped(err, evt.DocID, evt.Cid.String()))
			db.stats.markDropped(dropCollection)
			continue
		}
		entries = append(entries, mergeEntry{evt: evt, col: col, index: i})
	}

	// Collect the unique lock keys and their queues, sorted for deadlock-safe ordering.
	type lockKey struct {
		key   string
		queue *mergeQueue
	}
	seen := make(map[string]struct{}, len(entries))
	locks := make([]lockKey, 0, len(entries))
	docLocked := false
	colLocked := false
	for _, e := range entries {
		var k string
		var seenKey string
		var q *mergeQueue
		if e.col.Version().IsBranchable {
			k = e.evt.CollectionID
			seenKey = "c:" + k // namespace prefix avoids collision with docID keys
			q = db.colMergeQueue
			colLocked = true
		} else {
			k = e.evt.DocID
			seenKey = "d:" + k
			q = db.docMergeQueue
			docLocked = true
		}
		if _, ok := seen[seenKey]; !ok {
			seen[seenKey] = struct{}{}
			locks = append(locks, lockKey{key: k, queue: q})
		}
	}
	sort.Slice(locks, func(i, j int) bool { return locks[i].key < locks[j].key })

	// Acquire one semaphore slot per queue used to count the batch as one merge.
	if docLocked {
		<-db.docMergeQueue.sem
		defer func() { db.docMergeQueue.sem <- struct{}{} }()
	}
	if colLocked {
		<-db.colMergeQueue.sem
		defer func() { db.colMergeQueue.sem <- struct{}{} }()
	}

	// Acquire per-key locks without consuming additional semaphore slots.
	for _, lk := range locks {
		lk.queue.addNoSem(lk.key)
	}
	defer func() {
		for _, lk := range locks {
			lk.queue.doneNoSem(lk.key)
		}
	}()

	for start := 0; start < len(entries); start += mergeChunkSize {
		end := min(start+mergeChunkSize, len(entries))
		chunk := entries[start:end]

		if err := db.mergeChunk(ctx, chunk); err == nil {
			db.publishMergeComplete(chunk)
			for _, e := range chunk {
				merged[e.index] = true
			}
			continue
		}

		// Isolate the failure so the events that can merge still land.
		for i := range chunk {
			if err := db.mergeChunk(ctx, chunk[i:i+1]); err != nil {
				errs = append(errs, NewErrMergeEventDropped(err, chunk[i].evt.DocID, chunk[i].evt.Cid.String()))
				db.stats.markDropped(mergeDropReason(err))
				continue
			}
			db.publishMergeComplete(chunk[i : i+1])
			merged[chunk[i].index] = true
		}
	}

	return merged, errors.Join(errs...)
}

// txnAttempts is how many times to try a transaction before giving up, never fewer than
// once so a zero retry budget still makes an attempt.
func (db *DB) txnAttempts() int {
	if max := db.MaxTxnRetries(); max > 1 {
		return max
	}
	return 1
}

// namedDocs renders the documents a transaction touched, as collection/docID. Badger
// reports conflicts without naming the contended key, so this is the only lead available
// for working out which documents contend with each other.
func namedDocs(entries []mergeEntry) string {
	docIDs := make([]string, len(entries))
	for i, e := range entries {
		docIDs[i] = e.col.Name() + "/" + e.evt.DocID
	}
	return strings.Join(docIDs, ",")
}

// mergeChunk merges every event of the chunk inside one transaction, retrying the
// whole chunk on transaction conflict. Isolating a failing event is the caller's job.
func (db *DB) mergeChunk(ctx context.Context, entries []mergeEntry) error {
	// Held so that exhausting the retry budget can report the conflict that caused it and
	// where the last one was raised.
	var conflictErr error
	var phase string
	// Whether each event created its document, kept until the transaction commits so a
	// retried attempt does not count its events twice.
	creates := make([]bool, 0, len(entries))
	for i := 0; i < db.txnAttempts(); i++ {
		txn, err := db.NewTxn(false)
		if err != nil {
			return err
		}
		txnCtx := InitContext(ctx, txn)

		creates = creates[:0]
		var mergeErr error
		for _, e := range entries {
			isCreate, err := db.mergeInTxn(txnCtx, e.col, e.evt)
			if err != nil {
				mergeErr, phase = err, phaseRead
				break
			}
			creates = append(creates, isCreate)
		}

		if mergeErr != nil {
			txn.Discard()
			if errors.Is(mergeErr, corekv.ErrTxnConflict) {
				conflictErr = mergeErr
				db.stats.chunkConflicts.Add(1)
				continue
			}
			return mergeErr
		}

		if err := txn.Commit(); err != nil {
			txn.Discard()
			if errors.Is(err, corekv.ErrTxnConflict) {
				conflictErr = err
				db.stats.chunkConflicts.Add(1)
				phase = phaseCommit
				continue
			}
			return err
		}

		for _, isCreate := range creates {
			db.stats.markCreateOrUpdate(isCreate)
		}
		return nil
	}

	// The chunk used its whole retry budget without committing. The caller then re-runs it
	// one event at a time, so this counts conflict pressure rather than loss. What was
	// actually lost is named in the caller's error.
	db.stats.markExhausted()

	log.InfoContext(ctx, "merge chunk exhausted its retries",
		corelog.Int("attempts", db.txnAttempts()),
		corelog.String("phase", phase),
		corelog.String("docIDs", namedDocs(entries)),
	)
	return client.NewErrMaxTxnRetries(conflictErr)
}

func (db *DB) publishMergeComplete(entries []mergeEntry) {
	for _, e := range entries {
		db.events.Publish(event.NewMessage(event.MergeCompleteName, event.MergeComplete{Merge: e.evt}))
	}
}

func (db *DB) executeMerge(ctx context.Context, col *collection, dagMerge event.Merge) error {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return NewErrCreateMergeTxn(err, dagMerge.DocID, dagMerge.Cid.String())
	}
	defer txn.Discard()

	isCreate, err := db.mergeInTxn(ctx, col, dagMerge)
	if err != nil {
		return err
	}

	if err := txn.Commit(); err != nil {
		return err
	}
	db.stats.markCreateOrUpdate(isCreate)

	// send a complete event so we can track merges in the integration tests
	db.events.Publish(event.NewMessage(event.MergeCompleteName, event.MergeComplete{Merge: dagMerge}))
	return nil
}

// mergeInTxn executes the merge logic for a single event using the transaction already
// present on ctx.  It does not commit; the caller is responsible for committing.
//
// Reports whether the event created the document rather than updating one already held,
// which the caller counts once the transaction commits.
func (db *DB) mergeInTxn(ctx context.Context, col *collection, dagMerge event.Merge) (bool, error) {
	key, exists, err := getDocHeadstoreKey(ctx, col, dagMerge.DocID)
	if err != nil {
		return false, err
	}

	mt := newMergeTarget()
	if exists {
		mt, err = getHeadsAsMergeTarget(ctx, key)
		if err != nil {
			return false, NewErrGetMergeTargetHeads(err, dagMerge.DocID, string(key.Bytes()))
		}
	}

	// No local heads means the merge is creating the document rather than updating one
	// that already exists here.
	newDocCreateMode := len(mt.heads) == 0

	mp, err := db.newMergeProcessor(ctx, col, newDocCreateMode)
	if err != nil {
		return false, err
	}

	if err = mp.loadComposites(ctx, dagMerge.Cid, mt); err != nil {
		return false, NewErrLoadComposites(err, dagMerge.Cid.String(), dagMerge.DocID)
	}

	if err = mp.mergeComposites(ctx); err != nil {
		return false, NewErrMergeComposites(err, dagMerge.DocID)
	}

	for docID, oldDoc := range mp.docIDs {
		if err = syncIndexedDoc(ctx, docID, mp.col, oldDoc); err != nil {
			return false, NewErrSyncIndexedDoc(err, docID.String())
		}
	}

	return newDocCreateMode, nil
}

const maxConcurrentMerges = 32

// mergeQueue is synchronization source to ensure that concurrent
// document merges do not cause transaction conflicts.
type mergeQueue struct {
	keys  map[string]chan struct{}
	mutex sync.Mutex
	// sem limits the total number of merges running concurrently across all keys,
	// providing backpressure when many documents are being merged simultaneously.
	sem chan struct{}
}

func newMergeQueue() *mergeQueue {
	sem := make(chan struct{}, maxConcurrentMerges)
	for range maxConcurrentMerges {
		sem <- struct{}{}
	}
	return &mergeQueue{
		keys: make(map[string]chan struct{}),
		sem:  sem,
	}
}

// add adds a key to the queue. If the key is already in the queue, it will
// wait for the key to be removed from the queue. For every add call, done must
// be called to remove the key from the queue. Otherwise, subsequent add calls will
// block forever.
// add also acquires a slot from the global semaphore to bound concurrency.
func (m *mergeQueue) add(key string) {
	<-m.sem // acquire global slot before waiting for per-key turn
	m.mutex.Lock()
	done, ok := m.keys[key]
	if !ok {
		m.keys[key] = make(chan struct{})
	}
	m.mutex.Unlock()
	if ok {
		m.sem <- struct{}{} // release slot while waiting for per-key turn
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
	m.sem <- struct{}{} // release global slot
}

// addNoSem acquires the per-key slot without consuming a semaphore slot.
// Use when the caller already holds a batch-level concurrency limit.
func (m *mergeQueue) addNoSem(key string) {
	m.mutex.Lock()
	done, ok := m.keys[key]
	if !ok {
		m.keys[key] = make(chan struct{})
	}
	m.mutex.Unlock()
	if ok {
		<-done
		m.addNoSem(key)
	}
}

// doneNoSem releases the per-key slot without touching the semaphore.
// The caller is responsible for releasing any semaphore slot it holds.
func (m *mergeQueue) doneNoSem(key string) {
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
		blockLS:          blockLS,
		encBlockLS:       encBlockLS,
		col:              col,
		db:               db,
		docIDs:           make(map[client.DocID]*client.Document),
		composites:       list.New(),
		blockDocRefs:     make(map[string]resolvedDocRef),
		newDocCreateMode: newDocCreateMode,
	}, nil
}

type mergeTarget struct {
	heads      map[cid.Cid]*coreblock.Block
	headHeight uint64
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
	if block.Delta.GetPriority() >= mt.headHeight {
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
				newMT.headHeight = childBlock.Delta.GetPriority()
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
		crdt, docRef, err := mp.initCRDTForType(ctx, block, blockLink)
		if err != nil {
			return NewErrInitCRDTForMerge(err, blockLink.String())
		}

		// A signature block is not in AllLinks, so nothing else records who owns it.
		if dagBlock.Signature != nil {
			if err := mp.setBlockDocIDMapping(ctx, docRef.docID, dagBlock.Signature.Cid); err != nil {
				return err
			}
		}

		// If the CRDT is nil, it means the field is not part
		// of the collection definition and we can safely ignore it.
		if crdt == nil {
			// The block is owned from here on but never reaches updateHeads, so its marker
			// is cleared here. Taking ownership and clearing the marker in one transaction
			// is what lets a concurrent sweep conflict rather than reclaim the block.
			if dagBlock.Signature != nil {
				txn := datastore.CtxMustGetTxn(ctx)
				if err := txn.Blockstore().MarkAsMerged(ctx, dagBlock.Signature.Cid); err != nil {
					return err
				}
			}
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

		err = coreblock.ProcessBlock(ctx, crdt, block, blockLink)
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

func (mp *mergeProcessor) initCRDTForType(
	ctx context.Context,
	block *coreblock.Block,
	blockLink cidlink.Link,
) (crdt.ReplicatedData, resolvedDocRef, error) {
	txn := datastore.CtxMustGetTxn(ctx)
	crdtUnion := block.Delta

	collectionShortID, err := id.GetCollectionShortID(ctx, mp.col.Version().CollectionID)
	if err != nil {
		return nil, resolvedDocRef{}, NewErrGetCollectionShortIDForMerge(err, mp.col.Version().CollectionID)
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
			return nil, resolvedDocRef{}, NewErrParseDocIDMerge(err, blockLink.Cid.String())
		}
		docID, err := client.NewDocIDFromString(docRef.docID)
		if err != nil {
			return nil, resolvedDocRef{}, err
		}
		err = mp.trackMergedDocument(ctx, docID)
		if err != nil {
			return nil, resolvedDocRef{}, err
		}
		return crdt.NewDocComposite(
			txn.Datastore(),
			mp.col.Version().VersionID,
			keys.DataStoreKey{
				CollectionShortID: collectionShortID,
				DocShortID:        docRef.docShortID,
			}.WithFieldID(core.COMPOSITE_NAMESPACE),
		), docRef, nil

	case crdtUnion.IsCollection():
		return crdt.NewCollection(
			mp.col.Version().VersionID,
			keys.NewHeadstoreColKey(collectionShortID),
		), resolvedDocRef{}, nil

	default:
		// A field block is always processed as a child of its composite block, which records
		// the owning document in currentCompositeDocRef. A field block's delta must be merged
		// into that document - never one resolved from the block-CID owner index, since a field
		// block can be shared across documents.
		if mp.currentCompositeDocRef == nil {
			return nil, resolvedDocRef{}, NewErrParseDocIDMerge(client.ErrMalformedDocID, blockLink.Cid.String())
		}
		docRef := *mp.currentCompositeDocRef
		docID, err := client.NewDocIDFromString(docRef.docID)
		if err != nil {
			return nil, resolvedDocRef{}, err
		}
		err = mp.trackMergedDocument(ctx, docID)
		if err != nil {
			return nil, resolvedDocRef{}, err
		}

		field := crdtUnion.GetFieldName()
		fd, ok := mp.col.Version().GetFieldByName(field)
		if !ok {
			// The field is not part of the collection definition, so there is no delta to
			// merge. The document is still returned: the block belongs to it either way, and
			// the caller records that ownership.
			return nil, docRef, nil
		}

		fieldShortID, err := id.GetShortFieldID(ctx, collectionShortID, fd.FieldID)
		if err != nil {
			return nil, resolvedDocRef{}, NewErrGetShortFieldIDMerge(err, fd.FieldID, field)
		}

		fieldCRDT, err := crdt.FieldLevelCRDTWithStore(
			txn.Datastore(),
			mp.col.Version().VersionID,
			fd.Typ,
			fd.Kind,
			keys.DataStoreKey{
				CollectionShortID: collectionShortID,
				DocShortID:        docRef.docShortID,
			}.WithFieldID(fmt.Sprint(fieldShortID)),
			field,
		)
		if err != nil {
			return nil, resolvedDocRef{}, err
		}
		return fieldCRDT, docRef, nil
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
	owners, err := blockowner.DocIDs(
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
	owners, err := blockowner.DocIDs(
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
		// All heads have the same height so overwriting is ok.
		mt.headHeight = block.Delta.GetPriority()
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
