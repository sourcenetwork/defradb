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

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/datastore/badger/v4"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/events"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/db/base"
	"github.com/sourcenetwork/defradb/internal/merkle/clock"
	merklecrdt "github.com/sourcenetwork/defradb/internal/merkle/crdt"
)

func (db *db) handleMerges(ctx context.Context, merges events.Subscription[events.DAGMerge]) {
	for {
		select {
		case <-ctx.Done():
			return
		case merge, ok := <-merges:
			if !ok {
				return
			}
			go func() {
				err := db.executeMerge(ctx, merge)
				if err != nil {
					log.ErrorContextE(
						ctx,
						"Failed to execute merge",
						err,
						corelog.String("CID", merge.Cid.String()),
						corelog.String("Error", err.Error()),
					)
				}
			}()
		}
	}
}

func (db *db) executeMerge(ctx context.Context, dagMerge events.DAGMerge) error {
	defer func() {
		// Notify the caller that the merge is complete.
		if dagMerge.MergeCompleteChan != nil {
			close(dagMerge.MergeCompleteChan)
		}
	}()
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
	ls.SetReadStorage(txn.DAGstore().AsIPLDStorage())

	docID, err := getDocIDFromBlock(ctx, ls, dagMerge.Cid)
	if err != nil {
		return err
	}
	dsKey := base.MakeDataStoreKeyWithCollectionAndDocID(col.Description(), docID.String())

	mp, err := db.newMergeProcessor(txn, ls, col, dsKey)
	if err != nil {
		return err
	}

	mt, err := getHeadsAsMergeTarget(ctx, txn, dsKey)
	if err != nil {
		return err
	}

	err = mp.loadComposites(ctx, dagMerge.Cid, mt)
	if err != nil {
		return err
	}

	for retry := 0; retry < db.MaxTxnRetries(); retry++ {
		err := mp.mergeComposites(ctx)
		if err != nil {
			return err
		}
		err = syncIndexedDoc(ctx, docID, col)
		if err != nil {
			return err
		}
		err = txn.Commit(ctx)
		if err != nil {
			if errors.Is(err, badger.ErrTxnConflict) {
				txn, err = db.NewTxn(ctx, false)
				if err != nil {
					return err
				}
				ctx = SetContextTxn(ctx, txn)
				mp.txn = txn
				mp.ls.SetReadStorage(txn.DAGstore().AsIPLDStorage())
				// Reset the CRDTs to avoid reusing the old transaction.
				mp.mCRDTs = make(map[string]merklecrdt.MerkleCRDT)
				continue
			}
			return err
		}
		break
	}

	return nil
}

type mergeProcessor struct {
	txn        datastore.Txn
	ls         linking.LinkSystem
	mCRDTs     map[string]merklecrdt.MerkleCRDT
	col        *collection
	dsKey      core.DataStoreKey
	composites *list.List
}

func (db *db) newMergeProcessor(
	txn datastore.Txn,
	ls linking.LinkSystem,
	col *collection,
	dsKey core.DataStoreKey,
) (*mergeProcessor, error) {
	return &mergeProcessor{
		txn:        txn,
		ls:         ls,
		mCRDTs:     make(map[string]merklecrdt.MerkleCRDT),
		col:        col,
		dsKey:      dsKey,
		composites: list.New(),
	}, nil
}

type mergeTarget struct {
	heads      map[cid.Cid]*coreblock.Block
	headHeigth uint64
}

func newMergeTarget() mergeTarget {
	return mergeTarget{
		heads: make(map[cid.Cid]*coreblock.Block),
	}
}

// loadComposites retrieves and stores into the merge processor the composite blocks for the given
// document until it reaches a block that has already been merged or until we reach the genesis block.
func (mp *mergeProcessor) loadComposites(
	ctx context.Context,
	blockCid cid.Cid,
	mt mergeTarget,
) error {
	nd, err := mp.ls.Load(linking.LinkContext{Ctx: ctx}, cidlink.Link{Cid: blockCid}, coreblock.SchemaPrototype)
	if err != nil {
		return err
	}

	block, err := coreblock.GetFromNode(nd)
	if err != nil {
		return err
	}

	if _, ok := mt.heads[blockCid]; ok {
		// We've already processed this block.
		return nil
	}

	if block.Delta.GetPriority() >= mt.headHeigth {
		mp.composites.PushFront(block)
		for _, link := range block.Links {
			if link.Name == core.HEAD {
				err := mp.loadComposites(ctx, link.Cid, mt)
				if err != nil {
					return err
				}
			}
		}
	} else {
		newMT := newMergeTarget()
		for _, b := range mt.heads {
			for _, link := range b.Links {
				nd, err := mp.ls.Load(linking.LinkContext{Ctx: ctx}, link.Link, coreblock.SchemaPrototype)
				if err != nil {
					return err
				}

				childBlock, err := coreblock.GetFromNode(nd)
				if err != nil {
					return err
				}

				newMT.heads[link.Cid] = childBlock
				newMT.headHeigth = childBlock.Delta.GetPriority()
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
	return nil
}

// processBlock merges the block and its children to the datastore and sets the head accordingly.
func (mp *mergeProcessor) processBlock(
	ctx context.Context,
	block *coreblock.Block,
	blockLink cidlink.Link,
) error {
	crdt, err := mp.initCRDTForType(block.Delta.GetFieldName())
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

	for _, link := range block.Links {
		if link.Name == core.HEAD {
			continue
		}

		nd, err := mp.ls.Load(linking.LinkContext{Ctx: ctx}, link.Link, coreblock.SchemaPrototype)
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

func (mp *mergeProcessor) initCRDTForType(
	field string,
) (merklecrdt.MerkleCRDT, error) {
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

func getDocIDFromBlock(ctx context.Context, ls linking.LinkSystem, cid cid.Cid) (client.DocID, error) {
	nd, err := ls.Load(linking.LinkContext{Ctx: ctx}, cidlink.Link{Cid: cid}, coreblock.SchemaPrototype)
	if err != nil {
		return client.DocID{}, err
	}
	block, err := coreblock.GetFromNode(nd)
	if err != nil {
		return client.DocID{}, err
	}
	return client.NewDocIDFromString(string(block.Delta.GetDocID()))
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
	headset := clock.NewHeadSet(
		txn.Headstore(),
		dsKey.WithFieldId(core.COMPOSITE_NAMESPACE).ToHeadStoreKey(),
	)

	cids, _, err := headset.List(ctx)
	if err != nil {
		return mergeTarget{}, err
	}

	mt := newMergeTarget()
	for _, cid := range cids {
		b, err := txn.DAGstore().Get(ctx, cid)
		if err != nil {
			return mergeTarget{}, err
		}

		block, err := coreblock.GetFromBytes(b.RawData())
		if err != nil {
			return mergeTarget{}, err
		}

		mt.heads[cid] = block
		// All heads have the same height so overwriting is ok.
		mt.headHeigth = block.Delta.GetPriority()
	}
	return mt, nil
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

	if isDeletedDoc {
		return col.deleteIndexedDoc(ctx, oldDoc)
	} else if isNewDoc {
		return col.indexNewDoc(ctx, doc)
	} else {
		return col.updateDocIndex(ctx, oldDoc, doc)
	}
}
