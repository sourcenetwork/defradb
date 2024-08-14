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
	"bytes"
	"container/list"
	"context"
	"sync"

	"github.com/ipfs/go-cid"
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

	ls := cidlink.DefaultLinkSystem()
	ls.SetReadStorage(txn.Blockstore().AsIPLDStorage())

	docID, err := client.NewDocIDFromString(dagMerge.DocID)
	if err != nil {
		return err
	}
	dsKey := base.MakeDataStoreKeyWithCollectionAndDocID(col.Description(), docID.String())

	mp, err := db.newMergeProcessor(txn, ls, col, dsKey)
	if err != nil {
		return err
	}

	mt, err := getHeadsAsMergeTarget(ctx, txn, dsKey.WithFieldId(core.COMPOSITE_NAMESPACE))
	if err != nil {
		return err
	}

	err = mp.loadBlocks(ctx, dagMerge.Cid, mt, false)
	if err != nil {
		return err
	}

	err = mp.mergeBlocks(ctx, false)
	if err != nil {
		return err
	}

	mp.sendPendingEncryptionRequest()

	if !mp.hasPendingCompositeBlock {
		err = syncIndexedDoc(ctx, docID, col)
		if err != nil {
			return err
		}
	}

	err = txn.Commit(ctx)
	if err != nil {
		return err
	}

	// send a complete event so we can track merges in the integration tests
	db.events.Publish(event.NewMessage(event.MergeCompleteName, dagMerge))
	return nil
}

type encryptionMergeGroup struct {
	compositeKey core.EncStoreDocKey
	fieldsKeys   []core.EncStoreDocKey
}

func createMergeGroups(keyEvent encryption.KeyRetrievedEvent) map[string]encryptionMergeGroup {
	mergeGroups := make(map[string]encryptionMergeGroup)

	for encStoreKey := range keyEvent.Keys {
		g := mergeGroups[encStoreKey.DocID]

		if encStoreKey.FieldName.HasValue() {
			g.fieldsKeys = append(g.fieldsKeys, encStoreKey)
		} else {
			g.compositeKey = encStoreKey
		}

		mergeGroups[encStoreKey.DocID] = g
	}

	return mergeGroups
}

func (db *db) mergeEncryptedBlocks(ctx context.Context, keyEvent encryption.KeyRetrievedEvent) error {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	col, err := getCollectionFromRootSchema(ctx, db, keyEvent.SchemaRoot)
	if err != nil {
		return err
	}

	ls := cidlink.DefaultLinkSystem()
	ls.SetReadStorage(txn.Blockstore().AsIPLDStorage())

	for docID, mergeGroup := range createMergeGroups(keyEvent) {
		docID, err := client.NewDocIDFromString(docID)
		if err != nil {
			return err
		}
		dsKey := base.MakeDataStoreKeyWithCollectionAndDocID(col.Description(), docID.String())

		mp, err := db.newMergeProcessor(txn, ls, col, dsKey)
		if err != nil {
			return err
		}

		var blocks []*coreblock.Block
		// if merge ground includes composite, we process only the composite block and skip the field
		// blocks, as they will be processed as part of the composite block anyway.
		// Otherwise, we load the blocks for each field
		if mergeGroup.compositeKey.DocID != "" {
			cids, err := getHeads(ctx, txn, dsKey.WithFieldId(core.COMPOSITE_NAMESPACE))
			if err != nil {
				return err
			}

			blocks, err = loadBlocksWithKeyIDFromBlockstore(ctx, txn, cids, mergeGroup.compositeKey.KeyID)
			if err != nil {
				return err
			}
		} else {
			for _, fieldStoreKey := range mergeGroup.fieldsKeys {
				fd, ok := mp.col.Definition().GetFieldByName(fieldStoreKey.FieldName.Value())
				if !ok {
					return client.NewErrFieldNotExist(fieldStoreKey.FieldName.Value())
				}

				fieldDsKey := dsKey.WithFieldId(fd.ID.String())

				cids, err := getHeads(ctx, txn, fieldDsKey)
				if err != nil {
					return err
				}

				fieldBlocks, err := loadBlocksWithKeyIDFromBlockstore(ctx, txn, cids, fieldStoreKey.KeyID)
				if err != nil {
					return err
				}

				blocks = append(blocks, fieldBlocks...)
			}
		}

		for _, block := range blocks {
			mp.blocks.PushFront(block)
		}
		err = mp.mergeBlocks(ctx, true)
		if err != nil {
			return err
		}

		// TODO: test is doc field was indexed after decryption
		err = syncIndexedDoc(ctx, docID, col)
		if err != nil {
			return err
		}
	}

	err = txn.Commit(ctx)
	if err != nil {
		return err
	}

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
	txn                          datastore.Txn
	lsys                         linking.LinkSystem
	mCRDTs                       map[string]merklecrdt.MerkleCRDT
	col                          *collection
	dsKey                        core.DataStoreKey
	blocks                       *list.List
	pendingEncryptionKeyRequests map[core.EncStoreDocKey]struct{}
	hasPendingCompositeBlock     bool
}

func (db *db) newMergeProcessor(
	txn datastore.Txn,
	lsys linking.LinkSystem,
	col *collection,
	dsKey core.DataStoreKey,
) (*mergeProcessor, error) {
	return &mergeProcessor{
		txn:                          txn,
		lsys:                         lsys,
		mCRDTs:                       make(map[string]merklecrdt.MerkleCRDT),
		col:                          col,
		dsKey:                        dsKey,
		blocks:                       list.New(),
		pendingEncryptionKeyRequests: make(map[core.EncStoreDocKey]struct{}),
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

// loadBlocks retrieves and stores into the merge processor the blocks for the given
// CID until it reaches a block that has already been merged or until we reach the genesis block.
func (mp *mergeProcessor) loadBlocks(
	ctx context.Context,
	blockCid cid.Cid,
	mt mergeTarget,
	willDecrypt bool,
) error {
	if b, ok := mt.heads[blockCid]; ok {
		// the head is already known, but the block might be encrypted
		// if this time we try to decrypt it, we load the block
		if !b.IsEncrypted() || !willDecrypt {
			// We've already processed this block.
			return nil
		}
	}

	nd, err := mp.lsys.Load(linking.LinkContext{Ctx: ctx}, cidlink.Link{Cid: blockCid}, coreblock.SchemaPrototype)
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
		mp.blocks.PushFront(block)
		for _, prevCid := range block.GetPrevBlockCids() {
			err := mp.loadBlocks(ctx, prevCid, mt, willDecrypt)
			if err != nil {
				return err
			}
		}
	} else {
		newMT := newMergeTarget()
		for _, b := range mt.heads {
			for _, link := range b.Links {
				if link.Name == core.HEAD {
					nd, err := mp.lsys.Load(linking.LinkContext{Ctx: ctx}, link.Link, coreblock.SchemaPrototype)
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
		}
		return mp.loadBlocks(ctx, blockCid, newMT, willDecrypt)
	}
	return nil
}

func (mp *mergeProcessor) mergeBlocks(ctx context.Context, withDecryption bool) error {
	for e := mp.blocks.Front(); e != nil; e = e.Next() {
		block := e.Value.(*coreblock.Block)
		link, err := block.GenerateLink()
		if err != nil {
			return err
		}
		err = mp.processBlock(ctx, block, link, withDecryption)
		if err != nil {
			return err
		}
	}
	return nil
}

// processEncryptedBlock decrypts the block if it is encrypted and returns the decrypted block.
// If the block is encrypted and we were not able to decrypt it, it returns true as the second return value
// which indicates that the we should skip merging of the block.
// If we were able to decrypt the block, we return the decrypted block and false as the second return value.
func (mp *mergeProcessor) processEncryptedBlock(
	ctx context.Context,
	dagBlock *coreblock.Block,
	withDecryption bool,
) (*coreblock.Block, bool, error) {
	if dagBlock.IsEncrypted() {
		plainTextBlock, err := decryptBlock(ctx, dagBlock)
		if err != nil {
			return nil, false, err
		}
		if plainTextBlock != nil {
			return plainTextBlock, false, nil
		} else {
			blockEnc := dagBlock.Encryption
			// we weren't able to decrypt the block, so we request the encryption key unless it's a decryption pass
			if !withDecryption && (dagBlock.Delta.IsComposite() && blockEnc.Type == coreblock.DocumentEncrypted) ||
				blockEnc.Type == coreblock.FieldEncrypted {
				docID := string(dagBlock.Delta.GetDocID())
				fieldName := immutable.None[string]()
				if blockEnc.Type == coreblock.FieldEncrypted {
					fieldName = immutable.Some(dagBlock.Delta.GetFieldName())
				}
				mp.addPendingEncryptionRequest(docID, fieldName, string(blockEnc.KeyID))
			}
			return dagBlock, true, nil
		}
	}
	return dagBlock, false, nil
}

func (mp *mergeProcessor) addPendingEncryptionRequest(docID string, fieldName immutable.Option[string], keyID string) {
	mp.pendingEncryptionKeyRequests[core.NewEncStoreDocKey(docID, fieldName, keyID)] = struct{}{}
	if !fieldName.HasValue() {
		mp.hasPendingCompositeBlock = true
	}
}

func (mp *mergeProcessor) sendPendingEncryptionRequest() {
	n := len(mp.pendingEncryptionKeyRequests)
	if n == 0 {
		return
	}
	schemaRoot := mp.col.SchemaRoot()
	storeKeys := make([]core.EncStoreDocKey, 0, n)
	for k := range mp.pendingEncryptionKeyRequests {
		storeKeys = append(storeKeys, k)
	}
	mp.col.db.events.Publish(encryption.NewRequestKeysMessage(schemaRoot, storeKeys))
}

// processBlock merges the block and its children to the datastore and sets the head accordingly.
// withDecryption is a flag that indicates this is the decryption pass instead of a normal merge.
func (mp *mergeProcessor) processBlock(
	ctx context.Context,
	dagBlock *coreblock.Block,
	blockLink cidlink.Link,
	withDecryption bool,
) error {
	block, skipMerge, err := mp.processEncryptedBlock(ctx, dagBlock, withDecryption)
	if err != nil {
		return err
	}

	crdt, err := mp.initCRDTForType(dagBlock.Delta.GetFieldName())
	if err != nil {
		return err
	}

	// If the CRDT is nil, it means the field is not part
	// of the schema and we can safely ignore it.
	if crdt == nil {
		return nil
	}

	err = crdt.Clock().ProcessBlock(ctx, block, blockLink, skipMerge, withDecryption)
	if err != nil {
		return err
	}

	for _, link := range dagBlock.Links {
		if link.Name == core.HEAD {
			continue
		}

		nd, err := mp.lsys.Load(linking.LinkContext{Ctx: ctx}, link.Link, coreblock.SchemaPrototype)
		if err != nil {
			return err
		}

		childBlock, err := coreblock.GetFromNode(nd)
		if err != nil {
			return err
		}

		if err := mp.processBlock(ctx, childBlock, link.Link, withDecryption); err != nil {
			return err
		}
	}

	return nil
}

func decryptBlock(ctx context.Context, block *coreblock.Block) (*coreblock.Block, error) {
	optFieldName := immutable.None[string]()
	blockEnc := block.Encryption
	if blockEnc.Type == coreblock.FieldEncrypted {
		optFieldName = immutable.Some(block.Delta.GetFieldName())
	}

	encStoreKey := core.NewEncStoreDocKey(string(block.Delta.GetDocID()), optFieldName, string(blockEnc.KeyID))

	if block.Delta.IsComposite() {
		// for composite blocks there is nothing to decrypt
		// so we just check if we have the encryption key for child blocks
		bytes, err := encryption.GetKey(ctx, encStoreKey)
		if err != nil {
			return nil, err
		}
		if len(bytes) == 0 {
			return nil, nil
		}
		return block, nil
	}

	clonedCRDT := block.Delta.Clone()
	bytes, err := encryption.DecryptDoc(ctx, encStoreKey, clonedCRDT.GetData())
	if err != nil {
		return nil, err
	}
	if len(bytes) == 0 {
		return nil, nil
	}
	clonedCRDT.SetData(bytes)
	return &coreblock.Block{Delta: clonedCRDT, Links: block.Links}, nil
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
			mp.dsKey.WithFieldId(core.COMPOSITE_NAMESPACE),
			"",
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
		mp.dsKey.WithFieldId(fd.ID.String()),
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

// loadBlocksWithKeyIDFromBlockstore loads the blocks from the blockstore that have given encryption
// keyID until it reaches a block with a different keyID or without any.
// The returned blocks are ordered from the newest to the oldest.
func loadBlocksWithKeyIDFromBlockstore(
	ctx context.Context,
	txn datastore.Txn,
	cids []cid.Cid,
	keyID string,
) ([]*coreblock.Block, error) {
	var blocks []*coreblock.Block
	for len(cids) > 0 {
		cid := cids[0]
		block, err := loadBlockFromBlockStore(ctx, txn, cid)
		if err != nil {
			return nil, err
		}

		if block.Encryption != nil && bytes.Equal(block.Encryption.KeyID, []byte(keyID)) {
			blocks = append(blocks, block)
			cids = append(cids, block.GetPrevBlockCids()...)
		}
		cids = cids[1:]
	}
	return blocks, nil
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

	// this can happen we received an encrypted document that we haven't decrypted yet
	if isNewDoc && isDeletedDoc {
		return nil
	}

	if isNewDoc {
		return col.indexNewDoc(ctx, doc)
	} else if isDeletedDoc {
		return col.deleteIndexedDoc(ctx, oldDoc)
	} else {
		return col.updateDocIndex(ctx, oldDoc, doc)
	}
}
