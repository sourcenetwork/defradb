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
	"sync"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/blockstore"
	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/encryption"
	"github.com/sourcenetwork/defradb/internal/keys"
)

const (
	// mergeBlockBatchSize is the maximum number of blocks to process in a single transaction during merge operations.
	mergeBlockBatchSize = 50
)

func (db *DB) Merge(ctx context.Context, evt event.Merge) error {
	col, err := getCollectionFromCollectionID(ctx, db, evt.CollectionID)
	if err != nil {
		log.ErrorContextE(
			ctx,
			"Failed to execute merge",
			err,
			corelog.Any("Event", evt))
		return err
	}

	if col.Version().IsBranchable {
		// As collection commits link to document composite commits, all events
		// recieved for branchable collections must be processed serially else
		// they may otherwise cause a transaction conflict.
		if err := db.colMergeQueue.add(ctx, evt.CollectionID); err != nil {
			return err
		}
		defer db.colMergeQueue.done(evt.CollectionID)
	} else {
		// ensure only one merge per docID
		if err := db.docMergeQueue.add(ctx, evt.DocID); err != nil {
			return err
		}
		defer db.docMergeQueue.done(evt.DocID)
	}

	// retry the merge process if a conflict occurs
	//
	// conficts occur when a user updates a document
	// while a merge is in progress.
	for i := 0; i < db.MaxTxnRetries(); i++ {
		err = db.executeMerge(ctx, col, evt)
		if errors.Is(err, corekv.ErrTxnConflict) {
			continue
		}
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

// MergeBatchWithTxn merges multiple events using an existing transaction from context.
func (db *DB) MergeBatchWithTxn(ctx context.Context, evts []event.Merge) (func(), error) {
	if len(evts) == 0 {
		return nil, nil
	}

	byCollection := make(map[string][]event.Merge)
	for _, evt := range evts {
		byCollection[evt.CollectionID] = append(byCollection[evt.CollectionID], evt)
	}

	for collectionID, colEvts := range byCollection {
		col, err := getCollectionFromCollectionID(ctx, db, collectionID)
		if err != nil {
			log.ErrorContextE(
				ctx,
				"Failed to get collection for batch merge with txn",
				err,
				corelog.String("CollectionID", collectionID))
			continue
		}

		if col.Version().IsBranchable {
			for _, evt := range colEvts {
				if err := db.mergeWithTxn(ctx, col, evt); err != nil {
					log.ErrorContextE(ctx, "Failed to merge branchable event with txn", err)
				}
			}
			continue
		}

		batchResult, err := db.executeMergeBatchWithTxn(ctx, col, colEvts)
		if err != nil {
			log.ErrorContextE(
				ctx,
				"Failed to execute batch merge with txn",
				err,
				corelog.String("CollectionID", collectionID),
				corelog.Int("EventCount", len(colEvts)))
			continue
		}

		if batchResult != nil && len(batchResult.DocIDs) > 0 && len(col.indexes) > 0 {
			for _, docIDStr := range batchResult.DocIDs {
				docID, err := client.NewDocIDFromString(docIDStr)
				if err != nil {
					log.ErrorContextE(ctx, "Failed to parse docID for index sync", err)
					continue
				}
				oldDoc := batchResult.OldDocs[docIDStr]

				if oldDoc == nil {
					if fields, ok := batchResult.MergedFields[docIDStr]; ok && len(fields) > 0 {
						newDoc, buildErr := buildDocFromMergeFields(ctx, docIDStr, col, fields)
						if buildErr != nil {
							log.ErrorContextE(ctx, "Failed to build doc from merge data, falling back to read", buildErr)
							if err := syncIndexedDoc(ctx, docID, col, oldDoc); err != nil {
								log.ErrorContextE(ctx, "Index sync after merge failed", err,
									corelog.String("DocID", docIDStr))
							}
							continue
						}
						if err := syncDocIndex(ctx, col, nil, newDoc); err != nil {
							log.ErrorContextE(ctx, "Index sync from merge data failed", err,
								corelog.String("DocID", docIDStr))
						}
						continue
					}
				}

				if err := syncIndexedDoc(ctx, docID, col, oldDoc); err != nil {
					log.ErrorContextE(ctx, "Index sync after merge failed", err,
						corelog.String("DocID", docIDStr))
				}
			}
		}
	}

	return nil, nil
}

// executeMergeBatchWithTxn merges events using an existing transaction from context.
func (db *DB) executeMergeBatchWithTxn(
	ctx context.Context, col *collection, evts []event.Merge,
) (*MergeBatchResult, error) {
	docIDs := make([]string, 0, len(evts))
	for _, evt := range evts {
		if evt.DocID != "" {
			docIDs = append(docIDs, evt.DocID)
		}
	}

	sortedDocIDs := make([]string, len(docIDs))
	copy(sortedDocIDs, docIDs)
	sort.Strings(sortedDocIDs)

	addedDocIDs := make([]string, 0, len(sortedDocIDs))
	for _, docID := range sortedDocIDs {
		if err := db.docMergeQueue.add(ctx, docID); err != nil {
			for _, added := range addedDocIDs {
				db.docMergeQueue.done(added)
			}
			return nil, err
		}
		addedDocIDs = append(addedDocIDs, docID)
	}
	defer func() {
		for _, docID := range addedDocIDs {
			db.docMergeQueue.done(docID)
		}
	}()

	return db.executeMergeBatchWritesOnly(ctx, col, evts)
}

// MergeBatchResult contains the result of a batch merge operation.
type MergeBatchResult struct {
	CollectionID string
	DocIDs       []string
	OldDocs      map[string]*client.Document
	MergedFields map[string]map[string][]byte
}

// executeMergeBatchWritesOnly performs batch merge writes only, returning docIDs for later index sync.
func (db *DB) executeMergeBatchWritesOnly(
	ctx context.Context, col *collection, evts []event.Merge,
) (*MergeBatchResult, error) {
	allDocIDs := make(map[string]*client.Document)
	var allMergedFields map[string]map[string][]byte

	for _, dagMerge := range evts {
		var key keys.HeadstoreKey
		if dagMerge.DocID != "" {
			key = keys.HeadstoreDocKey{
				DocID:   dagMerge.DocID,
				FieldID: core.COMPOSITE_NAMESPACE,
			}
		} else {
			shortID, err := id.GetShortCollectionID(ctx, col.Version().CollectionID)
			if err != nil {
				return nil, err
			}
			key = keys.NewHeadstoreColKey(shortID)
		}

		mt, err := getHeadsAsMergeTarget(ctx, key)
		if err != nil {
			return nil, err
		}

		mergeCtx := ctx
		if len(mt.heads) == 0 {
			mergeCtx = coreblock.ContextWithNewDocCreateMode(ctx)
		}

		mp, err := db.newMergeProcessor(mergeCtx, col)
		if err != nil {
			return nil, err
		}

		err = mp.loadComposites(mergeCtx, dagMerge.Cid, mt)
		if err != nil {
			return nil, err
		}

		_, err = mp.mergeComposites(mergeCtx)
		if err != nil {
			return nil, err
		}

		for docID, oldDoc := range mp.docIDs {
			docIDStr := docID.String()
			allDocIDs[docIDStr] = oldDoc
		}

		for docIDStr, fields := range mp.mergedFields {
			if allMergedFields == nil {
				allMergedFields = make(map[string]map[string][]byte)
			}
			allMergedFields[docIDStr] = fields
		}
	}

	sortedDocIDs := make([]string, 0, len(allDocIDs))
	for docIDStr := range allDocIDs {
		sortedDocIDs = append(sortedDocIDs, docIDStr)
	}
	sort.Strings(sortedDocIDs)

	return &MergeBatchResult{
		CollectionID: col.Version().CollectionID,
		DocIDs:       sortedDocIDs,
		OldDocs:      allDocIDs,
		MergedFields: allMergedFields,
	}, nil
}

// mergeWithTxn performs a single merge using an existing transaction from context.
func (db *DB) mergeWithTxn(ctx context.Context, col *collection, evt event.Merge) error {
	var key keys.HeadstoreKey
	if evt.DocID != "" {
		key = keys.HeadstoreDocKey{
			DocID:   evt.DocID,
			FieldID: core.COMPOSITE_NAMESPACE,
		}
	} else {
		shortID, err := id.GetShortCollectionID(ctx, col.Version().CollectionID)
		if err != nil {
			return err
		}
		key = keys.NewHeadstoreColKey(shortID)
	}

	mt, err := getHeadsAsMergeTarget(ctx, key)
	if err != nil {
		return err
	}

	mergeCtx := ctx
	if len(mt.heads) == 0 {
		mergeCtx = coreblock.ContextWithNewDocCreateMode(ctx)
	}

	mp, err := db.newMergeProcessor(mergeCtx, col)
	if err != nil {
		return err
	}

	err = mp.loadComposites(mergeCtx, evt.Cid, mt)
	if err != nil {
		return err
	}

	ctx, err = mp.mergeComposites(mergeCtx)
	if err != nil {
		return err
	}

	if err := mp.syncIndexes(ctx); err != nil {
		return err
	}

	return nil
}

func (db *DB) executeMerge(ctx context.Context, col *collection, dagMerge event.Merge) error {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	var key keys.HeadstoreKey
	if dagMerge.DocID != "" {
		key = keys.HeadstoreDocKey{
			DocID:   dagMerge.DocID,
			FieldID: core.COMPOSITE_NAMESPACE,
		}
	} else {
		shortID, err := id.GetShortCollectionID(ctx, col.Version().CollectionID)
		if err != nil {
			return err
		}

		key = keys.NewHeadstoreColKey(shortID)
	}

	mt, err := getHeadsAsMergeTarget(ctx, key)
	if err != nil {
		return err
	}

	mergeCtx := ctx
	if len(mt.heads) == 0 {
		mergeCtx = coreblock.ContextWithNewDocCreateMode(ctx)
	}

	mp, err := db.newMergeProcessor(mergeCtx, col)
	if err != nil {
		return err
	}

	err = mp.loadComposites(mergeCtx, dagMerge.Cid, mt)
	if err != nil {
		return err
	}

	// mergeComposites may commit and start new transactions for large DAGs,
	// so we need to use the returned context which contains the current transaction
	ctx, err = mp.mergeComposites(mergeCtx)
	if err != nil {
		return err
	}

	if err := mp.syncIndexes(ctx); err != nil {
		return err
	}

	// Get the current transaction (may be different from original if batching occurred)
	txn = datastore.CtxMustGetTxn(ctx)
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
// be called to remove the key from the queue. Returns an error if the context
// is cancelled while waiting.
func (m *mergeQueue) add(ctx context.Context, key string) error {
	for {
		m.mutex.Lock()
		done, ok := m.keys[key]
		if !ok {
			m.keys[key] = make(chan struct{})
			m.mutex.Unlock()
			return nil
		}
		m.mutex.Unlock()
		select {
		case <-done:
		case <-ctx.Done():
			return ctx.Err()
		}
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

	// blocksProcessed tracks how many blocks have been processed in the current transaction.
	blocksProcessed int

	// mergedFields collects field values during merge for new documents.
	mergedFields map[string]map[string][]byte
}

func (db *DB) newMergeProcessor(
	ctx context.Context,
	col *collection,
) (*mergeProcessor, error) {
	txn := datastore.CtxMustGetTxn(ctx)

	blockLS := cidlink.DefaultLinkSystem()
	blockLS.SetReadStorage(blockstore.NewIPLDStore(txn.Blockstore()))

	encBlockLS := cidlink.DefaultLinkSystem()
	encBlockLS.SetReadStorage(blockstore.NewIPLDStore(txn.Encstore()))

	return &mergeProcessor{
		blockLS:    blockLS,
		encBlockLS: encBlockLS,
		col:        col,
		db:         db,
		docIDs:     make(map[client.DocID]*client.Document),
		composites: list.New(),
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
		return err
	}

	block, err := coreblock.GetFromNode(nd)
	if err != nil {
		return err
	}

	// In the simplest case, the new block or its children will link to the current head/heads (merge target)
	// of the composite DAG. However, the new block and its children might have branched off from an older block.
	// In this case, we also need to walk back the merge target's DAG until we reach a common block.
	if block.Delta.GetPriority() >= mt.headHeight {
		mp.composites.PushFront(block)
		for _, head := range block.Heads {
			err := mp.loadComposites(ctx, head.Cid, mt)
			if err != nil {
				return err
			}
		}
	} else {
		newMT := newMergeTarget()
		for _, b := range mt.heads {
			for _, link := range b.Heads {
				nd, err := mp.blockLS.Load(linking.LinkContext{Ctx: ctx}, link, coreblock.BlockSchemaPrototype)
				if err != nil {
					return err
				}

				childBlock, err := coreblock.GetFromNode(nd)
				if err != nil {
					return err
				}

				newMT.heads[link.Cid] = childBlock
				newMT.headHeight = childBlock.Delta.GetPriority()
			}
		}
		return mp.loadComposites(ctx, blockCid, newMT)
	}
	return nil
}

func (mp *mergeProcessor) mergeComposites(ctx context.Context) (context.Context, error) {
	for e := mp.composites.Front(); e != nil; e = e.Next() {
		block := e.Value.(*coreblock.Block)
		link, err := block.GenerateLink()
		if err != nil {
			return ctx, err
		}
		ctx, err = mp.processBlock(ctx, block, link)
		if err != nil {
			return ctx, err
		}
	}

	return ctx, nil
}

func (mp *mergeProcessor) loadEncryptionBlock(
	ctx context.Context,
	encLink cidlink.Link,
) (*coreblock.Encryption, error) {
	nd, err := mp.encBlockLS.Load(linking.LinkContext{Ctx: ctx}, encLink, coreblock.EncryptionSchemaPrototype)
	if err != nil {
		return nil, err
	}

	return coreblock.GetEncryptionBlockFromNode(nd)
}

// processEncryptedBlock decrypts the block if it is encrypted and returns the decrypted block.
// If the block is encrypted and we were not able to decrypt it, it returns false as the second return value
// which indicates that the we can't read the block.
// If we were able to decrypt the block, we return the decrypted block and true as the second return value.
func (mp *mergeProcessor) processEncryptedBlock(
	ctx context.Context,
	dagBlock *coreblock.Block,
) (*coreblock.Block, bool, error) {
	if dagBlock.IsEncrypted() {
		encBlock, err := mp.loadEncryptionBlock(ctx, *dagBlock.Encryption)
		if err != nil {
			return nil, false, err
		}

		if encBlock == nil {
			return dagBlock, false, nil
		}

		plainTextBlock, err := decryptBlock(ctx, dagBlock, encBlock)
		if err != nil {
			return nil, false, err
		}
		if plainTextBlock != nil {
			return plainTextBlock, true, nil
		}
	}
	return dagBlock, true, nil
}

// commitBatchIfNeeded checks if the current transaction has processed enough blocks and commits it.
func (mp *mergeProcessor) commitBatchIfNeeded(ctx context.Context) (context.Context, error) {
	if mp.blocksProcessed < mergeBlockBatchSize {
		return ctx, nil
	}

	txn := datastore.CtxMustGetTxn(ctx)
	err := txn.Commit()
	if err != nil {
		return ctx, err
	}

	// Create a new transaction using a fresh context to avoid reusing the committed transaction.
	// ensureContextTxn would otherwise see the old (committed) transaction and try to reuse it.
	freshCtx := context.Background()
	newCtx, newTxn, err := ensureContextTxn(freshCtx, mp.db, false)
	if err != nil {
		return ctx, err
	}

	mp.blockLS.SetReadStorage(blockstore.NewIPLDStore(newTxn.Blockstore()))
	mp.encBlockLS.SetReadStorage(blockstore.NewIPLDStore(newTxn.Encstore()))

	mp.blocksProcessed = 0

	return newCtx, nil
}

// processBlock merges the block and its children to the datastore and sets the head accordingly.
func (mp *mergeProcessor) processBlock(
	ctx context.Context,
	dagBlock *coreblock.Block,
	blockLink cidlink.Link,
) (context.Context, error) {
	block, canRead, err := mp.processEncryptedBlock(ctx, dagBlock)
	if err != nil {
		return ctx, err
	}

	if canRead {
		crdt, err := mp.initCRDTForType(ctx, dagBlock.Delta)
		if err != nil {
			return ctx, err
		}

		// If the CRDT is nil, it means the field is not part
		// of the schema and we can safely ignore it.
		if crdt == nil {
			return ctx, nil
		}

		err = coreblock.ProcessBlock(ctx, crdt, block, blockLink)
		if err != nil {
			return ctx, err
		}

		if coreblock.IsNewDocCreateMode(ctx) && !block.Delta.IsComposite() && !block.Delta.IsCollection() {
			docID := string(block.Delta.GetDocID())
			fieldName := block.Delta.GetFieldName()
			data := block.Delta.GetData()
			if len(data) > 0 && fieldName != "" {
				if mp.mergedFields == nil {
					mp.mergedFields = make(map[string]map[string][]byte)
				}
				if mp.mergedFields[docID] == nil {
					mp.mergedFields[docID] = make(map[string][]byte)
				}
				mp.mergedFields[docID][fieldName] = data
			}
		}

		mp.blocksProcessed++
		ctx, err = mp.commitBatchIfNeeded(ctx)
		if err != nil {
			return ctx, err
		}
	}

	// For collection blocks, Links point to document composites which are processed in their own merge events.
	if dagBlock.Delta.IsCollection() {
		return ctx, nil
	}

	for _, link := range dagBlock.Links {
		nd, err := mp.blockLS.Load(linking.LinkContext{Ctx: ctx}, link.Link, coreblock.BlockSchemaPrototype)
		if err != nil {
			return ctx, err
		}

		childBlock, err := coreblock.GetFromNode(nd)
		if err != nil {
			return ctx, err
		}

		ctx, err = mp.processBlock(ctx, childBlock, link.Link)
		if err != nil {
			return ctx, err
		}
	}

	return ctx, nil
}

func decryptBlock(
	ctx context.Context,
	block *coreblock.Block,
	encBlock *coreblock.Encryption,
) (*coreblock.Block, error) {
	_, encryptor := encryption.EnsureContextWithEncryptor(ctx)

	if block.Delta.IsComposite() || block.Delta.IsCollection() {
		// for composite blocks there is nothing to decrypt
		return block, nil
	}

	bytes, err := encryptor.Decrypt(block.Delta.GetData(), encBlock.Key)
	if err != nil {
		return nil, err
	}
	if len(bytes) == 0 {
		return nil, nil
	}
	newBlock := block.Clone()
	newBlock.Delta.SetData(bytes)
	return newBlock, nil
}

func (mp *mergeProcessor) initCRDTForType(ctx context.Context, crdtUnion crdt.CRDT) (crdt.ReplicatedData, error) {
	txn := datastore.CtxMustGetTxn(ctx)

	shortID, err := id.GetShortCollectionID(ctx, mp.col.Version().CollectionID)
	if err != nil {
		return nil, err
	}

	switch {
	case crdtUnion.IsComposite():
		docID, err := client.NewDocIDFromString(string(crdtUnion.GetDocID()))
		if err != nil {
			return nil, err
		}
		err = mp.trackMergedDocument(ctx, docID)
		if err != nil {
			return nil, err
		}
		return crdt.NewDocComposite(
			txn.Datastore(),
			mp.col.Version().VersionID,
			keys.DataStoreKey{
				CollectionShortID: shortID,
				DocID:             docID.String(),
			}.WithFieldID(core.COMPOSITE_NAMESPACE),
		), nil

	case crdtUnion.IsCollection():
		return crdt.NewCollection(
			mp.col.Version().VersionID,
			keys.NewHeadstoreColKey(shortID),
		), nil

	default:
		docID, err := client.NewDocIDFromString(string(crdtUnion.GetDocID()))
		if err != nil {
			return nil, err
		}
		err = mp.trackMergedDocument(ctx, docID)
		if err != nil {
			return nil, err
		}

		field := crdtUnion.GetFieldName()
		fd, ok := mp.col.Version().GetFieldByName(field)
		if !ok {
			// If the field is not part of the schema, we can safely ignore it.
			return nil, nil
		}

		fieldShortID, err := id.GetShortFieldID(ctx, shortID, fd.FieldID)
		if err != nil {
			return nil, err
		}

		return crdt.FieldLevelCRDTWithStore(
			txn.Datastore(),
			mp.col.Version().VersionID,
			fd.Typ,
			fd.Kind,
			keys.DataStoreKey{
				CollectionShortID: shortID,
				DocID:             docID.String(),
			}.WithFieldID(fmt.Sprint(fieldShortID)),
			field,
		)
	}
}

// trackMergedDocument tracks the current version of the document so we
// can correctly sync indexes after a merge.
func (mp *mergeProcessor) trackMergedDocument(ctx context.Context, docID client.DocID) error {
	// Skip tracking for collections without indexes - no need to read old doc
	if len(mp.col.indexes) == 0 {
		mp.docIDs[docID] = nil
		return nil
	}
	_, exists := mp.docIDs[docID]
	if exists {
		return nil
	}
	// Skip the old doc read for new documents
	if coreblock.IsNewDocCreateMode(ctx) {
		mp.docIDs[docID] = nil
		return nil
	}
	doc, err := mp.col.Get(ctx, docID, false)
	if err != nil && !errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized) {
		return nil
	}
	mp.docIDs[docID] = doc
	return nil
}

func getCollectionFromCollectionID(ctx context.Context, db *DB, collectionID string) (*collection, error) {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return nil, err
	}
	defer txn.Discard()

	cols, err := db.getCollections(
		ctx,
		client.CollectionFetchOptions{
			CollectionID: immutable.Some(collectionID),
		},
	)
	if err != nil {
		return nil, err
	}
	if len(cols) == 0 {
		return nil, client.NewErrCollectionNotFoundForSchema(collectionID)
	}
	// We currently only support one active collection per root schema
	// so it is safe to return the first one.
	return cols[0].(*collection), nil
}

// getHeadsAsMergeTarget retrieves the heads of the composite DAG for the given document
// and returns them as a merge target.
func getHeadsAsMergeTarget(ctx context.Context, key keys.HeadstoreKey) (mergeTarget, error) {
	cids, err := getHeads(ctx, key)

	if err != nil {
		return mergeTarget{}, err
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
		return nil, err
	}

	block, err := coreblock.GetFromBytes(b.RawData())
	if err != nil {
		return nil, err
	}

	return block, nil
}

// syncIndexes syncs indexes for all documents processed by this merge processor.
func (mp *mergeProcessor) syncIndexes(ctx context.Context) error {
	if len(mp.col.indexes) == 0 {
		return nil
	}

	for docID, oldDoc := range mp.docIDs {
		if oldDoc == nil {
			docIDStr := docID.String()
			if fields, ok := mp.mergedFields[docIDStr]; ok && len(fields) > 0 {
				newDoc, err := buildDocFromMergeFields(ctx, docIDStr, mp.col, fields)
				if err == nil {
					if err := syncDocIndex(ctx, mp.col, nil, newDoc); err != nil {
						return err
					}
					continue
				}
			}
		}

		if err := syncIndexedDoc(ctx, docID, mp.col, oldDoc); err != nil {
			return err
		}
	}
	return nil
}

func syncIndexedDoc(
	ctx context.Context,
	docID client.DocID,
	col *collection,
	oldDoc *client.Document,
) error {
	newDoc, err := col.Get(ctx, docID, false)
	if err != nil && !errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized) {
		return err
	}
	return syncDocIndex(ctx, col, oldDoc, newDoc)
}

// buildDocFromMergeFields constructs a *client.Document from field values collected during merge.
func buildDocFromMergeFields(
	ctx context.Context,
	docIDStr string,
	col *collection,
	fields map[string][]byte,
) (*client.Document, error) {
	docID, err := client.NewDocIDFromString(docIDStr)
	if err != nil {
		return nil, err
	}
	doc, err := client.NewDocWithID(ctx, docID, col.Version())
	if err != nil {
		return nil, err
	}
	for fieldName, rawBytes := range fields {
		fd, ok := col.Version().GetFieldByName(fieldName)
		if !ok {
			continue
		}
		var val any
		if err := cbor.Unmarshal(rawBytes, &val); err != nil {
			continue
		}
		typedVal, err := core.NormalizeFieldValue(fd, val)
		if err != nil {
			continue
		}
		if err := doc.Set(ctx, fieldName, typedVal); err != nil {
			continue
		}
	}
	doc.Clean()
	return doc, nil
}

// syncDocIndex updates the index for a document given its old and new state.
func syncDocIndex(ctx context.Context, col *collection, oldDoc, newDoc *client.Document) error {
	if oldDoc != nil && newDoc != nil {
		return col.updateDocIndex(ctx, oldDoc, newDoc)
	} else if oldDoc == nil && newDoc != nil {
		return col.indexNewDoc(ctx, newDoc)
	} else if oldDoc != nil && newDoc == nil {
		return col.deleteIndexedDoc(ctx, oldDoc)
	}
	return nil
}
