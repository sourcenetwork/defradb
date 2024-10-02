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
	"sync"

	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/db/base"
	"github.com/sourcenetwork/defradb/internal/encryption"
	"github.com/sourcenetwork/defradb/internal/merkle/clock"
	merklecrdt "github.com/sourcenetwork/defradb/internal/merkle/crdt"
)

func (db *db) executeMerge(ctx context.Context, dagMerge event.Merge) error {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	col, err := getCollectionFromRootSchema(ctx, db, dagMerge.SchemaRoot)
	if err != nil {
		return err
	}

	docID, err := client.NewDocIDFromString(dagMerge.DocID)
	if err != nil {
		return err
	}
	dsKey := base.MakeDataStoreKeyWithCollectionAndDocID(col.Description(), docID.String())

	mp, err := db.newMergeProcessor(txn, col, dsKey)
	if err != nil {
		return err
	}

	mt, err := getHeadsAsMergeTarget(ctx, txn, dsKey.WithFieldID(core.COMPOSITE_NAMESPACE))
	if err != nil {
		return err
	}

	err = mp.loadComposites(ctx, dagMerge.Cid, mt)
	if err != nil {
		return err
	}

	err = mp.mergeComposites(ctx)
	if err != nil {
		return err
	}

	err = syncIndexedDoc(ctx, docID, col)
	if err != nil {
		return err
	}

	err = txn.Commit(ctx)
	if err != nil {
		return err
	}

	// send a complete event so we can track merges in the integration tests
	db.events.Publish(event.NewMessage(event.MergeCompleteName, dagMerge))
	return nil
}

// mergeQueue is synchronization source to ensure that concurrent
// document merges do not cause transaction conflicts.
type mergeQueue struct {
	docs  map[string]chan struct{}
	mutex sync.Mutex
}

func newMergeQueue() *mergeQueue {
	return &mergeQueue{
		docs: make(map[string]chan struct{}),
	}
}

// add adds a docID to the queue. If the docID is already in the queue, it will
// wait for the docID to be removed from the queue. For every add call, done must
// be called to remove the docID from the queue. Otherwise, subsequent add calls will
// block forever.
func (m *mergeQueue) add(docID string) {
	m.mutex.Lock()
	done, ok := m.docs[docID]
	if !ok {
		m.docs[docID] = make(chan struct{})
	}
	m.mutex.Unlock()
	if ok {
		<-done
		m.add(docID)
	}
}

func (m *mergeQueue) done(docID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	done, ok := m.docs[docID]
	if ok {
		delete(m.docs, docID)
		close(done)
	}
}

type mergeProcessor struct {
	txn        datastore.Txn
	blockLS    linking.LinkSystem
	encBlockLS linking.LinkSystem
	mCRDTs     map[string]merklecrdt.MerkleCRDT
	col        *collection
	dsKey      core.DataStoreKey
	// composites is a list of composites that need to be merged.
	composites *list.List
	// missingEncryptionBlocks is a list of blocks that we failed to fetch
	missingEncryptionBlocks map[cidlink.Link]struct{}
	// availableEncryptionBlocks is a list of blocks that we have successfully fetched
	availableEncryptionBlocks map[cidlink.Link]*coreblock.Encryption
}

func (db *db) newMergeProcessor(
	txn datastore.Txn,
	col *collection,
	dsKey core.DataStoreKey,
) (*mergeProcessor, error) {
	blockLS := cidlink.DefaultLinkSystem()
	blockLS.SetReadStorage(txn.Blockstore().AsIPLDStorage())

	encBlockLS := cidlink.DefaultLinkSystem()
	encBlockLS.SetReadStorage(txn.Encstore().AsIPLDStorage())

	return &mergeProcessor{
		txn:                       txn,
		blockLS:                   blockLS,
		encBlockLS:                encBlockLS,
		mCRDTs:                    make(map[string]merklecrdt.MerkleCRDT),
		col:                       col,
		dsKey:                     dsKey,
		composites:                list.New(),
		missingEncryptionBlocks:   make(map[cidlink.Link]struct{}),
		availableEncryptionBlocks: make(map[cidlink.Link]*coreblock.Encryption),
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

	nd, err := mp.blockLS.Load(linking.LinkContext{Ctx: ctx}, cidlink.Link{Cid: blockCid}, coreblock.SchemaPrototype)
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
				nd, err := mp.blockLS.Load(linking.LinkContext{Ctx: ctx}, link, coreblock.SchemaPrototype)
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

func (mp *mergeProcessor) mergeComposites(ctx context.Context) error {
	for e := mp.composites.Front(); e != nil; e = e.Next() {
		block := e.Value.(*coreblock.Block)
		link, err := block.GenerateLink()
		if err != nil {
			return err
		}
		err = mp.processBlock(ctx, block, link)
		if err != nil {
			return err
		}
	}

	return mp.tryFetchMissingBlocksAndMerge(ctx)
}

func (mp *mergeProcessor) tryFetchMissingBlocksAndMerge(ctx context.Context) error {
	for len(mp.missingEncryptionBlocks) > 0 {
		links := make([]cidlink.Link, 0, len(mp.missingEncryptionBlocks))
		for link := range mp.missingEncryptionBlocks {
			links = append(links, link)
		}
		msg, results := encryption.NewRequestKeysMessage(links)
		mp.col.db.events.Publish(msg)

		res := <-results.Get()
		if res.Error != nil {
			return res.Error
		}

		clear(mp.missingEncryptionBlocks)

		for i := range res.Items {
			_, link, err := cid.CidFromBytes(res.Items[i].Link)
			if err != nil {
				return err
			}
			var encBlock coreblock.Encryption
			err = encBlock.Unmarshal(res.Items[i].Block)
			if err != nil {
				return err
			}

			mp.availableEncryptionBlocks[cidlink.Link{Cid: link}] = &encBlock
		}

		err := mp.mergeComposites(ctx)
		if err != nil {
			return err
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
		if errors.Is(err, ipld.ErrNotFound{}) {
			mp.missingEncryptionBlocks[encLink] = struct{}{}
			return nil, nil
		}
		return nil, err
	}

	return coreblock.GetEncryptionBlockFromNode(nd)
}

func (mp *mergeProcessor) tryGetEncryptionBlock(
	ctx context.Context,
	encLink cidlink.Link,
) (*coreblock.Encryption, error) {
	if encBlock, ok := mp.availableEncryptionBlocks[encLink]; ok {
		return encBlock, nil
	}
	if _, ok := mp.missingEncryptionBlocks[encLink]; ok {
		return nil, nil
	}

	encBlock, err := mp.loadEncryptionBlock(ctx, encLink)
	if err != nil {
		return nil, err
	}

	if encBlock != nil {
		mp.availableEncryptionBlocks[encLink] = encBlock
	}

	return encBlock, nil
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
		encBlock, err := mp.tryGetEncryptionBlock(ctx, *dagBlock.Encryption)
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
		return err
	}

	if canRead {
		crdt, err := mp.initCRDTForType(dagBlock.Delta.GetFieldName())
		if err != nil {
			return err
		}

		// If the CRDT is nil, it means the field is not part
		// of the schema and we can safely ignore it.
		if crdt == nil {
			return nil
		}

		err = crdt.Clock().ProcessBlock(ctx, block, blockLink)
		if err != nil {
			return err
		}
	}

	for _, link := range dagBlock.Links {
		nd, err := mp.blockLS.Load(linking.LinkContext{Ctx: ctx}, link.Link, coreblock.SchemaPrototype)
		if err != nil {
			return err
		}

		childBlock, err := coreblock.GetFromNode(nd)
		if err != nil {
			return err
		}

		if err := mp.processBlock(ctx, childBlock, link.Link); err != nil {
			return err
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

	if block.Delta.IsComposite() {
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

func (mp *mergeProcessor) initCRDTForType(field string) (merklecrdt.MerkleCRDT, error) {
	mcrdt, exists := mp.mCRDTs[field]
	if exists {
		return mcrdt, nil
	}

	schemaVersionKey := core.CollectionSchemaVersionKey{
		SchemaVersionID: mp.col.Schema().VersionID,
		CollectionID:    mp.col.ID(),
	}

	if field == "" {
		mcrdt = merklecrdt.NewMerkleCompositeDAG(
			mp.txn,
			schemaVersionKey,
			mp.dsKey.WithFieldID(core.COMPOSITE_NAMESPACE),
		)
		mp.mCRDTs[field] = mcrdt
		return mcrdt, nil
	}

	fd, ok := mp.col.Definition().GetFieldByName(field)
	if !ok {
		// If the field is not part of the schema, we can safely ignore it.
		return nil, nil
	}

	mcrdt, err := merklecrdt.InstanceWithStore(
		mp.txn,
		schemaVersionKey,
		fd.Typ,
		fd.Kind,
		mp.dsKey.WithFieldID(fd.ID.String()),
		field,
	)
	if err != nil {
		return nil, err
	}

	mp.mCRDTs[field] = mcrdt
	return mcrdt, nil
}

func getCollectionFromRootSchema(ctx context.Context, db *db, rootSchema string) (*collection, error) {
	cols, err := db.getCollections(
		ctx,
		client.CollectionFetchOptions{
			SchemaRoot: immutable.Some(rootSchema),
		},
	)
	if err != nil {
		return nil, err
	}
	if len(cols) == 0 {
		return nil, client.NewErrCollectionNotFoundForSchema(rootSchema)
	}
	// We currently only support one active collection per root schema
	// so it is safe to return the first one.
	return cols[0].(*collection), nil
}

// getHeadsAsMergeTarget retrieves the heads of the composite DAG for the given document
// and returns them as a merge target.
func getHeadsAsMergeTarget(ctx context.Context, txn datastore.Txn, dsKey core.DataStoreKey) (mergeTarget, error) {
	cids, err := getHeads(ctx, txn, dsKey)

	if err != nil {
		return mergeTarget{}, err
	}

	mt := newMergeTarget()
	for _, cid := range cids {
		block, err := loadBlockFromBlockStore(ctx, txn, cid)
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
func getHeads(ctx context.Context, txn datastore.Txn, dsKey core.DataStoreKey) ([]cid.Cid, error) {
	headset := clock.NewHeadSet(txn.Headstore(), dsKey.ToHeadStoreKey())

	cids, _, err := headset.List(ctx)
	if err != nil {
		return nil, err
	}

	return cids, nil
}

// loadBlockFromBlockStore loads a block from the blockstore.
func loadBlockFromBlockStore(ctx context.Context, txn datastore.Txn, cid cid.Cid) (*coreblock.Block, error) {
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

func syncIndexedDoc(
	ctx context.Context,
	docID client.DocID,
	col *collection,
) error {
	// remove transaction from old context
	oldCtx := SetContextTxn(ctx, nil)

	oldDoc, err := col.Get(oldCtx, docID, false)
	isNewDoc := errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized)
	if !isNewDoc && err != nil {
		return err
	}

	doc, err := col.Get(ctx, docID, false)
	isDeletedDoc := errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized)
	if !isDeletedDoc && err != nil {
		return err
	}

	if isNewDoc {
		return col.indexNewDoc(ctx, doc)
	} else if isDeletedDoc {
		return col.deleteIndexedDoc(ctx, oldDoc)
	} else {
		return col.updateDocIndex(ctx, oldDoc, doc)
	}
}
