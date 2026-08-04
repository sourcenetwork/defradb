// Copyright 2026 Democratized Data Foundation
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

	"github.com/sourcenetwork/corekv"

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/utils"
)

// Keep logical document cleanup below the store's transaction size limit.
const purgeChunkSize = 100

type purgeTarget struct {
	docID      client.DocID
	docShortID uint64
}

// PurgeByDocIDs permanently removes local state for the given documents.
// When pruneHistory is true, it also removes reachable blocks with no other owner.
//
// Without a caller transaction, the operation is resumable and commits logical cleanup in chunks.
// With one, the caller owns the commit and the full purge must fit in that transaction.
func (c *collection) PurgeByDocIDs(
	ctx context.Context,
	docIDs []client.DocID,
	pruneHistory bool,
	opts ...options.Enumerable[options.PurgeByDocIDsOptions],
) error {
	ctx, _, hasCallerTxn := getTxnAndSetCtxForCollection(ctx, c)

	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)
	if err := c.db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodePurgeDocumentPerm); err != nil {
		return err
	}
	if pruneHistory && c.def.IsBranchable {
		return ErrCannotPruneBranchableCollection
	}

	if hasCallerTxn {
		return c.purgeChunk(ctx, docIDs, pruneHistory)
	}
	return c.purgeWithoutCallerTxn(ctx, docIDs, pruneHistory)
}

func (c *collection) purgeWithoutCallerTxn(
	ctx context.Context,
	docIDs []client.DocID,
	pruneHistory bool,
) error {
	ctx, lockTxn, err := ensureContextTxnShim(ctx, c.db)
	if err != nil {
		return err
	}
	defer lockTxn.Discard()

	multistore := datastore.NewMultistore(c.db.rootstore, c.db.lockSet, c.db.blockStoreChunkSize)
	systemstore := multistore.Systemstore()
	shortID, err := id.GetUncachedCollectionShortID(ctx, c.def.CollectionID, systemstore)
	if err != nil {
		return err
	}
	c.db.lockSet.CollectionLock(lockTxn, shortID)

	targets, aliases, err := c.resolvePurgeTargets(ctx, systemstore, shortID, docIDs)
	if err != nil {
		return err
	}
	if err := c.hardDeleteSearchableEncryption(ctx, shortID, aliases); err != nil {
		return err
	}

	// History can exceed a backend transaction limit for a single document. Prune it directly
	// under the collection lock, then remove the bounded logical state transactionally.
	if pruneHistory {
		for _, target := range targets {
			if err := c.hardDeleteDocumentBlocks(ctx, systemstore, target.docShortID, nil); err != nil {
				return err
			}
		}
	}

	for i := 0; i < len(targets); i += purgeChunkSize {
		end := min(i+purgeChunkSize, len(targets))
		if err := c.purgeLogicalChunk(ctx, lockTxn.ID(), shortID, targets[i:end]); err != nil {
			return err
		}
	}

	return lockTxn.Commit()
}

func (c *collection) purgeLogicalChunk(
	ctx context.Context,
	txnID uint64,
	shortID uint32,
	targets []purgeTarget,
) error {
	basicTxn := datastore.NewTxnFrom(
		c.db.rootstore,
		c.db.lockSet,
		txnID,
		false,
		c.db.blockStoreChunkSize,
	)
	txn := wrapDatastoreTxn(basicTxn, c.db)
	ctx = InitContext(ctx, txn)
	defer txn.Discard()

	for _, target := range targets {
		if err := c.purgeOneDoc(ctx, shortID, target, false, nil); err != nil {
			return err
		}
	}
	return txn.Commit()
}

func (c *collection) purgeChunk(
	ctx context.Context,
	docIDs []client.DocID,
	pruneHistory bool,
) error {
	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	shortID, err := id.GetCollectionShortID(ctx, c.def.CollectionID)
	if err != nil {
		return err
	}
	c.db.lockSet.CollectionLock(txn, shortID)

	targets, aliases, err := c.resolvePurgeTargets(ctx, txn.Systemstore(), shortID, docIDs)
	if err != nil {
		return err
	}
	if err := c.hardDeleteSearchableEncryption(ctx, shortID, aliases); err != nil {
		return err
	}

	var prunedOwners map[string]struct{}
	if pruneHistory {
		prunedOwners = make(map[string]struct{})
	}
	for _, target := range targets {
		if err := c.purgeOneDoc(ctx, shortID, target, pruneHistory, prunedOwners); err != nil {
			return err
		}
	}
	return txn.Commit()
}

func (c *collection) resolvePurgeTargets(
	ctx context.Context,
	systemstore corekv.Reader,
	shortID uint32,
	docIDs []client.DocID,
) ([]purgeTarget, map[string]struct{}, error) {
	targets := make([]purgeTarget, 0, len(docIDs))
	aliases := make(map[string]struct{}, len(docIDs))
	seen := make(map[uint64]struct{}, len(docIDs))

	for _, docID := range docIDs {
		aliases[docID.String()] = struct{}{}
		docShortID, found, err := id.GetDocShortIDFromStore(ctx, systemstore, shortID, docID.String())
		if err != nil {
			return nil, nil, err
		}
		if !found {
			continue
		}
		if _, ok := seen[docShortID]; ok {
			continue
		}
		seen[docShortID] = struct{}{}
		targets = append(targets, purgeTarget{docID: docID, docShortID: docShortID})

		docAliases, err := id.GetDocIDAliasesFromStore(ctx, systemstore, docShortID)
		if err != nil {
			return nil, nil, err
		}
		for _, alias := range docAliases {
			aliases[alias] = struct{}{}
		}
	}

	return targets, aliases, nil
}

func (c *collection) purgeOneDoc(
	ctx context.Context,
	shortID uint32,
	target purgeTarget,
	pruneHistory bool,
	prunedOwners map[string]struct{},
) error {
	// Index entries depend on field values, so remove them before document data.
	if err := c.deleteIndexedDocWithID(ctx, target.docID, true); err != nil {
		return err
	}

	multistore := datastore.NewMultistore(c.db.rootstore, c.db.lockSet, c.db.blockStoreChunkSize)
	systemstore := multistore.Systemstore()
	if err := deleteDatastoreKey(ctx, multistore.Datastore(), keys.PrimaryDataStoreKey{
		CollectionShortID: shortID,
		DocShortID:        target.docShortID,
	}); err != nil {
		return err
	}

	for _, instanceType := range []keys.InstanceType{keys.ValueKey, keys.PriorityKey, keys.DeletedKey} {
		prefix := keys.DataStoreKey{
			CollectionShortID: shortID,
			InstanceType:      instanceType,
			DocShortID:        target.docShortID,
		}
		if err := c.hardDeleteDatastorePrefix(ctx, prefix); err != nil {
			return err
		}
	}

	if pruneHistory {
		if err := c.hardDeleteDocumentBlocks(ctx, systemstore, target.docShortID, prunedOwners); err != nil {
			return err
		}
	} else {
		if err := c.hardDeleteHeadstoreForDoc(ctx, target.docShortID); err != nil {
			return err
		}
	}

	return id.DeleteDocIDMappings(ctx, systemstore, target.docShortID)
}

func deleteDatastoreKey(ctx context.Context, store datastore.Keyedstore, key keys.Key) error {
	if err := store.Delete(ctx, key); err != nil && !errors.Is(err, corekv.ErrNotFound) {
		return err
	}
	return nil
}

func (c *collection) hardDeleteSearchableEncryption(
	ctx context.Context,
	shortID uint32,
	wanted map[string]struct{},
) error {
	if len(wanted) == 0 {
		return nil
	}

	ds := datastore.NewMultistore(c.db.rootstore, c.db.lockSet, c.db.blockStoreChunkSize).Datastore()
	_, transactional := corekv.TryGetCtxTxn(ctx)
	for {
		iter, err := ds.Iterator(ctx, datastore.IterOptions{
			Prefix:   keys.DatastoreSE{CollectionShortID: shortID},
			KeysOnly: true,
		})
		if err != nil {
			return err
		}

		keysToDelete := make([]keys.DatastoreSE, 0, hardDeleteChunkSize)
		hasMore := false
		for {
			hasNext, err := iter.Next()
			if err != nil {
				return errors.Join(err, iter.Close())
			}
			if !hasNext {
				break
			}
			key, err := keys.NewDatastoreSEFromString(string(iter.Key()))
			if err != nil {
				return errors.Join(err, iter.Close())
			}
			if _, ok := wanted[key.DocID]; ok {
				keysToDelete = append(keysToDelete, key)
				if !transactional && len(keysToDelete) == hardDeleteChunkSize {
					hasMore = true
					break
				}
			}
		}
		if err := iter.Close(); err != nil {
			return err
		}

		for i := range keysToDelete {
			if err := deleteDatastoreKey(ctx, ds, &keysToDelete[i]); err != nil {
				return err
			}
		}
		if !hasMore {
			return nil
		}
	}
}

func (c *collection) hardDeleteHeadstoreForDoc(ctx context.Context, docShortID uint64) error {
	headstore := datastore.NewMultistore(c.db.rootstore, c.db.lockSet, c.db.blockStoreChunkSize).Headstore()

	hasMore := true
	for hasMore {
		prefix := keys.HeadstoreDocKey{DocShortID: docShortID}
		iter, err := headstore.Iterator(ctx, corekv.IterOptions{
			Prefix:   prefix.Bytes(),
			KeysOnly: true,
		})
		if err != nil {
			return NewErrCreateTruncateIterator(err)
		}

		keysToDelete := make([][]byte, 0, hardDeleteChunkSize)
		for range hardDeleteChunkSize {
			hasNext, err := iter.Next()
			if err != nil {
				return errors.Join(err, iter.Close())
			}
			if !hasNext {
				hasMore = false
				break
			}
			keysToDelete = append(keysToDelete, append([]byte(nil), iter.Key()...))
		}

		if err := iter.Close(); err != nil {
			return err
		}
		for _, key := range keysToDelete {
			if err := headstore.Delete(ctx, key); err != nil {
				return NewErrTruncateHeadstoreKey(err, string(key))
			}
		}
	}
	return nil
}
