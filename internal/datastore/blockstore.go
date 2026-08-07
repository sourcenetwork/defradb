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
	"encoding/binary"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	ipfsBlockstore "github.com/ipfs/boxo/blockstore"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/blockstore"
)

const mergedCacheSize = 100_000

// globalMergedCache is a process-wide LRU cache of CIDs known to be merged.
// A cache hit in IsMerged avoids 2 KV reads per CID, which matters under sustained
// P2P or batch write load where the same CID chains are traversed repeatedly.
//
// Only IsMerged writes to it, and only from committed state. An entry added inside a
// transaction would survive a rollback and report a block that was never stored.
var globalMergedCache = mustNewMergedCache(mergedCacheSize)

func mustNewMergedCache(size int) *lru.Cache[string, struct{}] {
	cache, err := lru.New[string, struct{}](size)
	if err != nil {
		panic(err)
	}
	return cache
}

// Blockstore proxies the ipld.DAGService under the /core namespace for future-proofing
type Blockstore interface {
	ipfsBlockstore.Blockstore
	// Mark the block as merged by removing the to-merge index.
	MarkAsMerged(ctx context.Context, k cid.Cid) error
	// BatchMarkAsMerged marks multiple blocks as merged in a single pass.
	BatchMarkAsMerged(ctx context.Context, cids []cid.Cid) error
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
	toMergeIndexPrefix = byte('m')
	// toMergeValueLen is the length of a current-format marker value: an 8-byte
	// big-endian unix timestamp of when the block was fetched.
	toMergeValueLen = 8
)

func newToMergeKey(cid []byte) []byte {
	l := len(cid)
	key := make([]byte, l+1)
	copy(key[1:], cid)
	key[0] = toMergeIndexPrefix
	return key
}

// newToMergeValue encodes the time a block was fetched. Only the orphan sweep reads
// this value, to tell a fetch still in flight from an abandoned one; IsMerged checks
// the marker's presence, not its contents.
func newToMergeValue(t time.Time) []byte {
	v := make([]byte, toMergeValueLen)
	binary.BigEndian.PutUint64(v, uint64(t.Unix()))
	return v
}

// toMergeTime decodes a marker value. A value that is not a full timestamp (an older
// single-byte marker) decodes as (zero, false); the sweep treats those as abandoned.
func toMergeTime(v []byte) (time.Time, bool) {
	if len(v) != toMergeValueLen {
		return time.Time{}, false
	}
	return time.Unix(int64(binary.BigEndian.Uint64(v)), 0), true
}

func (bs *bstore) IsMerged(ctx context.Context, cid cid.Cid) (bool, error) {
	cidStr := cid.String()
	if _, ok := globalMergedCache.Get(cidStr); ok {
		return true, nil
	}
	hasBlock, err := bs.Has(ctx, cid)
	if err != nil {
		return false, NewErrCheckBlockExists(err)
	}
	if !hasBlock {
		return false, nil
	}
	notMerged, err := bs.store.Has(ctx, newToMergeKey(cid.Bytes()))
	if err != nil {
		return false, NewErrCheckBlockMergeStatus(err)
	}
	merged := !notMerged
	if merged {
		globalMergedCache.Add(cidStr, struct{}{})
	}
	return merged, nil
}

func (bs *bstore) MarkAsMerged(ctx context.Context, cid cid.Cid) error {
	if err := bs.store.Delete(ctx, newToMergeKey(cid.Bytes())); err != nil {
		return NewErrMarkBlockAsMerged(err)
	}
	return nil
}

func (bs *bstore) BatchMarkAsMerged(ctx context.Context, cids []cid.Cid) error {
	for _, c := range cids {
		if err := bs.store.Delete(ctx, newToMergeKey(c.Bytes())); err != nil {
			return NewErrMarkBlockAsMerged(err)
		}
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
	err = bs.store.Set(ctx, newToMergeKey(block.Cid().Bytes()), newToMergeValue(time.Now()))
	if err != nil {
		return NewErrStoreBlock(err)
	}
	err = bs.store.Set(ctx, block.Cid().Bytes(), block.RawData())
	if err != nil {
		return NewErrStoreBlock(err)
	}
	return nil
}

// PutMany stores multiple blocks to the blockstore.
func (bs *p2pBlockStore) PutMany(ctx context.Context, blocks []blocks.Block) error {
	for _, b := range blocks {
		exists, err := bs.store.Has(ctx, b.Cid().Bytes())
		if err == nil && exists {
			continue
		}
		err = bs.store.Set(ctx, newToMergeKey(b.Cid().Bytes()), newToMergeValue(time.Now()))
		if err != nil {
			return NewErrStoreBlock(err)
		}
		err = bs.store.Set(ctx, b.Cid().Bytes(), b.RawData())
		if err != nil {
			return NewErrStoreBlock(err)
		}
	}
	return nil
}
