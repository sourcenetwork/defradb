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
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/errors"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/utils"
)

// We don't want to have to hold large volumes of IDs in memory, so we chunk
// our deletes.
const hardDeleteChunkSize int = 10000

func (c *collection) Truncate(
	ctx context.Context, opts ...options.Enumerable[options.TruncateCollectionOptions],
) error {
	ctx, _, _ = getTxnAndSetCtxForCollection(ctx, c)

	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := c.db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeTruncateCollectionPerm); err != nil {
		return err
	}

	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return err
	}

	defer txn.Discard()

	err = c.truncate(ctx)
	if err != nil {
		return err
	}

	return txn.Commit()
}

func (c *collection) truncate(
	ctx context.Context,
) error {
	shortID, err := id.GetShortCollectionID(ctx, c.def.CollectionID)
	if err != nil {
		return err
	}

	txn := datastore.CtxMustGetTxn(ctx)
	c.db.lockSet.CollectionLock(txn, shortID)

	err = c.hardDeleteDocKeysAndHeadstore(ctx, shortID)
	if err != nil {
		return err
	}

	err = c.hardDeleteDatastorePrefix(ctx, keys.PrimaryDataStoreKey{
		CollectionShortID: shortID,
	})
	if err != nil {
		return err
	}

	err = c.hardDeleteDatastorePrefix(ctx, &keys.IndexDataStoreKey{
		CollectionShortID: shortID,
	})
	if err != nil {
		return err
	}

	err = c.hardDeleteDatastorePrefix(ctx, keys.DatastoreSE{
		CollectionShortID: shortID,
	})
	if err != nil {
		return err
	}

	err = c.hardDeleteDatastorePrefix(ctx, keys.ViewCacheKey{
		CollectionShortID: shortID,
	})
	if err != nil {
		return err
	}

	err = c.hardDeleteCollectionBlocks(ctx, shortID)
	if err != nil {
		return err
	}

	return nil
}

// hardDeleteDocKeysAndHeadstore iterates through the `keys.DataStoreKey` for this collection
// and deletes both them, *and* the headstore keys for those found documents.
//
// The headstore keys must be discovered based on datastore keys, as the headstore keys are not
// indexed by collection id, and so cannot be found independently.
func (c *collection) hardDeleteDocKeysAndHeadstore(
	ctx context.Context,
	colShortID uint32,
) error {
	txn := datastore.CtxMustGetTxn(ctx)

	prefix := keys.DataStoreKey{
		CollectionShortID: colShortID,
	}

	ds := txn.Datastore()

	iter, err := ds.Iterator(ctx, datastore.IterOptions{
		Prefix:   prefix,
		KeysOnly: true,
	})
	if err != nil {
		return NewErrCreateTruncateIterator(err)
	}

	keysToDelete := make([]keys.DataStoreKey, 0, hardDeleteChunkSize)
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

		key, err := keys.NewDataStoreKey(string(iter.Key()))
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
		err := ds.Delete(ctx, key)
		if err != nil {
			return NewErrTruncateDatastoreKey(err, key.ToString())
		}

		// Headstore keys are implicitly protected by the lockset on the datastore, as
		// any document-head writes are done in the same transaction as the datastore-document
		// writes.
		//
		// Because the datastore read-locks are only ever released when the transaction closes,
		// we do not need to worry about timing or order-of-operation issues, *unless* we change
		// when the datastore read-locks are released.
		err = c.hardDeleteDocumentBlocks(ctx, key.DocID)
		if err != nil {
			return err
		}
	}

	if hasMore {
		return c.hardDeleteDocKeysAndHeadstore(ctx, colShortID)
	}

	return nil
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
		return NewErrCreateTruncateIterator(err)
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
			return NewErrTruncateDatastoreKey(err, string(key))
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
		return NewErrCreateTruncateIterator(err)
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
			return NewErrTruncateHeadstoreKey(err, string(key.Bytes()))
		}

		err = deleteBlocks(ctx, key.Cid)
		if err != nil {
			return NewErrTruncateDeleteBlocks(err, key.Cid.String())
		}
	}

	if hasMore {
		return c.hardDeleteDocumentBlocks(ctx, docID)
	}

	return nil
}

func (c *collection) hardDeleteCollectionBlocks(
	ctx context.Context,
	shortID uint32,
) error {
	txn := datastore.CtxMustGetTxn(ctx)

	headstore := txn.Headstore()
	prefix := keys.HeadstoreColKey{
		CollectionShortID: shortID,
	}

	iter, err := headstore.Iterator(ctx, corekv.IterOptions{
		Prefix:   prefix.Bytes(),
		KeysOnly: true,
	})
	if err != nil {
		return NewErrCreateTruncateIterator(err)
	}

	keysToDelete := make([]keys.HeadstoreColKey, 0, hardDeleteChunkSize)
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

		key, err := keys.NewHeadstoreColKeyFromString(string(iter.Key()))
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
			return NewErrTruncateHeadstoreKey(err, string(key.Bytes()))
		}

		err = deleteBlocks(ctx, key.Cid)
		if err != nil {
			return NewErrTruncateDeleteBlocks(err, key.Cid.String())
		}
	}

	if hasMore {
		return c.hardDeleteCollectionBlocks(ctx, shortID)
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
			err = blockstore.DeleteBlock(ctx, decodedBlock.Encryption.Cid)
			if err != nil {
				return err
			}
		}

		if decodedBlock.Signature != nil {
			err = blockstore.DeleteBlock(ctx, decodedBlock.Signature.Cid)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
