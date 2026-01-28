// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package datastore

import (
	"context"

	ipfsBlockstore "github.com/ipfs/boxo/blockstore"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/blockstore"
)

// Blockstore proxies the ipld.DAGService under the /core namespace for future-proofing
type Blockstore interface {
	ipfsBlockstore.Blockstore
	// Mark the block as merged by removing the to-merge index.
	MarkAsMerged(ctx context.Context, k cid.Cid) error
	// Check if the block has been merged. It will return false if either the CID is not found
	// or the CID is found AND the to-merge index is also found.
	IsMerged(ctx context.Context, k cid.Cid) (bool, error)
}

func newBlockstore(store corekv.ReaderWriter) *bstore {
	return &bstore{
		Blockstore: blockstore.NewBlockstore(store),
		store:      store,
	}
}

type bstore struct {
	*blockstore.Blockstore

	store corekv.ReaderWriter
}

var _ Blockstore = (*bstore)(nil)

const (
	objectMarker       = byte(0xff)
	toMergeIndexPrefix = byte('m')
)

func newToMergeKey(cid []byte) []byte {
	l := len(cid)
	key := make([]byte, l+1)
	copy(key[1:], cid)
	key[0] = toMergeIndexPrefix
	return key
}

func (bs *bstore) IsMerged(ctx context.Context, cid cid.Cid) (bool, error) {
	hasBlock, err := bs.Has(ctx, cid)
	if err != nil {
		return false, err
	}
	if !hasBlock {
		return false, nil
	}
	notMerged, err := bs.store.Has(ctx, newToMergeKey(cid.Bytes()))
	if err != nil {
		return false, err
	}
	return !notMerged, nil
}

func (bs *bstore) MarkAsMerged(ctx context.Context, cid cid.Cid) error {
	return bs.store.Delete(ctx, newToMergeKey(cid.Bytes()))
}

type p2pBlockStore struct {
	*bstore
}

var _ Blockstore = (*p2pBlockStore)(nil)

// Put stores a block to the blockstore.
func (bs *p2pBlockStore) Put(ctx context.Context, block blocks.Block) error {
	// Has is cheaper than Set, so see if we already have it
	exists, err := bs.store.Has(ctx, block.Cid().Bytes())
	if err == nil && exists {
		return nil // already stored.
	}
	err = bs.store.Set(ctx, newToMergeKey(block.Cid().Bytes()), []byte{objectMarker})
	if err != nil {
		return err
	}
	return bs.store.Set(ctx, block.Cid().Bytes(), block.RawData())
}

// PutMany stores multiple blocks to the blockstore.
func (bs *p2pBlockStore) PutMany(ctx context.Context, blocks []blocks.Block) error {
	for _, b := range blocks {
		exists, err := bs.store.Has(ctx, b.Cid().Bytes())
		if err == nil && exists {
			continue
		}
		err = bs.store.Set(ctx, newToMergeKey(b.Cid().Bytes()), []byte{objectMarker})
		if err != nil {
			return err
		}
		err = bs.store.Set(ctx, b.Cid().Bytes(), b.RawData())
		if err != nil {
			return err
		}
	}
	return nil
}
