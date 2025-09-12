// Copyright 2022 Democratized Data Foundation
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
	"bytes"
	"context"

	"github.com/ipfs/boxo/blockstore"
	dshelp "github.com/ipfs/boxo/datastore/dshelp"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipld/go-ipld-prime/storage/bsadapter"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/errors"
)

// Blockstore implementation taken from:
//  `https://github.com/ipfs/go-ipfs-blockstore/blob/master/blockstore.go`
// Needed a custom implementation that didn't rely on the ds.Batching interface.
//
// All datastore operations in DefraDB are interfaced by ReaderWriter. This
// simplifies the interface to just that of read/write operations, leaving the
// management of the datastore to the parent objects. This also allows us to swap
// between a regular ds.Datastore, and a ds.Txn which as of:
//  `https://github.com/ipfs/go-datastore/issues/114` no longer implements ds.Datastore.
//
// The original blockstore.Blockstore implementation relied on ds.Batching, so it
// could internally use store.Batch() to optimize the PutMany function.

// However, in DefraDB, since we rely on a single rootstore for all our various
// substores (data, heads, blocks), which includes a Txn/Batch system already, our
// respective substores don't need to optimize or worry about Batching/Txn.
// Hence the simplified ReaderWriter.

// NewBlockstore returns a default Blockstore implementation
// using the provided datastore.Batching backend.
func newBlockstore(store corekv.ReaderWriter) *bstore {
	return &bstore{
		store: store,
	}
}

func newIPLDStore(store blockstore.Blockstore) *bsadapter.Adapter {
	return &bsadapter.Adapter{Wrapped: store}
}

type bstore struct {
	store corekv.ReaderWriter

	rehash bool
}

var _ Blockstore = (*bstore)(nil)

// AsIPLDStorage returns an IPLDStorage instance.
//
// It wraps the blockstore in an IPLD Blockstore adapter for use with
// the IPLD LinkSystem.
func (bs *bstore) AsIPLDStorage() IPLDStorage {
	return newIPLDStore(bs)
}

// HashOnRead enables or disables rehashing of blocks on read.
func (bs *bstore) HashOnRead(enabled bool) {
	bs.rehash = enabled
}

// Get returns a block from the blockstore.
func (bs *bstore) Get(ctx context.Context, k cid.Cid) (blocks.Block, error) {
	if !k.Defined() {
		return nil, ipld.ErrNotFound{Cid: k}
	}
	mergedKey, notMergedKey := getKeys(k)
	bdata, err := bs.store.Get(ctx, mergedKey)
	if errors.Is(err, corekv.ErrNotFound) {
		bdata, err = bs.store.Get(ctx, notMergedKey)
		if errors.Is(err, corekv.ErrNotFound) {
			return nil, ipld.ErrNotFound{Cid: k}
		}
	}
	if err != nil {
		return nil, err
	}
	if bs.rehash {
		rbcid, err := k.Prefix().Sum(bdata)
		if err != nil {
			return nil, err
		}

		if !rbcid.Equals(k) {
			return nil, ErrHashMismatch
		}

		return blocks.NewBlockWithCid(bdata, rbcid)
	}
	return blocks.NewBlockWithCid(bdata, k)
}

// Put stores a block to the blockstore.
func (bs *bstore) Put(ctx context.Context, block blocks.Block) error {
	mergedKey, notMergedKey := getKeys(block.Cid())

	// Has is cheaper than Set, so see if we already have it
	exists, err := bs.store.Has(ctx, mergedKey)
	if err == nil && exists {
		return nil // already stored.
	}
	exists, err = bs.store.Has(ctx, notMergedKey)
	if err == nil && exists {
		return nil // already stored.
	}
	return bs.store.Set(ctx, notMergedKey, block.RawData())
}

// PutMany stores multiple blocks to the blockstore.
func (bs *bstore) PutMany(ctx context.Context, blocks []blocks.Block) error {
	for _, b := range blocks {
		mergedKey, notMergedKey := getKeys(b.Cid())
		exists, err := bs.store.Has(ctx, mergedKey)
		if err == nil && exists {
			continue
		}
		exists, err = bs.store.Has(ctx, notMergedKey)
		if err == nil && exists {
			continue
		}

		err = bs.store.Set(ctx, notMergedKey, b.RawData())
		if err != nil {
			return err
		}
	}
	return nil
}

// Has returns whether a block is stored in the blockstore.
func (bs *bstore) Has(ctx context.Context, k cid.Cid) (bool, error) {
	mergedKey, notMergedKey := getKeys(k)
	exists, err := bs.store.Has(ctx, mergedKey)
	if err == nil && exists {
		return true, nil
	}
	return bs.store.Has(ctx, notMergedKey)
}

// GetSize returns the size of a block in the blockstore.
func (bs *bstore) GetSize(ctx context.Context, k cid.Cid) (int, error) {
	mergedKey, notMergedKey := getKeys(k)
	buf, err := bs.store.Get(ctx, mergedKey)
	if errors.Is(err, corekv.ErrNotFound) {
		buf, err = bs.store.Get(ctx, notMergedKey)
		if errors.Is(err, corekv.ErrNotFound) {
			return -1, ipld.ErrNotFound{Cid: k}
		}
	}
	return len(buf), err
}

// DeleteBlock removes a block from the blockstore.
func (bs *bstore) DeleteBlock(ctx context.Context, k cid.Cid) error {
	mergedKey, notMergedKey := getKeys(k)
	err := bs.store.Delete(ctx, mergedKey)
	if errors.Is(err, corekv.ErrNotFound) {
		return bs.store.Delete(ctx, notMergedKey)
	}
	return err
}

// AllKeysChan runs a query for keys from the blockstore.
//
// AllKeysChan respects context.
//
// TODO this is very simplistic, in the future, take dsq.Query as a param?
func (bs *bstore) AllKeysChan(ctx context.Context) (<-chan cid.Cid, error) {
	// KeysOnly, because that would be _a lot_ of data.
	iter, err := bs.store.Iterator(ctx, corekv.IterOptions{
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

	output := make(chan cid.Cid, dsq.KeysOnlyBufSize)
	go func() {
		defer func() {
			//nolint:errcheck
			iter.Close() // ensure exit (signals early exit, too)
			close(output)
		}()

		for {
			hasNext, err := iter.Next()
			if err != nil {
				log.ErrorContextE(ctx, "Error iterating through keys", err)
				break
			}

			if !hasNext {
				break
			}

			key := cleanKey(iter.Key())
			hash, err := dshelp.DsKeyToMultihash(ds.RawKey(string(key)))
			if err != nil {
				log.ErrorContextE(ctx, "Error parsing key from binary", err)
				continue
			}
			k := cid.NewCidV1(cid.Raw, hash)
			select {
			case <-ctx.Done():
				return
			case output <- k:
			}
		}
	}()

	return output, nil
}

// MarkAsMerged sets a block as merged.
func (bs *bstore) MarkAsMerged(ctx context.Context, k cid.Cid) error {
	mergedKey, notMergedKey := getKeys(k)
	exists, err := bs.store.Has(ctx, mergedKey)
	if err == nil && exists {
		// already marked as merged
		return nil
	}
	data, err := bs.store.Get(ctx, notMergedKey)
	if errors.Is(err, corekv.ErrNotFound) {
		return ipld.ErrNotFound{Cid: k}
	}
	err = bs.store.Delete(ctx, notMergedKey)
	if err != nil {
		return err
	}
	return bs.store.Set(ctx, mergedKey, data)
}

// IsMerged returns whether a block has been marked as merged.
func (bs *bstore) IsMerged(ctx context.Context, k cid.Cid) (bool, error) {
	mergedKey, _ := getKeys(k)
	exists, err := bs.store.Has(ctx, mergedKey)
	if err == nil && exists {
		// already marked as merged
		return true, nil
	}
	return false, err
}

// getKeys returns both the mergedKey and the notMergedKey for a given cid.
func getKeys(k cid.Cid) (mergedKey, notMergedKey []byte) {
	key := dshelp.MultihashToDsKey(k.Hash())
	mergedKey = append(key.Bytes(), []byte("/1")...)
	notMergedKey = append(key.Bytes(), []byte("/0")...)
	return mergedKey, notMergedKey
}

// cleanKey removes the merged or notMerged suffix from the key.
func cleanKey(key []byte) []byte {
	return bytes.TrimSuffix(bytes.TrimSuffix(key, []byte("/0")), []byte("/1"))
}
