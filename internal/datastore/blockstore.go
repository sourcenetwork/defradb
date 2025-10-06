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

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/storage/bsadapter"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/blockstore"
)

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

// AsIPLDStorage returns an IPLDStorage instance.
//
// It wraps the blockstore in an IPLD Blockstore adapter for use with
// the IPLD LinkSystem.
func (bs *bstore) AsIPLDStorage() IPLDStorage {
	return &bsadapter.Adapter{Wrapped: bs}
}

const (
	objectMarker       = byte(0xff)
	toMergeIndexPrefix = "/tm"
)

func newToMergeKey(cid string) []byte {
	return []byte(toMergeIndexPrefix + "/" + cid)
}

func (bs *bstore) IsMerged(ctx context.Context, cid cid.Cid) (bool, error) {
	hasBlock, err := bs.Has(ctx, cid)
	if err != nil {
		return false, err
	}
	if !hasBlock {
		return false, nil
	}
	notMerged, err := bs.store.Has(ctx, newToMergeKey(cid.String()))
	if err != nil {
		return false, err
	}
	return !notMerged, nil
}

func (bs *bstore) MarkAsMerged(ctx context.Context, cid cid.Cid) error {
	return bs.store.Delete(ctx, newToMergeKey(cid.String()))
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
	err = bs.store.Set(ctx, newToMergeKey(block.Cid().String()), []byte{objectMarker})
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
		err = bs.store.Set(ctx, newToMergeKey(b.Cid().String()), []byte{objectMarker})
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
