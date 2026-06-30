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

const purgeChunkSize = 10

// PurgeByDocIDs permanently removes all state for the given documents from this node.
// The collection is locked for the duration; no concurrent document reads or writes
// on this collection will progress while this executes.
//
// Unlike DeleteDocument (a soft-delete CRDT operation), PurgeByDocIDs performs a hard,
// irreversible removal of datastore values, headstore entries, and — when pruneHistory
// is true — all blockstore blocks reachable from each document's head CIDs.
func (c *collection) PurgeByDocIDs(
	ctx context.Context,
	docIDs []client.DocID,
	pruneHistory bool,
	opts ...options.Enumerable[options.TruncateCollectionOptions],
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

	shortID, err := id.GetCollectionShortID(ctx, c.def.CollectionID)
	if err != nil {
		return err
	}

	c.db.lockSet.CollectionLock(txn, shortID)

	for i := 0; i < len(docIDs); i += purgeChunkSize {
		end := min(i+purgeChunkSize, len(docIDs))
		if err := c.purgeChunk(ctx, shortID, docIDs[i:end], pruneHistory); err != nil {
			return err
		}
	}

	return txn.Commit()
}

func (c *collection) purgeChunk(
	ctx context.Context,
	shortID uint32,
	docIDs []client.DocID,
	pruneHistory bool,
) error {
	for _, docID := range docIDs {
		if err := c.purgeOneDoc(ctx, shortID, docID, pruneHistory); err != nil {
			return err
		}
	}
	return nil
}

func (c *collection) purgeOneDoc(
	ctx context.Context,
	shortID uint32,
	docID client.DocID,
	pruneHistory bool,
) error {
	docShortID, found, err := id.GetDocShortID(ctx, shortID, docID.String())
	if err != nil {
		return err
	}
	if !found {
		return nil
	}

	// InstanceType sits between CollectionShortID and DocShortID in the encoded key, so we
	// must include it in the prefix; a key without InstanceType matches nothing.
	for _, itype := range []keys.InstanceType{keys.ValueKey, keys.PriorityKey, keys.DeletedKey} {
		prefix := keys.DataStoreKey{
			CollectionShortID: shortID,
			InstanceType:      itype,
			DocShortID:        docShortID,
		}
		if err := c.hardDeleteDatastorePrefix(ctx, prefix); err != nil {
			return err
		}
	}

	systemstore := datastore.NewMultistore(c.db.rootstore, c.db.lockSet, c.db.blockStoreChunkSize).Systemstore()
	if pruneHistory {
		if err := c.hardDeleteDocumentBlocks(ctx, systemstore, docShortID); err != nil {
			return err
		}
	} else {
		if err := c.hardDeleteHeadstoreForDoc(ctx, docShortID); err != nil {
			return err
		}
	}

	return id.DeleteDocIDMappings(ctx, systemstore, docShortID)
}

// hardDeleteHeadstoreForDoc deletes all headstore entries for the given document
// without deleting the referenced blocks.
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

		for i := 0; i < hardDeleteChunkSize; i++ {
			hasNext, err := iter.Next()
			if err != nil {
				return errors.Join(err, iter.Close())
			}
			if !hasNext {
				hasMore = false
				break
			}
			rawKey := iter.Key()
			keyCopy := make([]byte, len(rawKey))
			copy(keyCopy, rawKey)
			keysToDelete = append(keysToDelete, keyCopy)
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
