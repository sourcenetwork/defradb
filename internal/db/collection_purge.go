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

// Keep implicit purge transactions below the store's size limit.
const purgeChunkSize = 100

// PurgeByDocIDs permanently removes local state for the given documents.
// When pruneHistory is true, it also removes reachable blocks with no other owner.
//
// Without a caller transaction, each chunk commits independently. With one, the caller owns
// the commit and the full purge must fit in that transaction.
func (c *collection) PurgeByDocIDs(
	ctx context.Context,
	docIDs []client.DocID,
	pruneHistory bool,
	opts ...options.Enumerable[options.TruncateCollectionOptions],
) error {
	ctx, _, hasCallerTxn := getTxnAndSetCtxForCollection(ctx, c)

	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := c.db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeTruncateCollectionPerm); err != nil {
		return err
	}

	if hasCallerTxn {
		return c.purgeChunk(ctx, docIDs, pruneHistory)
	}

	for i := 0; i < len(docIDs); i += purgeChunkSize {
		end := min(i+purgeChunkSize, len(docIDs))
		if err := c.purgeChunk(ctx, docIDs[i:end], pruneHistory); err != nil {
			return err
		}
	}

	return nil
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

	// Hide staged owner-edge deletions from snapshot-based ownership checks.
	var prunedOwners map[string]struct{}
	if pruneHistory {
		prunedOwners = make(map[string]struct{})
	}

	for _, docID := range docIDs {
		if err := c.purgeOneDoc(ctx, shortID, docID, pruneHistory, prunedOwners); err != nil {
			return err
		}
	}

	return txn.Commit()
}

func (c *collection) purgeOneDoc(
	ctx context.Context,
	shortID uint32,
	docID client.DocID,
	pruneHistory bool,
	prunedOwners map[string]struct{},
) error {
	docShortID, found, err := id.GetDocShortID(ctx, shortID, docID.String())
	if err != nil {
		return err
	}
	if !found {
		return nil
	}

	// Index entries are keyed by the document's field values, so they must be deleted before
	// the datastore prefixes holding those values are removed.
	if err := c.deleteIndexedDocWithID(ctx, docID, true); err != nil {
		return err
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
		if err := c.hardDeleteDocumentBlocks(ctx, systemstore, docShortID, prunedOwners); err != nil {
			return err
		}
	} else {
		if err := c.hardDeleteHeadstoreForDoc(ctx, docShortID); err != nil {
			return err
		}
	}

	return id.DeleteDocIDMappings(ctx, systemstore, docShortID)
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
