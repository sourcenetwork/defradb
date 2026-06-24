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
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/errors"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/action"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/utils"
)

// We don't want to have to hold large volumes of IDs in memory, so we chunk
// our deletes.
const hardDeleteChunkSize int = 10000

type truncatePrefix []byte

func (p truncatePrefix) Bytes() []byte {
	return p
}

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

	ctx, txn, err := ensureContextTxnShim(ctx, c.db)
	if err != nil {
		return err
	}

	defer txn.Discard()

	multistore := datastore.NewMultistore(c.db.rootstore, c.db.lockSet, c.db.blockStoreChunkSize)

	shortID, err := id.GetUncachedShortCollectionID(ctx, c.def.CollectionID, multistore.Systemstore())
	if err != nil {
		return err
	}

	c.db.lockSet.CollectionLock(txn, shortID)

	// Clear the transaction on the context used to write the action execution information, otherwise
	// corekv will pick it up again, writing using the transaction.
	// https://github.com/sourcenetwork/corekv/issues/107
	txnFreeCtx := datastore.CtxSetTxn(ctx, nil)
	err = action.Register(txnFreeCtx, multistore, c.db.events, c.def.CollectionID, client.TruncateAction)
	if err != nil {
		return err
	}

	err = c.truncate(ctx)
	if err != nil {
		errErr := action.Set(
			txnFreeCtx,
			multistore,
			c.db.events,
			c.def.CollectionID,
			client.TruncateAction,
			client.ErroredActionStatus,
		)
		return errors.Join(errErr, err)
	}

	err = action.Complete(txnFreeCtx, multistore, c.db.events, c.def.CollectionID, client.TruncateAction)
	if err != nil {
		return err
	}

	return txn.Commit()
}

func (c *collection) truncate(
	ctx context.Context,
) error {
	sysStore := datastore.NewMultistore(c.db.rootstore, c.db.lockSet, c.db.blockStoreChunkSize).Systemstore()
	shortID, err := id.GetUncachedShortCollectionID(ctx, c.def.CollectionID, sysStore)
	if err != nil {
		return err
	}

	ctx = corekv.SetCtxTxn(ctx, nil)

	// The following operations must be performed without a transaction, due to store-level
	// transaction size limits.  This lack of protection means that they must be performed
	// in the order that will never result in orphaned key-values, so that a reattempt at the
	// truncate can eventually clear all store key-values within the collection.
	//
	// It is not possible to use inner transactions to protect the deletion of individual
	// documents as some stores such as leveldb do not support the opening of multiple transactions
	// at the same time.

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

// hardDeleteDocKeysAndHeadstore deletes document data and matching headstore keys for this collection.
// Datastore keys are used as the document index so block cleanup can happen before data keys are removed.
func (c *collection) hardDeleteDocKeysAndHeadstore(
	ctx context.Context,
	colShortID uint32,
) error {
	multistore := datastore.NewMultistore(c.db.rootstore, c.db.lockSet, c.db.blockStoreChunkSize)
	ds := multistore.Datastore()
	systemstore := multistore.Systemstore()

	deletedDocIDs := make(map[uint64]struct{})
	for _, instanceType := range []keys.InstanceType{keys.ValueKey, keys.PriorityKey, keys.DeletedKey} {
		instancePrefix := keys.DataStoreKey{
			CollectionShortID: colShortID,
			InstanceType:      instanceType,
		}
		if err := ds.Delete(ctx, instancePrefix); err != nil && !errors.Is(err, corekv.ErrNotFound) {
			return NewErrTruncateDatastoreKey(err, instancePrefix.ToString())
		}

		// If there are more keys than we wish to load into memory at once, this will be set to
		// true, and we'll continue the delete in another pass.
		hasMore := true

		for hasMore {
			iter, err := ds.Iterator(ctx, datastore.IterOptions{
				Prefix:   truncatePrefix(append(instancePrefix.Bytes(), '/')),
				KeysOnly: true,
			})
			if err != nil {
				return NewErrCreateTruncateIterator(err)
			}

			keysToDelete := make([]keys.DataStoreKey, 0, hardDeleteChunkSize)

			for i := 0; i < hardDeleteChunkSize; i++ {
				hasNext, err := iter.Next()
				if err != nil {
					return errors.Join(err, iter.Close())
				}
				if !hasNext {
					hasMore = false
					break
				}

				key, err := keys.DecodeDataStoreKey(iter.Key())
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
				// Headstore keys are implicitly protected by the lockset on the datastore, as
				// any document-head writes are done in the same transaction as the datastore-document
				// writes.
				//
				// Because the datastore read-locks are only ever released when the transaction closes,
				// we do not need to worry about timing or order-of-operation issues, *unless* we change
				// when the datastore read-locks are released.
				if key.DocShortID != 0 {
					if _, done := deletedDocIDs[key.DocShortID]; !done {
						err = c.hardDeleteDocumentBlocks(ctx, systemstore, colShortID, key.DocShortID)
						if err != nil {
							return err
						}
						if err := id.DeleteDocIDMappings(ctx, systemstore, colShortID, key.DocShortID); err != nil {
							return err
						}
						deletedDocIDs[key.DocShortID] = struct{}{}
					}
				}

				// Not all store implementations support mutations whilst iterating, so whilst it would
				// be simpler and probably more efficient to delete whilst iterating, it would not work
				// with all supported corekv store implementations.
				//
				// The deletion of the datastore key should be done after deleting the blocks - this way if
				// deleting a block errors, the index provided by the datastore key is preserved, and the
				// truncate can be resumed later.
				err := ds.Delete(ctx, key)
				if err != nil {
					return NewErrTruncateDatastoreKey(err, key.ToString())
				}
			}
		}
	}

	return nil
}

func (c *collection) hardDeleteDatastorePrefix(
	ctx context.Context,
	prefix keys.Key,
) error {
	ds := datastore.NewMultistore(c.db.rootstore, c.db.lockSet, c.db.blockStoreChunkSize).Datastore()

	// If there are more keys than we wish to load into memory at once, this will be set to
	// true, and we'll continue the delete in another pass.
	hasMore := true

	for hasMore {
		iter, err := ds.Iterator(ctx, datastore.IterOptions{
			Prefix:   prefix,
			KeysOnly: true,
		})
		if err != nil {
			return NewErrCreateTruncateIterator(err)
		}

		keysToDelete := make([][]byte, 0, hardDeleteChunkSize)

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
		datastore, ok := ds.(unsafestore)
		if !ok {
			return NewErrTruncateDatastoreKey(errors.New("datastore does not expose unsafe writer"), prefix.ToString())
		}

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
	}

	return nil
}

func (c *collection) hardDeleteDocumentBlocks(
	ctx context.Context,
	systemstore corekv.ReaderWriter,
	collectionShortID uint32,
	docShortID uint64,
) error {
	headstore := datastore.NewMultistore(c.db.rootstore, c.db.lockSet, c.db.blockStoreChunkSize).Headstore()

	// If there are more keys than we wish to load into memory at once, this will be set to
	// true, and we'll continue the delete in another pass.
	hasMore := true

	for hasMore {
		prefix := keys.HeadstoreDocKey{
			DocShortID: docShortID,
		}

		iter, err := headstore.Iterator(ctx, corekv.IterOptions{
			Prefix:   prefix.Bytes(),
			KeysOnly: true,
		})
		if err != nil {
			return NewErrCreateTruncateIterator(err)
		}

		keysToDelete := make([]keys.HeadstoreDocKey, 0, hardDeleteChunkSize)

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
			err = c.deleteBlocks(ctx, systemstore, collectionShortID, key.Cid)
			if err != nil {
				return NewErrTruncateDeleteBlocks(err, key.Cid.String())
			}

			// Not all store implementations support mutations whilst iterating, so whilst it would
			// be simpler and probably more efficient to delete whilst iterating, it would not work
			// with all supported corekv store implementations.
			//
			// The deletion of the headstore key should be done after deleting the blocks - this way if
			// deleting a block errors, the index provided by the headstore key is preserved, and the
			// truncate can be resumed later.
			err := headstore.Delete(ctx, key.Bytes())
			if err != nil {
				return NewErrTruncateHeadstoreKey(err, string(key.Bytes()))
			}
		}
	}

	return nil
}

func (c *collection) hardDeleteCollectionBlocks(
	ctx context.Context,
	shortID uint32,
) error {
	headstore := datastore.NewMultistore(c.db.rootstore, c.db.lockSet, c.db.blockStoreChunkSize).Headstore()

	// If there are more keys than we wish to load into memory at once, this will be set to
	// true, and we'll continue the delete in another pass.
	hasMore := true

	for hasMore {
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
			err = c.deleteBlocks(ctx, nil, 0, key.Cid)
			if err != nil {
				return NewErrTruncateDeleteBlocks(err, key.Cid.String())
			}

			// Not all store implementations support mutations whilst iterating, so whilst it would
			// be simpler and probably more efficient to delete whilst iterating, it would not work
			// with all supported corekv store implementations.
			//
			// The deletion of the headstore key should be done after deleting the blocks - this way if
			// deleting a block errors, the index provided by the headstore key is preserved, and the
			// truncate can be resumed later.
			err := headstore.Delete(ctx, key.Bytes())
			if err != nil {
				return NewErrTruncateHeadstoreKey(err, string(key.Bytes()))
			}
		}
	}

	return nil
}

// deleteBlocks deletes the block of the given cid and all the blocks it links to, if
// a block with this cid is found.
//
// If the block is not found, it will not error.
func (c *collection) deleteBlocks(
	ctx context.Context,
	systemstore corekv.ReaderWriter,
	collectionShortID uint32,
	currentCid cid.Cid,
) error {
	blockstore := datastore.NewMultistore(c.db.rootstore, c.db.lockSet, c.db.blockStoreChunkSize).Blockstore()

	deleteBlockMapping := func(blockCID cid.Cid) error {
		if systemstore == nil || collectionShortID == 0 {
			return nil
		}
		return id.DeleteBlockDocIDMapping(ctx, systemstore, blockCID)
	}

	deleteBlock := func(blockCID cid.Cid) error {
		if err := blockstore.DeleteBlock(ctx, blockCID); err != nil {
			return err
		}
		return deleteBlockMapping(blockCID)
	}

	type block struct {
		id    cid.Cid
		block *coreblock.Block
	}

	coreBlock, isFound, err := getBlock(ctx, blockstore, currentCid)
	if err != nil {
		return err
	}
	if !isFound {
		return deleteBlockMapping(currentCid)
	}

	toDelete := []*block{
		{
			id:    currentCid,
			block: coreBlock,
		},
	}

	i := -1
	isReversed := false
	increment := func() bool {
		if isReversed {
			i--
		} else {
			i++
		}

		if !isReversed && i == len(toDelete) {
			// if we have reached the end of the set, reverse direction - the children are now
			// gaurenteed to be deleted before their parents.
			isReversed = true
			i--
			return true
		}

		if i == -1 && isReversed {
			// we only need to iterate through twice, once in either direction, once we have finished
			// iterating in reverse, all blocks should have been deleted.
			return false
		}

		return true
	}

	for increment() {
		currentBlock := toDelete[i]

		if currentBlock.block == nil {
			coreBlock, isFound, err := getBlock(ctx, blockstore, currentBlock.id)
			if err != nil {
				return err
			}
			if !isFound {
				if err := deleteBlockMapping(currentBlock.id); err != nil {
					return err
				}
				continue
			}

			currentBlock.block = coreBlock
		}

		if currentBlock.block.Encryption != nil {
			err := deleteBlock(currentBlock.block.Encryption.Cid)
			if err != nil {
				return err
			}
		}

		if currentBlock.block.Signature != nil {
			err := deleteBlock(currentBlock.block.Signature.Cid)
			if err != nil {
				return err
			}
		}

		if isReversed {
			// If we are now iterating in reverse order, all the children of this block should
			// have been deleted, and we are now free to delete this block.
			err := deleteBlock(currentBlock.id)
			if err != nil {
				return err
			}
			continue
		}

		switch {
		case currentBlock.block.Delta.IsField():
			// If this block is a field block, we can delete it immediately after blocks such
			// as encryption and signature are deleted.  Whilst it may have children, these
			// children will be referenced by the composite commit, which will only be deleted
			// after all of *its* children, meaning nothing will be orphaned if an error is thrown
			// at some point during the truncate.
			err := deleteBlock(currentBlock.id)
			if err != nil {
				return err
			}

		default:
			for _, link := range currentBlock.block.AllLinks() {
				toDelete = append(toDelete, &block{
					id: link.Cid,
				})
			}
		}
	}

	return nil
}

func getBlock(ctx context.Context, blockstore datastore.Blockstore, id cid.Cid) (*coreblock.Block, bool, error) {
	rawBlock, err := blockstore.Get(ctx, id)
	if errors.Is(err, ipld.ErrNotFound{}) {
		// We are looping through the links in a simple way that may result in us
		// attempting to delete blocks we have already deleted, this can include
		// blocks deleted by walking the dag pointed-to from another headstore key
		// (another call to `deleteBlocks`).
		//
		// If we encounter such a block, we can skip over the error and continue.
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	decodedBlock, err := coreblock.GetFromBytes(rawBlock.RawData())
	if err != nil {
		return nil, false, err
	}

	return decodedBlock, true, nil
}
