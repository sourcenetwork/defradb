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

	collectionShortID, err := id.GetUncachedCollectionShortID(ctx, c.def.CollectionID, multistore.Systemstore())
	if err != nil {
		return err
	}

	c.db.lockSet.CollectionLock(txn, collectionShortID)

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
	collectionShortID, err := id.GetUncachedCollectionShortID(ctx, c.def.CollectionID, sysStore)
	if err != nil {
		return err
	}

	// Truncate deletes through root stores, so keep these operations out of the outer txn.
	ctx = corekv.SetCtxTxn(ctx, nil)

	// The following operations must be performed without a transaction, due to store-level
	// transaction size limits.  This lack of protection means that they must be performed
	// in the order that will never result in orphaned key-values, so that a reattempt at the
	// truncate can eventually clear all store key-values within the collection.
	//
	// It is not possible to use inner transactions to protect the deletion of individual
	// documents as some stores such as leveldb do not support the opening of multiple transactions
	// at the same time.

	err = c.hardDeleteDocKeysAndHeadstore(ctx, collectionShortID)
	if err != nil {
		return err
	}

	err = c.hardDeleteDatastorePrefix(ctx, keys.PrimaryDataStoreKey{
		CollectionShortID: collectionShortID,
	})
	if err != nil {
		return err
	}

	err = c.hardDeleteDatastorePrefix(ctx, &keys.IndexDataStoreKey{
		CollectionShortID: collectionShortID,
	})
	if err != nil {
		return err
	}

	err = c.hardDeleteDatastorePrefix(ctx, keys.DatastoreSE{
		CollectionShortID: collectionShortID,
	})
	if err != nil {
		return err
	}

	err = c.hardDeleteDatastorePrefix(ctx, keys.ViewCacheKey{
		CollectionShortID: collectionShortID,
	})
	if err != nil {
		return err
	}

	err = c.hardDeleteCollectionBlocks(ctx, collectionShortID)
	if err != nil {
		return err
	}

	return nil
}

// hardDeleteDocKeysAndHeadstore deletes document data and matching headstore keys for this collection.
// Datastore keys are used as the document index so block cleanup can happen before data keys are removed.
func (c *collection) hardDeleteDocKeysAndHeadstore(
	ctx context.Context,
	collectionShortID uint32,
) error {
	multistore := datastore.NewMultistore(c.db.rootstore, c.db.lockSet, c.db.blockStoreChunkSize)
	ds := multistore.Datastore()
	systemstore := multistore.Systemstore()

	deletedDocIDs := make(map[uint64]struct{})
	for _, instanceType := range []keys.InstanceType{keys.ValueKey, keys.PriorityKey, keys.DeletedKey} {
		instancePrefix := keys.DataStoreKey{
			CollectionShortID: collectionShortID,
			InstanceType:      instanceType,
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
						err = c.removeDocumentHeads(ctx, systemstore, key.DocShortID, true, nil)
						if err != nil {
							return err
						}
						if err := id.DeleteDocIDMappings(ctx, systemstore, key.DocShortID); err != nil {
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
			return NewErrTruncateDatastoreKey(
				errors.New(errUnsafeDatastoreWriter),
				prefix.ToString(),
			)
		}

		// Unsafe permits untyped keys here. The collection lock already excludes concurrent writes.
		underlyingStore := datastore.Unsafe()

		for _, key := range keysToDelete {
			err := underlyingStore.Delete(ctx, key)
			if err != nil {
				return NewErrTruncateDatastoreKey(err, string(key))
			}
		}
	}

	return nil
}

// removeDocumentHeads removes document head entries and block ownership edges.
// When pruneBlocks is true, blocks without another owner are deleted as well.
func (c *collection) removeDocumentHeads(
	ctx context.Context,
	systemstore corekv.ReaderWriter,
	docShortID uint64,
	pruneBlocks bool,
	removedOwnerKeys map[string]struct{},
) error {
	headstore := datastore.NewMultistore(c.db.rootstore, c.db.lockSet, c.db.blockStoreChunkSize).Headstore()
	docIDs, err := id.GetDocIDAliasesFromStore(ctx, systemstore, docShortID)
	if err != nil {
		return err
	}

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
			err = c.deleteBlocks(ctx, systemstore, docIDs, key.Cid, pruneBlocks, removedOwnerKeys)
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
	collectionShortID uint32,
) error {
	headstore := datastore.NewMultistore(c.db.rootstore, c.db.lockSet, c.db.blockStoreChunkSize).Headstore()

	// If there are more keys than we wish to load into memory at once, this will be set to
	// true, and we'll continue the delete in another pass.
	hasMore := true

	for hasMore {
		prefix := keys.HeadstoreColKey{
			CollectionShortID: collectionShortID,
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
			// Document blocks and owner edges were deleted in hardDeleteDocKeysAndHeadstore,
			// so this remaining collection DAG can be removed unconditionally.
			err = c.deleteBlocks(ctx, nil, nil, key.Cid, true, nil)
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

// deleteBlocks releases the document owners across the reachable DAG and deletes ownerless blocks.
// A nil systemstore deletes the DAG unconditionally. Missing blocks are ignored.
func (c *collection) deleteBlocks(
	ctx context.Context,
	systemstore corekv.ReaderWriter,
	docIDs []string,
	currentCid cid.Cid,
	pruneBlocks bool,
	removedOwnerKeys map[string]struct{},
) error {
	multistore := datastore.NewMultistore(c.db.rootstore, c.db.lockSet, c.db.blockStoreChunkSize)
	blockstore := multistore.Blockstore()
	encstore := multistore.Encstore()

	deleteBlockMapping := func(blockCID cid.Cid) (bool, error) {
		if systemstore == nil {
			return true, nil
		}
		for _, docID := range docIDs {
			if err := id.DeleteBlockDocIDMapping(ctx, systemstore, blockCID, docID); err != nil {
				return false, err
			}
			if removedOwnerKeys != nil {
				key := keys.NewBlockCIDToDocIDKey(blockCID.String(), docID).Bytes()
				removedOwnerKeys[string(key)] = struct{}{}
			}
		}
		hasOwners, err := id.BlockHasOwnersExcept(ctx, systemstore, blockCID, removedOwnerKeys)
		if err != nil {
			return false, err
		}
		return !hasOwners, nil
	}

	type block struct {
		id        cid.Cid
		block     *coreblock.Block
		canDelete bool
	}
	type visit struct {
		id       cid.Cid
		expanded bool
		block    *block
	}

	stack := []visit{{id: currentCid}}
	seen := make(map[string]struct{})
	postorder := make([]*block, 0)
	encryptionToDelete := make(map[string]cid.Cid)
	seenEncryption := make(map[string]struct{})

	for len(stack) > 0 {
		last := len(stack) - 1
		current := stack[last]
		stack = stack[:last]
		if current.expanded {
			postorder = append(postorder, current.block)
			continue
		}

		key := current.id.String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		coreBlock, found, err := getBlock(ctx, blockstore, current.id)
		if err != nil {
			return err
		}
		if !found {
			if _, err := deleteBlockMapping(current.id); err != nil {
				return err
			}
			continue
		}

		currentBlock := &block{id: current.id, block: coreBlock}
		currentBlock.canDelete, err = deleteBlockMapping(current.id)
		if err != nil {
			return err
		}
		if coreBlock.Encryption != nil {
			encryptionCID := coreBlock.Encryption.Cid
			encryptionKey := encryptionCID.String()
			if _, ok := seenEncryption[encryptionKey]; !ok {
				seenEncryption[encryptionKey] = struct{}{}
				canDelete, err := deleteBlockMapping(encryptionCID)
				if err != nil {
					return err
				}
				if canDelete {
					encryptionToDelete[encryptionKey] = encryptionCID
				}
			}
		}

		stack = append(stack, visit{expanded: true, block: currentBlock})
		links := coreBlock.AllLinks()
		for i := len(links) - 1; i >= 0; i-- {
			stack = append(stack, visit{id: links[i].Cid})
		}
	}

	if !pruneBlocks {
		return nil
	}

	for _, encryptionCID := range encryptionToDelete {
		if err := deleteBlockIfPresent(ctx, encstore, encryptionCID); err != nil {
			return err
		}
	}
	for _, currentBlock := range postorder {
		if !currentBlock.canDelete {
			continue
		}
		// Signature blocks inherit their parent block's ownership.
		if currentBlock.block.Signature != nil {
			if err := deleteBlockIfPresent(ctx, blockstore, currentBlock.block.Signature.Cid); err != nil {
				return err
			}
		}
		if err := deleteBlockIfPresent(ctx, blockstore, currentBlock.id); err != nil {
			return err
		}
	}

	return nil
}

func deleteBlockIfPresent(ctx context.Context, blockstore datastore.Blockstore, blockID cid.Cid) error {
	err := blockstore.DeleteBlock(ctx, blockID)
	if errors.Is(err, corekv.ErrNotFound) || errors.Is(err, ipld.ErrNotFound{}) {
		return nil
	}
	return err
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
