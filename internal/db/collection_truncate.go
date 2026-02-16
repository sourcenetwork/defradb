// Copyright 2025 Democratized Data Foundation
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
	ipld "github.com/ipfs/go-ipld-format"

	"github.com/sourcenetwork/corekv"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/errors"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// We don't want to have to hold large volumes of IDs in memory, so we chunk
// our deletes.
const hardDeleteChunkSize int = 10000

// truncateChunkSize is the number of truncate operations in each transaction.
const truncateChunkSize int = 500

func (c *collection) Truncate(
	ctx context.Context,
) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()
	if err := c.db.checkNodeAccess(ctx, acpTypes.NodeCollectionTruncatePerm); err != nil {
		return err
	}
	return c.truncateChunked(ctx)
}

// truncateChunked performs truncation using multiple smaller transactions.
func (c *collection) truncateChunked(ctx context.Context) error {
	origCtx := ctx

	txnCtx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return err
	}

	shortID, err := id.GetShortCollectionID(txnCtx, c.def.CollectionID)
	if err != nil {
		txn.Discard()
		return err
	}

	txn.Discard()
	ctx = origCtx

	for {
		hasMore, err := c.truncateDocKeysChunk(ctx, shortID)
		if err != nil {
			return err
		}
		if !hasMore {
			break
		}
	}

	for {
		hasMore, err := c.truncatePrefixChunk(ctx, shortID, keys.PrimaryDataStoreKey{
			CollectionShortID: shortID,
		})
		if err != nil {
			return err
		}
		if !hasMore {
			break
		}
	}

	for {
		hasMore, err := c.truncatePrefixChunk(ctx, shortID, &keys.IndexDataStoreKey{
			CollectionShortID: shortID,
		})
		if err != nil {
			return err
		}
		if !hasMore {
			break
		}
	}

	for {
		hasMore, err := c.truncatePrefixChunk(ctx, shortID, keys.DatastoreSE{
			CollectionShortID: shortID,
		})
		if err != nil {
			return err
		}
		if !hasMore {
			break
		}
	}

	for {
		hasMore, err := c.truncatePrefixChunk(ctx, shortID, keys.ViewCacheKey{
			CollectionShortID: shortID,
		})
		if err != nil {
			return err
		}
		if !hasMore {
			break
		}
	}

	for {
		hasMore, err := c.truncateCollectionBlocksChunk(ctx, shortID)
		if err != nil {
			return err
		}
		if !hasMore {
			break
		}
	}

	return nil
}

// truncateDocKeysChunk deletes one chunk of doc keys and their headstore entries.
func (c *collection) truncateDocKeysChunk(ctx context.Context, colShortID uint32) (bool, error) {
	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return false, err
	}
	defer txn.Discard()

	c.db.lockSet.CollectionLock(txn, colShortID)

	prefix := keys.DataStoreKey{
		CollectionShortID: colShortID,
	}

	ds := txn.Datastore()
	iter, err := ds.Iterator(ctx, datastore.IterOptions{
		Prefix:   prefix,
		KeysOnly: true,
	})
	if err != nil {
		return false, err
	}

	keysToDelete := make([]keys.DataStoreKey, 0, truncateChunkSize)
	hasMore := false

	for range truncateChunkSize {
		hasNext, err := iter.Next()
		if err != nil {
			return false, errors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}

		key, err := keys.NewDataStoreKey(string(iter.Key()))
		if err != nil {
			return false, errors.Join(err, iter.Close())
		}
		keysToDelete = append(keysToDelete, key)
	}

	if hasNext, _ := iter.Next(); hasNext {
		hasMore = true
	}

	if err := iter.Close(); err != nil {
		return false, err
	}

	if len(keysToDelete) == 0 {
		return false, nil
	}

	for _, key := range keysToDelete {
		if err := ds.Delete(ctx, key); err != nil {
			return false, err
		}
		if err := c.hardDeleteDocumentBlocks(ctx, key.DocID); err != nil {
			return false, err
		}
	}

	if err := txn.Commit(); err != nil {
		return false, err
	}

	return hasMore, nil
}

// truncatePrefixChunk deletes one chunk of keys with the given prefix.
func (c *collection) truncatePrefixChunk(ctx context.Context, colShortID uint32, prefix keys.Key) (bool, error) {
	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return false, err
	}
	defer txn.Discard()

	c.db.lockSet.CollectionLock(txn, colShortID)

	iter, err := txn.Datastore().Iterator(ctx, datastore.IterOptions{
		Prefix:   prefix,
		KeysOnly: true,
	})
	if err != nil {
		return false, err
	}

	keysToDelete := make([][]byte, 0, truncateChunkSize)
	hasMore := false

	for i := 0; i < truncateChunkSize; i++ {
		hasNext, err := iter.Next()
		if err != nil {
			return false, errors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}
		key := make([]byte, len(iter.Key()))
		copy(key, iter.Key())
		keysToDelete = append(keysToDelete, key)
	}

	if hasNext, _ := iter.Next(); hasNext {
		hasMore = true
	}
	if err := iter.Close(); err != nil {
		return false, err
	}
	if len(keysToDelete) == 0 {
		return false, nil
	}

	type unsafestore interface {
		Unsafe() corekv.ReaderWriter
	}
	ds, _ := txn.Datastore().(unsafestore)
	underlyingStore := ds.Unsafe()

	for _, key := range keysToDelete {
		if err := underlyingStore.Delete(ctx, key); err != nil {
			return false, err
		}
	}

	if err := txn.Commit(); err != nil {
		return false, err
	}

	return hasMore, nil
}

// truncateCollectionBlocksChunk deletes one chunk of collection blocks.
// Returns true if there are more blocks to delete.
func (c *collection) truncateCollectionBlocksChunk(ctx context.Context, colShortID uint32) (bool, error) {
	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return false, err
	}
	defer txn.Discard()

	c.db.lockSet.CollectionLock(txn, colShortID)

	headstore := txn.Headstore()
	prefix := keys.HeadstoreColKey{
		CollectionShortID: colShortID,
	}

	iter, err := headstore.Iterator(ctx, corekv.IterOptions{
		Prefix:   prefix.Bytes(),
		KeysOnly: true,
	})
	if err != nil {
		return false, err
	}

	keysToDelete := make([]keys.HeadstoreColKey, 0, truncateChunkSize)
	hasMore := false

	for range truncateChunkSize {
		hasNext, err := iter.Next()
		if err != nil {
			return false, errors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}

		key, err := keys.NewHeadstoreColKeyFromString(string(iter.Key()))
		if err != nil {
			return false, errors.Join(err, iter.Close())
		}
		keysToDelete = append(keysToDelete, key)
	}

	if hasNext, _ := iter.Next(); hasNext {
		hasMore = true
	}
	if err := iter.Close(); err != nil {
		return false, err
	}
	if len(keysToDelete) == 0 {
		return false, nil
	}

	blockstore := txn.Blockstore()

	for _, key := range keysToDelete {
		if err := headstore.Delete(ctx, key.Bytes()); err != nil {
			return false, err
		}
		if err := blockstore.DeleteBlock(ctx, key.Cid); err != nil {
			return false, err
		}
	}

	if err := txn.Commit(); err != nil {
		return false, err
	}

	return hasMore, nil
}

func (c *collection) hardDeleteDatastorePrefix(
	ctx context.Context,
	prefix keys.Key,
) error {
	txn := datastore.CtxMustGetTxn(ctx)

	iter, err := txn.Datastore().Iterator(ctx, datastore.IterOptions{
		Prefix:   prefix,
		KeysOnly: true,
	})
	if err != nil {
		return err
	}

	keysToDelete := make([][]byte, 0, hardDeleteChunkSize)
	// If there are more keys than we wish to load into memory at once, this will be set to
	// true, and we'll continue the delete in another pass.
	hasMore := true

	for i := 0; i < hardDeleteChunkSize; i++ {
		hasNext, err := iter.Next()
		if err != nil {
			return errors.Join(err, iter.Close())
		}
		if !hasNext {
			hasMore = false
			break
		}

		keysToDelete = append(keysToDelete, iter.Key())
	}

	err = iter.Close()
	if err != nil {
		return err
	}

	type unsafestore interface {
		Unsafe() corekv.ReaderWriter
	}
	datastore, _ := txn.Datastore().(unsafestore)

	// This `Unsafe` call is not technically required, it just allows us to
	// write this function using the `keys.Key` interface and call `Delete`
	// using an untyped key.
	//
	// Bypassing the lock system here is a safe side-effect, as this function
	// is only ever called within the context of a collection level write lock -
	// attempting to obtain a read lock would essentially be a no-op anyway.
	underlyingStore := datastore.Unsafe()

	for _, key := range keysToDelete {
		// Not all store implementations support mutations whilst iterating, so whilst it would
		// be simpler and probably more efficient to delete whilst iterating, it would not work
		// with all supported corekv store implementations.
		err := underlyingStore.Delete(ctx, key)
		if err != nil {
			return err
		}
	}

	if hasMore {
		return c.hardDeleteDatastorePrefix(ctx, prefix)
	}

	return nil
}

func (c *collection) hardDeleteDocumentBlocks(
	ctx context.Context,
	docID string,
) error {
	txn := datastore.CtxMustGetTxn(ctx)

	headstore := txn.Headstore()
	prefix := keys.HeadstoreDocKey{
		DocID: docID,
	}

	iter, err := headstore.Iterator(ctx, corekv.IterOptions{
		Prefix:   prefix.Bytes(),
		KeysOnly: true,
	})
	if err != nil {
		return err
	}

	keysToDelete := make([]keys.HeadstoreDocKey, 0, hardDeleteChunkSize)
	// If there are more keys than we wish to load into memory at once, this will be set to
	// true, and we'll continue the delete in another pass.
	hasMore := true

	for i := 0; i < hardDeleteChunkSize; i++ {
		hasNext, err := iter.Next()
		if err != nil {
			return errors.Join(err, iter.Close())
		}
		if !hasNext {
			hasMore = false
			break
		}

		key, err := keys.NewHeadstoreDocKey(string(iter.Key()))
		if err != nil {
			return errors.Join(err, iter.Close())
		}

		keysToDelete = append(keysToDelete, key)
	}

	err = iter.Close()
	if err != nil {
		return err
	}

	for _, key := range keysToDelete {
		// Not all store implementations support mutations whilst iterating, so whilst it would
		// be simpler and probably more efficient to delete whilst iterating, it would not work
		// with all supported corekv store implementations.
		err := headstore.Delete(ctx, key.Bytes())
		if err != nil {
			return err
		}

		err = deleteBlocks(ctx, key.Cid)
		if err != nil {
			return err
		}
	}

	if hasMore {
		return c.hardDeleteDocumentBlocks(ctx, docID)
	}

	return nil
}

// deleteBlocks deletes the block of the given cid and all the blocks it links to, if
// a block with this cid is found.
//
// If the block is not found, it will not error.
func deleteBlocks(ctx context.Context, head cid.Cid) error {
	txn := datastore.CtxMustGetTxn(ctx)
	blockstore := txn.Blockstore()

	toDelete := map[cid.Cid]struct{}{
		head: {},
	}
	for len(toDelete) != 0 {
		var currentBlockCid cid.Cid
		for v := range toDelete {
			// Pop the first key off of the `toDelete` set.
			currentBlockCid = v
			delete(toDelete, currentBlockCid)
			break
		}

		currentBlock, err := blockstore.Get(ctx, currentBlockCid)
		if errors.Is(err, ipld.ErrNotFound{}) {
			// We are looping through the links in a simple way that may result in us
			// attempting to delete blocks we have already deleted, this can include
			// blocks deleted by walking the dag pointed-to from another headstore key
			// (another call to `deleteBlocks`).
			//
			// If we encounter such a block, we can skip over the error and continue.
			continue
		}
		if err != nil {
			return err
		}

		err = blockstore.DeleteBlock(ctx, currentBlockCid)
		if err != nil {
			return err
		}

		decodedBlock, err := coreblock.GetFromBytes(currentBlock.RawData())
		if err != nil {
			return err
		}

		switch {
		case decodedBlock.Delta.IsField():
			// At the time of writing, field blocks do not have any links besides Encryption and Signature,
			// that will not already be linked to by other DAGs being deleted, so we have decided that the
			// compute that we will save by not trying to `Get` them is worth the risk of potentially missing
			// blocks in the future should this change.

		default:
			for _, link := range decodedBlock.AllLinks() {
				toDelete[link.Cid] = struct{}{}
			}
		}

		if decodedBlock.Encryption != nil {
			toDelete[decodedBlock.Encryption.Cid] = struct{}{}
		}

		if decodedBlock.Signature != nil {
			toDelete[decodedBlock.Signature.Cid] = struct{}{}
		}
	}

	return nil
}
