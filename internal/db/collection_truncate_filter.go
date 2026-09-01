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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

type truncateTarget struct {
	docID      client.DocID
	docShortID uint64
}

func (c *collection) truncateWithFilter(
	ctx context.Context,
	lockTxn datastore.Txn,
	filter any,
) error {
	multistore := datastore.NewMultistore(c.db.rootstore, c.db.lockSet, c.db.blockStoreChunkSize)
	systemstore := multistore.Systemstore()
	shortID, err := id.GetUncachedCollectionShortID(ctx, c.def.CollectionID, systemstore)
	if err != nil {
		return err
	}

	for {
		docIDs, err := c.matchingDocIDs(ctx, lockTxn.ID(), filter)
		if err != nil {
			return err
		}
		if len(docIDs) == 0 {
			return nil
		}
		targets, aliases, err := c.resolveTruncateTargets(ctx, systemstore, shortID, docIDs)
		if err != nil {
			return err
		}
		if err := c.hardDeleteSearchableEncryption(ctx, shortID, aliases); err != nil {
			return err
		}

		// History may exceed a backend transaction limit, so process it under the collection lock.
		for _, target := range targets {
			if err := c.removeDocumentHeads(
				ctx,
				systemstore,
				target.docShortID,
			); err != nil {
				return err
			}
		}

		for _, target := range targets {
			if err := c.truncateLogicalDocument(ctx, lockTxn.ID(), shortID, target); err != nil {
				return err
			}
		}
	}
}

func (c *collection) matchingDocIDs(
	ctx context.Context,
	txnID uint64,
	filter any,
) ([]client.DocID, error) {
	basicTxn := datastore.NewTxnFrom(
		c.db.rootstore,
		c.db.lockSet,
		txnID,
		true,
		c.db.blockStoreChunkSize,
	)
	txn := wrapDatastoreTxn(basicTxn, c.db)
	ctx = InitContext(ctx, txn)
	defer txn.Discard()

	selectionPlan, err := c.makeUnpermissionedSelectionPlan(ctx, filter, true)
	if err != nil {
		return nil, err
	}
	if err := selectionPlan.Init(); err != nil {
		return nil, err
	}
	if err := selectionPlan.Start(); err != nil {
		return nil, err
	}

	docIDs := make([]client.DocID, 0, hardDeleteChunkSize)
	for {
		next, err := selectionPlan.Next()
		if err != nil {
			return nil, errors.Join(err, selectionPlan.Close())
		}
		if !next {
			break
		}
		doc := selectionPlan.Value()
		docID, err := client.NewDocIDFromString(doc.GetID())
		if err != nil {
			return nil, errors.Join(err, selectionPlan.Close())
		}
		docIDs = append(docIDs, docID)
		if len(docIDs) == hardDeleteChunkSize {
			break
		}
	}
	if err := selectionPlan.Close(); err != nil {
		return nil, err
	}
	return docIDs, nil
}

func (c *collection) truncateLogicalDocument(
	ctx context.Context,
	txnID uint64,
	shortID uint32,
	target truncateTarget,
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

	if err := c.truncateOneDoc(ctx, shortID, target); err != nil {
		return err
	}
	return txn.Commit()
}

func (c *collection) resolveTruncateTargets(
	ctx context.Context,
	systemstore corekv.Reader,
	shortID uint32,
	docIDs []client.DocID,
) ([]truncateTarget, map[string]struct{}, error) {
	targets := make([]truncateTarget, 0, len(docIDs))
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
		targets = append(targets, truncateTarget{docID: docID, docShortID: docShortID})

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

func (c *collection) truncateOneDoc(
	ctx context.Context,
	shortID uint32,
	target truncateTarget,
) error {
	// Index entries depend on field values, so remove them before document data.
	if err := c.truncateIndexedDocWithID(ctx, target.docID); err != nil {
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
