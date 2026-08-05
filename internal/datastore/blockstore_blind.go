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
	"time"

	blocks "github.com/ipfs/go-block-format"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/immutable"
)

// blindWriteBlockstore is identical to p2pBlockStore but skips the Has() pre-check
// on Put/PutMany. Use it for write-only import paths (e.g. importCAR) where blocks
// are known to be new and the extra read round-trip adds no value.
//
// Because it does not read first, a block that has already merged can be given a fresh
// to-merge marker here. The orphan sweep checks document ownership rather than trusting
// the marker for that reason.
type blindWriteBlockstore struct {
	*bstore
}

var _ Blockstore = (*blindWriteBlockstore)(nil)

// BlindWriteP2PBlockstoreFrom returns a P2P blockstore that writes the to-merge index
// but skips the Has() guard on Put/PutMany. Use this on import paths where blocks are
// known to be new.
func BlindWriteP2PBlockstoreFrom(rootstore corekv.ReaderWriter, chunkSize immutable.Option[int]) Blockstore {
	return &blindWriteBlockstore{bstore: blockstoreFrom(rootstore, chunkSize)}
}

func (bs *blindWriteBlockstore) Put(ctx context.Context, block blocks.Block) error {
	if err := bs.store.Set(ctx, newToMergeKey(block.Cid().Bytes()), newToMergeValue(time.Now())); err != nil {
		return NewErrStoreBlock(err)
	}
	if err := bs.store.Set(ctx, block.Cid().Bytes(), block.RawData()); err != nil {
		return NewErrStoreBlock(err)
	}
	return nil
}

func (bs *blindWriteBlockstore) PutMany(ctx context.Context, blks []blocks.Block) error {
	for _, b := range blks {
		if err := bs.store.Set(ctx, newToMergeKey(b.Cid().Bytes()), newToMergeValue(time.Now())); err != nil {
			return NewErrStoreBlock(err)
		}
		if err := bs.store.Set(ctx, b.Cid().Bytes(), b.RawData()); err != nil {
			return NewErrStoreBlock(err)
		}
	}
	return nil
}
