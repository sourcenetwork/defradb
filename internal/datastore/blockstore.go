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
	"errors"
	"time"

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
// single-byte marker) decodes as (zero, false), and the sweep leaves those alone.
func toMergeTime(v []byte) (time.Time, bool) {
	if len(v) != toMergeValueLen {
		return time.Time{}, false
	}
	return time.Unix(int64(binary.BigEndian.Uint64(v)), 0), true
}

// IsMerged reports whether the block is stored and carries no to-merge marker. Callers
// use it to decide whether to skip fetching, and history prune can delete a block at
// any time, so it is answered from the store rather than remembered.
func (bs *bstore) IsMerged(ctx context.Context, cid cid.Cid) (bool, error) {
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
	return !notMerged, nil
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

	// rootstore backs the marker refresh. The refresh reads the marker and writes it back, which
	// has to happen in one transaction or a merge clearing the marker in between is undone.
	rootstore corekv.TxnReaderWriter
}

var _ Blockstore = (*p2pBlockStore)(nil)

// Put stores a block to the blockstore.
func (bs *p2pBlockStore) Put(ctx context.Context, block blocks.Block) error {
	// Has is cheaper than Set, so see if we already have it
	exists, err := bs.store.Has(ctx, block.Cid().Bytes())
	if err == nil && exists {
		// Already stored, but a block still waiting to merge needs its marker renewed so the
		// sweep does not treat it as abandoned.
		return bs.refreshMarker(ctx, block)
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

// refreshMarker moves an unmerged block's marker forward so a block that keeps arriving is not
// treated as abandoned. The read and the write share a transaction: reading the marker puts it in
// the transaction's read set, so a merge that clears it concurrently makes this commit fail rather
// than silently marking a merged block unmerged again.
func (bs *p2pBlockStore) refreshMarker(ctx context.Context, block blocks.Block) error {
	markerKey := newToMergeKey(block.Cid().Bytes())

	txn := bs.rootstore.NewTxn(false)
	defer txn.Discard()
	txnCtx := corekv.SetCtxTxn(ctx, txn)

	stillUnmerged, err := bs.store.Has(txnCtx, markerKey)
	if err != nil {
		return NewErrCheckBlockMergeStatus(err)
	}
	if !stillUnmerged {
		return nil
	}
	if err := bs.store.Set(txnCtx, markerKey, newToMergeValue(time.Now())); err != nil {
		return NewErrStoreBlock(err)
	}

	// A conflict means another transaction wrote the marker after this one read it: a merge
	// clearing it, another refresh stamping it, or the sweep reclaiming the block along with
	// it. The first two leave the marker where this refresh would have; the third leaves no
	// block to mark, and the next fetch writes both again.
	if err := txn.Commit(); err != nil && !errors.Is(err, corekv.ErrTxnConflict) {
		return NewErrStoreBlock(err)
	}
	return nil
}

// PutMany stores multiple blocks to the blockstore.
func (bs *p2pBlockStore) PutMany(ctx context.Context, blocks []blocks.Block) error {
	for _, b := range blocks {
		exists, err := bs.store.Has(ctx, b.Cid().Bytes())
		if err == nil && exists {
			if err := bs.refreshMarker(ctx, b); err != nil {
				return err
			}
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
