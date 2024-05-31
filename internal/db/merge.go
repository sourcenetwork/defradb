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
						corelog.String("cid", merge.Cid.String()),
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
	mp, err := db.newMergeProcessor(ctx, dagMerge.Cid, dagMerge.SchemaRoot)
	if err != nil {
		return err
	}
	mt, err := mp.getHeads(ctx)
	if err != nil {
		return err
	}
	err = mp.getComposites(ctx, dagMerge.Cid, mt)
	if err != nil {
		return err
	}
	err = mp.merge(ctx)
	if err != nil {
		return err
	}
	err = mp.syncIndexedDocs(ctx)
	if err != nil {
		return err
	}
	return txn.Commit(ctx)
}

type mergeProcessor struct {
	ctx              context.Context
	txn              datastore.Txn
	ls               linking.LinkSystem
	docID            client.DocID
	mCRDTs           map[uint32]merklecrdt.MerkleCRDT
	col              *collection
	schemaVersionKey core.CollectionSchemaVersionKey
	dsKey            core.DataStoreKey
	composites       *list.List
}

func (db *db) newMergeProcessor(ctx context.Context, cid cid.Cid, rootSchema string) (*mergeProcessor, error) {
	txn, ok := TryGetContextTxn(ctx)
	if !ok {
		return nil, ErrNoTransactionInContext
	}

	ls := cidlink.DefaultLinkSystem()
	ls.SetReadStorage(txn.DAGstore().AsIPLDStorage())
	nd, err := ls.Load(linking.LinkContext{Ctx: ctx}, cidlink.Link{Cid: cid}, coreblock.SchemaPrototype)
	if err != nil {
		return nil, err
	}

	block, err := coreblock.GetFromNode(nd)
	if err != nil {
		return nil, err
	}

	cols, err := db.getCollections(
		ctx,
		client.CollectionFetchOptions{
			SchemaRoot: immutable.Some(rootSchema),
		},
	)
	if err != nil {
		return nil, err
	}

	col := cols[0].(*collection)
	docID, err := client.NewDocIDFromString(string(block.Delta.GetDocID()))
	if err != nil {
		return nil, err
	}

	return &mergeProcessor{
		ctx:    ctx,
		txn:    txn,
		ls:     ls,
		docID:  docID,
		mCRDTs: make(map[uint32]merklecrdt.MerkleCRDT),
		col:    col,
		schemaVersionKey: core.CollectionSchemaVersionKey{
			SchemaVersionID: col.Schema().VersionID,
			CollectionID:    col.ID(),
		},
		dsKey:      base.MakeDataStoreKeyWithCollectionAndDocID(col.Description(), docID.String()),
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

// getComposites retrieves the composite blocks for the given document until it reaches a
// block that has already been merged or until we reach the genesis block.
func (mp *mergeProcessor) getComposites(ctx context.Context, blockCid cid.Cid, mt mergeTarget) error {
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
				err := mp.getComposites(ctx, link.Cid, mt)
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
		return mp.getComposites(ctx, blockCid, newMT)
	}
	return nil
}

// getHeads retrieves the heads of the composite DAG for the given document.
func (mp *mergeProcessor) getHeads(ctx context.Context) (mergeTarget, error) {
	headset := clock.NewHeadSet(
		mp.txn.Headstore(),
		mp.dsKey.WithFieldId(core.COMPOSITE_NAMESPACE).ToHeadStoreKey(),
	)

	cids, _, err := headset.List(ctx)
	if err != nil {
		return mergeTarget{}, err
	}

	mt := newMergeTarget()
	for _, cid := range cids {
		b, err := mp.txn.DAGstore().Get(ctx, cid)
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

func (mp *mergeProcessor) merge(ctx context.Context) error {
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

		b, err := mp.txn.DAGstore().Get(ctx, link.Cid)
		if err != nil {
			return err
		}

		childBlock, err := coreblock.GetFromBytes(b.RawData())
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
	if field == "" {
		return merklecrdt.NewMerkleCompositeDAG(
			mp.txn,
			mp.schemaVersionKey,
			mp.dsKey.WithFieldId(core.COMPOSITE_NAMESPACE),
			"",
		), nil
	}

	fd, ok := mp.col.Definition().GetFieldByName(field)
	if !ok {
		// If the field is not part of the schema, we can safely ignore it.
		return nil, nil
	}

	return merklecrdt.InstanceWithStore(
		mp.txn,
		mp.schemaVersionKey,
		fd.Typ,
		fd.Kind,
		mp.dsKey.WithFieldId(fd.ID.String()),
		field,
	)
}

func (mp *mergeProcessor) syncIndexedDocs(
	ctx context.Context,
) error {
	// remove transaction from old context
	oldCtx := SetContextTxn(ctx, nil)

	oldDoc, err := mp.col.Get(oldCtx, mp.docID, false)
	isNewDoc := errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized)
	if !isNewDoc && err != nil {
		return err
	}

	doc, err := mp.col.Get(ctx, mp.docID, false)
	isDeletedDoc := errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized)
	if !isDeletedDoc && err != nil {
		return err
	}

	if isDeletedDoc {
		return mp.col.deleteIndexedDoc(ctx, oldDoc)
	} else if isNewDoc {
		return mp.col.indexNewDoc(ctx, doc)
	} else {
		return mp.col.updateDocIndex(ctx, oldDoc, doc)
	}
}
