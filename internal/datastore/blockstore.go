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

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/blockstore"
)

// mergedCacheSize is the number of merged CID entries to cache.
const mergedCacheSize = 100_000

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
	mergedCache, _ := lru.New[string, struct{}](mergedCacheSize)
	return &bstore{
		Blockstore:  blockstore.NewBlockstore(store),
		store:       store,
		mergedCache: mergedCache,
	}
}

type bstore struct {
	*blockstore.Blockstore

	store corekv.ReaderWriter
	// mergedCache caches CIDs that are known to be merged.
	mergedCache *lru.Cache[string, struct{}]
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
	cidStr := cid.String()
	if bs.mergedCache != nil {
		if _, ok := bs.mergedCache.Get(cidStr); ok {
			return true, nil
		}
	}
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
	merged := !notMerged
	if merged && bs.mergedCache != nil {
		bs.mergedCache.Add(cidStr, struct{}{})
	}
	return merged, nil
}

func (bs *bstore) MarkAsMerged(ctx context.Context, cid cid.Cid) error {
	err := bs.store.Delete(ctx, newToMergeKey(cid.Bytes()))
	if err != nil {
		return err
	}
	if bs.mergedCache != nil {
		bs.mergedCache.Add(cid.String(), struct{}{})
	}
	return nil
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

// blindWriteBlockstore is a blockstore wrapper that skips existence checks on Put.
type blindWriteBlockstore struct {
	*bstore
}

var _ Blockstore = (*blindWriteBlockstore)(nil)

// Put stores a block without checking if it already exists.
func (bs *blindWriteBlockstore) Put(ctx context.Context, block blocks.Block) error {
	return bs.store.Set(ctx, block.Cid().Bytes(), block.RawData())
}

// PutMany stores multiple blocks without checking if they already exist.
func (bs *blindWriteBlockstore) PutMany(ctx context.Context, blocks []blocks.Block) error {
	for _, b := range blocks {
		err := bs.store.Set(ctx, b.Cid().Bytes(), b.RawData())
		if err != nil {
			return err
		}
	}
	return nil
}

// carImportBlockstore is optimized for CAR imports with CID verification.
type carImportBlockstore struct {
	*bstore
}

var _ Blockstore = (*carImportBlockstore)(nil)

// verifyCID checks that the block's content hashes to its claimed CID.
func verifyCID(b blocks.Block) bool {
	computedCID, err := b.Cid().Prefix().Sum(b.RawData())
	if err != nil {
		return false
	}
	return computedCID.Equals(b.Cid())
}

// Put stores a block after verifying its CID matches the content.
func (bs *carImportBlockstore) Put(ctx context.Context, block blocks.Block) error {
	if !verifyCID(block) {
		return nil
	}
	err := bs.store.Set(ctx, newToMergeKey(block.Cid().Bytes()), []byte{objectMarker})
	if err != nil {
		return err
	}
	return bs.store.Set(ctx, block.Cid().Bytes(), block.RawData())
}

// PutMany stores multiple blocks after verifying each CID matches its content.
func (bs *carImportBlockstore) PutMany(ctx context.Context, blocks []blocks.Block) error {
	for _, b := range blocks {
		if !verifyCID(b) {
			continue
		}
		err := bs.store.Set(ctx, newToMergeKey(b.Cid().Bytes()), []byte{objectMarker})
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
