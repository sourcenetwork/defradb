// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package store

import (
	"context"
	"errors"

	"github.com/sourcenetwork/defradb/core"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	dshelp "github.com/ipfs/go-ipfs-ds-help"
)

// Blockstore implementation taken from https://github.com/ipfs/go-ipfs-blockstore/blob/master/blockstore.go
// Needed a custom implementation that didn't rely on the ds.Batching interface.
//
// All datastore operations in DefraDB are interfaced by core.DSReaderWriter. This simplifies the interface to just
// that of read/write operations, leaving the management of the datastore to the parent objects. This also allows
// us to swap between a regular ds.Datastore, and a ds.Txn which as of https://github.com/ipfs/go-datastore/issues/114
// no longer implements ds.Datastore.
//
// The orginal blockstore.Blockstore implementation relied on ds.Batching, so it could internally use store.Batch()
// to optimize the PutMany function. However, in DefraDB, since we rely on a single rootstore for all our various
// substores (data, heads, blocks), which includes a Txn/Batch system already, our respective substores don't need
// to optimize or worry about Batching/Txn. Hence the simplified core.DSReaderWriter.

// ErrHashMismatch is an error returned when the hash of a block
// is different than expected.
var ErrHashMismatch = errors.New("block in storage has different hash than requested")

// ErrNotFound is an error returned when a block is not found.
var ErrNotFound = errors.New("blockstore: block not found")

// NewBlockstore returns a default Blockstore implementation
// using the provided datastore.Batching backend.
func NewBlockstore(store core.DSReaderWriter) blockstore.Blockstore {
	return &bstore{
		store: store,
	}
}

type bstore struct {
	store core.DSReaderWriter

	rehash bool
}

func (bs *bstore) HashOnRead(_ context.Context, enabled bool) {
	bs.rehash = enabled
}

func (bs *bstore) Get(ctx context.Context, k cid.Cid) (blocks.Block, error) {
	if !k.Defined() {
		log.Error("undefined cid in blockstore")
		return nil, ErrNotFound
	}
	bdata, err := bs.store.Get(ctx, dshelp.MultihashToDsKey(k.Hash()))
	if err == ds.ErrNotFound {
		return nil, ErrNotFound
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

func (bs *bstore) Put(ctx context.Context, block blocks.Block) error {
	k := dshelp.MultihashToDsKey(block.Cid().Hash())

	// Has is cheaper than Put, so see if we already have it
	exists, err := bs.store.Has(ctx, k)
	if err == nil && exists {
		return nil // already stored.
	}
	return bs.store.Put(ctx, k, block.RawData())
}

func (bs *bstore) PutMany(ctx context.Context, blocks []blocks.Block) error {
	for _, b := range blocks {
		k := dshelp.MultihashToDsKey(b.Cid().Hash())
		exists, err := bs.store.Has(ctx, k)
		if err == nil && exists {
			continue
		}

		err = bs.store.Put(ctx, k, b.RawData())
		if err != nil {
			return err
		}
	}
	return nil
}

func (bs *bstore) Has(ctx context.Context, k cid.Cid) (bool, error) {
	return bs.store.Has(ctx, dshelp.MultihashToDsKey(k.Hash()))
}

func (bs *bstore) GetSize(ctx context.Context, k cid.Cid) (int, error) {
	size, err := bs.store.GetSize(ctx, dshelp.MultihashToDsKey(k.Hash()))
	if err == ds.ErrNotFound {
		return -1, ErrNotFound
	}
	return size, err
}

func (bs *bstore) DeleteBlock(ctx context.Context, k cid.Cid) error {
	return bs.store.Delete(ctx, dshelp.MultihashToDsKey(k.Hash()))
}

// AllKeysChan runs a query for keys from the blockstore.
// this is very simplistic, in the future, take dsq.Query as a param?
//
// AllKeysChan respects context.
func (bs *bstore) AllKeysChan(ctx context.Context) (<-chan cid.Cid, error) {

	// KeysOnly, because that would be _a lot_ of data.
	q := dsq.Query{KeysOnly: true}
	res, err := bs.store.Query(ctx, q)
	if err != nil {
		return nil, err
	}

	output := make(chan cid.Cid, dsq.KeysOnlyBufSize)
	go func() {
		defer func() {
			res.Close() // ensure exit (signals early exit, too)
			close(output)
		}()

		for {
			e, ok := res.NextSync()
			if !ok {
				return
			}
			if e.Error != nil {
				log.Errorf("blockstore.AllKeysChan got err: %s", e.Error)
				return
			}

			// need to convert to key.Key using key.KeyFromDsKey.
			bk, err := dshelp.BinaryFromDsKey(ds.RawKey(e.Key))
			if err != nil {
				log.Warningf("error parsing key from binary: %s", err)
				continue
			}
			k := cid.NewCidV1(cid.Raw, bk)
			select {
			case <-ctx.Done():
				return
			case output <- k:
			}
		}
	}()

	return output, nil
}
