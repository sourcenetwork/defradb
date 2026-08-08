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
	"github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/utils"
)

type purgeTarget struct {
	docID      client.DocID
	docShortID uint64
}

// PurgeDocuments permanently removes local state for the given documents.
// When pruneHistory is true, it also removes reachable blocks with no other owner.
//
// Without a caller transaction, the operation is resumable and commits one document at a time.
// With one, the caller owns the commit and the full purge must fit in that transaction.
func (db *DB) PurgeDocuments(
	ctx context.Context,
	collectionName client.CollectionName,
	docIDs []client.DocID,
	pruneHistory bool,
	opts ...options.Enumerable[options.PurgeDocumentsOptions],
) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)
	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodePurgeDocumentPerm); err != nil {
		return err
	}
	if opt.Identity.HasValue() {
		ctx = identity.WithContext(ctx, opt.Identity)
	}

	_, hasCallerTxn := datastore.CtxTryGetTxn(ctx)
	if hasCallerTxn {
		ctx, txn, err := ensureContextTxn(ctx, db, false)
		if err != nil {
			return err
		}
		defer txn.Discard()

		col, err := db.getCollectionByName(ctx, collectionName)
		if err != nil {
			return err
		}
		if pruneHistory && col.Version().IsBranchable {
			return ErrCannotPruneBranchableCollection
		}
		return col.(*collection).purgeInCallerTxn(ctx, txn, docIDs, pruneHistory) //nolint:forcetypeassert
	}

	lookupCtx, lookupTxn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return err
	}
	col, err := db.getCollectionByName(lookupCtx, collectionName)
	lookupTxn.Discard()
	if err != nil {
		return err
	}
	if pruneHistory && col.Version().IsBranchable {
		return ErrCannotPruneBranchableCollection
	}

	ctx, lockTxn, err := ensureContextTxnShim(ctx, db)
	if err != nil {
		return err
	}
	defer lockTxn.Discard()
	return col.(*collection).purgeWithoutCallerTxn( //nolint:forcetypeassert
		ctx,
		lockTxn,
		docIDs,
		pruneHistory,
	)
}

func (c *collection) purgeWithoutCallerTxn(
	ctx context.Context,
	lockTxn datastore.Txn,
	docIDs []client.DocID,
	pruneHistory bool,
) error {
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

	// History can exceed a backend transaction limit for a single document. Process it directly
	// under the collection lock, then remove each document's logical state transactionally.
	for _, target := range targets {
		if err := c.removeDocumentHeads(
			ctx,
			systemstore,
			target.docShortID,
			pruneHistory,
			nil,
		); err != nil {
			return err
		}
	}

	for _, target := range targets {
		if err := c.purgeLogicalDocument(ctx, lockTxn.ID(), shortID, target); err != nil {
			return err
		}
	}

	return lockTxn.Commit()
}

func (c *collection) purgeLogicalDocument(
	ctx context.Context,
	txnID uint64,
	shortID uint32,
	target purgeTarget,
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

	if err := c.purgeOneDoc(ctx, shortID, target); err != nil {
		return err
	}
	return txn.Commit()
}

func (c *collection) purgeInCallerTxn(
	ctx context.Context,
	txn *Txn,
	docIDs []client.DocID,
	pruneHistory bool,
) error {
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

	var removedOwnerKeys map[string]struct{}
	if pruneHistory {
		removedOwnerKeys = make(map[string]struct{})
	}
	for _, target := range targets {
		if err := c.removeDocumentHeads(
			ctx,
			txn.Systemstore(),
			target.docShortID,
			pruneHistory,
			removedOwnerKeys,
		); err != nil {
			return err
		}
		if err := c.purgeOneDoc(ctx, shortID, target); err != nil {
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
) error {
	// Index entries depend on field values, so remove them before document data.
	if err := c.purgeIndexedDocWithID(ctx, target.docID); err != nil {
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
	return c.hardDeleteSearchableEncryptionInChunks(ctx, shortID, wanted, hardDeleteChunkSize)
}

func (c *collection) hardDeleteSearchableEncryptionInChunks(
	ctx context.Context,
	shortID uint32,
	wanted map[string]struct{},
	chunkSize int,
) error {
	if len(wanted) == 0 {
		return nil
	}

	ds := datastore.NewMultistore(c.db.rootstore, c.db.lockSet, c.db.blockStoreChunkSize).Datastore()
	_, transactional := corekv.TryGetCtxTxn(ctx)
	prefix := keys.DatastoreSE{CollectionShortID: shortID}
	var resumeAt *keys.DatastoreSE
	for {
		iterOpts := datastore.IterOptions{Prefix: &prefix, KeysOnly: true}
		if resumeAt != nil {
			iterOpts = datastore.IterOptions{
				Start:    resumeAt,
				End:      prefix.PrefixEnd(),
				KeysOnly: true,
			}
		}
		iter, err := ds.Iterator(ctx, iterOpts)
		if err != nil {
			return err
		}

		keysToDelete := make([]keys.DatastoreSE, 0, chunkSize)
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
				if !transactional && len(keysToDelete) == chunkSize {
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
		resume := keysToDelete[len(keysToDelete)-1]
		resumeAt = &resume
	}
}
