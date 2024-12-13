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

	"github.com/ipfs/go-cid"

	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/merkle/clock"
)

// DocHeadBlocksIterator is an iterator that iterates over the head blocks of a document.
type DocHeadBlocksIterator struct {
	ctx        context.Context
	blockstore datastore.Blockstore
	cids       []cid.Cid

	currentCid      cid.Cid
	currentBlock    *coreblock.Block
	currentRawBlock []byte
}

// NewHeadBlocksIterator creates a new DocHeadBlocksIterator.
func NewHeadBlocksIterator(
	ctx context.Context,
	headstore datastore.DSReaderWriter,
	blockstore datastore.Blockstore,
	docID string,
) (*DocHeadBlocksIterator, error) {
	headStoreKey := keys.HeadstoreDocKey{
		DocID:   docID,
		FieldID: core.COMPOSITE_NAMESPACE,
	}
	headset := clock.NewHeadSet(headstore, headStoreKey)
	cids, _, err := headset.List(ctx)
	if err != nil {
		return nil, err
	}
	return &DocHeadBlocksIterator{
		ctx:        ctx,
		blockstore: blockstore,
		cids:       cids,
	}, nil
}

// NewHeadBlocksIteratorFromTxn creates a new DocHeadBlocksIterator from a transaction.
func NewHeadBlocksIteratorFromTxn(
	ctx context.Context,
	txn datastore.Txn,
	docID string,
) (*DocHeadBlocksIterator, error) {
	return NewHeadBlocksIterator(ctx, txn.Headstore(), txn.Blockstore(), docID)
}

// Next advances the iterator to the next block.
func (h *DocHeadBlocksIterator) Next() (bool, error) {
	if len(h.cids) == 0 {
		return false, nil
	}
	nextCid := h.cids[0]
	h.cids = h.cids[1:]

	rawBlock, err := h.blockstore.Get(h.ctx, nextCid)
	if err != nil {
		return false, err
	}
	blk, err := coreblock.GetFromBytes(rawBlock.RawData())
	if err != nil {
		return false, err
	}

	h.currentCid = nextCid
	h.currentBlock = blk
	h.currentRawBlock = rawBlock.RawData()
	return true, nil
}

// CurrentCid returns the CID of the current block.
func (h *DocHeadBlocksIterator) CurrentCid() cid.Cid {
	return h.currentCid
}

// CurrentBlock returns the current block.
func (h *DocHeadBlocksIterator) CurrentBlock() *coreblock.Block {
	return h.currentBlock
}

// CurrentRawBlock returns the raw data of the current block.
func (h *DocHeadBlocksIterator) CurrentRawBlock() []byte {
	return h.currentRawBlock
}
