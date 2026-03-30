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
	"github.com/sourcenetwork/defradb/internal/encryption"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/utils"
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
		db.colMergeQueue.add(evt.CollectionID)
		defer db.colMergeQueue.done(evt.CollectionID)
	} else {
		// ensure only one merge per docID
		db.docMergeQueue.add(evt.DocID)
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

func (db *DB) executeMerge(ctx context.Context, col *collection, dagMerge event.Merge) error {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return NewErrCreateMergeTxn(err, dagMerge.DocID, dagMerge.Cid.String())
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
			return NewErrGetShortIDForMerge(err, col.Version().CollectionID)
		}

		key = keys.NewHeadstoreColKey(shortID)
	}

	mt, err := getHeadsAsMergeTarget(ctx, key)
	if err != nil {
		return NewErrGetMergeTargetHeads(err, dagMerge.DocID, string(key.Bytes()))
	}

	mp, err := db.newMergeProcessor(ctx, col)
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

	// docIDs contains all docIDs and their original values
	// that have been merged so far by the mergeProcessor
	// the original values are used to update indexes
	docIDs map[client.DocID]*client.Document

	// composites is a list of composites that need to be merged.
	composites *list.List
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

func (mp *mergeProcessor) loadEncryptionBlock(
	ctx context.Context,
	encLink cidlink.Link,
) (*coreblock.Encryption, error) {
	nd, err := mp.encBlockLS.Load(linking.LinkContext{Ctx: ctx}, encLink, coreblock.EncryptionSchemaPrototype)
	if err != nil {
		return nil, NewErrLoadEncryptionBlock(err, encLink.String())
	}

	enc, err := coreblock.GetEncryptionBlockFromNode(nd)
	if err != nil {
		return nil, NewErrLoadEncryptionBlock(err, encLink.String())
	}
	return enc, nil
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

// processBlock merges the block and its children to the datastore and sets the head accordingly.
func (mp *mergeProcessor) processBlock(
	ctx context.Context,
	dagBlock *coreblock.Block,
	blockLink cidlink.Link,
) error {
	block, canRead, err := mp.processEncryptedBlock(ctx, dagBlock)
	if err != nil {
		return NewErrProcessEncryptedBlock(err, blockLink.String())
	}

	if canRead {
		crdt, err := mp.initCRDTForType(ctx, dagBlock.Delta)
		if err != nil {
			return NewErrInitCRDTForMerge(err, blockLink.String())
		}

		// If the CRDT is nil, it means the field is not part
		// of the collection definition and we can safely ignore it.
		if crdt == nil {
			return nil
		}

		err = coreblock.ProcessBlock(ctx, crdt, block, blockLink)
		if err != nil {
			return NewErrProcessCRDTBlock(err, blockLink.String())
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
		return nil, NewErrGetShortIDForMerge(err, mp.col.Version().CollectionID)
	}

	switch {
	case crdtUnion.IsComposite():
		docID, err := client.NewDocIDFromString(string(crdtUnion.GetDocID()))
		if err != nil {
			return nil, NewErrParseDocIDMerge(err, string(crdtUnion.GetDocID()))
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
			return nil, NewErrParseDocIDMerge(err, string(crdtUnion.GetDocID()))
		}
		err = mp.trackMergedDocument(ctx, docID)
		if err != nil {
			return nil, err
		}

		field := crdtUnion.GetFieldName()
		fd, ok := mp.col.Version().GetFieldByName(field)
		if !ok {
			// If the field is not part of the collection definition, we can safely ignore it.
			return nil, nil
		}

		fieldShortID, err := id.GetShortFieldID(ctx, shortID, fd.FieldID)
		if err != nil {
			return nil, NewErrGetShortFieldIDMerge(err, fd.FieldID, field)
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
	_, exists := mp.docIDs[docID]
	if exists {
		return nil
	}
	doc, err := mp.col.GetDocument(ctx, docID)
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
	newDoc, err := col.GetDocument(ctx, docID)
	if err != nil && !errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized) {
		return err
	}
	// Both can be nil during concurrent P2P operations (e.g. delete + update)
	// where the document was already deleted and no prior indexed state exists.
	if oldDoc == nil && newDoc == nil {
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
