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
	"strconv"
	"strings"
	"time"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// purgeBatchChunkSize controls how many documents are deleted per transaction.
const purgeBatchChunkSize = 10

// PurgeByDocIDs permanently removes all documents with the given docIDs.
func (c *collection) PurgeByDocIDs(ctx context.Context, docIDs []string, pruneHistory bool) (*client.PurgeResult, error) {
	if len(docIDs) == 0 {
		return &client.PurgeResult{Count: 0}, nil
	}

	if err := c.db.checkNodeAccess(ctx, immutable.None[acpIdentity.Identity](), acpTypes.NodeCollectionTruncatePerm); err != nil {
		return nil, err
	}

	purgedIDs := make([]string, 0, len(docIDs))

	for i := 0; i < len(docIDs); i += purgeBatchChunkSize {
		if ctx.Err() != nil {
			return &client.PurgeResult{Count: int64(len(purgedIDs)), DocIDs: purgedIDs}, ctx.Err()
		}

		end := i + purgeBatchChunkSize
		if end > len(docIDs) {
			end = len(docIDs)
		}
		chunk := docIDs[i:end]

		result, err := c.purgeChunk(ctx, chunk, pruneHistory)
		if err != nil {
			if strings.Contains(err.Error(), "transaction conflict") {
				// Fall back to per-doc for the rest (including the failed chunk)
				perDocResult, perDocErr := c.purgeDoc(ctx, docIDs[i:])
				if perDocResult != nil {
					purgedIDs = append(purgedIDs, perDocResult.DocIDs...)
				}
				return &client.PurgeResult{Count: int64(len(purgedIDs)), DocIDs: purgedIDs}, perDocErr
			}
			return &client.PurgeResult{Count: int64(len(purgedIDs)), DocIDs: purgedIDs}, err
		}
		purgedIDs = append(purgedIDs, result.DocIDs...)
	}

	return &client.PurgeResult{Count: int64(len(purgedIDs)), DocIDs: purgedIDs}, nil
}

// purgeChunk purges a chunk of documents in a single transaction.
func (c *collection) purgeChunk(ctx context.Context, docIDs []string, pruneHistory bool) (*client.PurgeResult, error) {
	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return nil, err
	}
	defer txn.Discard()

	shortID, err := id.GetShortCollectionID(ctx, c.def.CollectionID)
	if err != nil {
		return nil, err
	}

	ds := txn.Datastore()
	version := c.Version()

	// Precompute numeric short field IDs from schema
	fieldIDs := make([]string, 0, len(version.Fields))
	for _, field := range version.Fields {
		if field.FieldID != "" {
			shortFieldID, err := id.GetShortFieldID(ctx, shortID, field.FieldID)
			if err != nil {
				return nil, err
			}
			fieldIDs = append(fieldIDs, strconv.FormatUint(uint64(shortFieldID), 10))
		}
	}

	type unsafestore interface {
		Unsafe() corekv.ReaderWriter
	}
	unsafeDS, _ := ds.(unsafestore)
	rawStore := unsafeDS.Unsafe()
	purgedIDs := make([]string, 0, len(docIDs))
	hasIndexes := len(version.Indexes) > 0

	for _, docIDStr := range docIDs {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Delete index entries first
		if hasIndexes {
			docID, err := client.NewDocIDFromString(docIDStr)
			if err != nil {
				return nil, err
			}
			if err := c.deleteIndexedDocWithID(ctx, docID); err != nil {
				return nil, err
			}
		}

		baseKey := keys.DataStoreKey{
			CollectionShortID: shortID,
			DocID:             docIDStr,
		}

		// Delete value keys by constructing them from numeric short field IDs
		valueKey := baseKey.WithValueFlag()
		_ = rawStore.Delete(ctx, valueKey.Bytes())
		for _, fid := range fieldIDs {
			_ = rawStore.Delete(ctx, valueKey.WithFieldID(fid).Bytes())
		}
		_ = rawStore.Delete(ctx, valueKey.WithFieldID(keys.DATASTORE_DOC_VERSION_FIELD_ID).Bytes())

		_ = ds.Delete(ctx, baseKey.WithPriorityFlag())
		_ = ds.Delete(ctx, baseKey.WithDeletedFlag())

		primaryKey := keys.PrimaryDataStoreKey{
			CollectionShortID: shortID,
			DocID:             docIDStr,
		}
		_ = ds.Delete(ctx, primaryKey)

		// Delete headstore entries and their blocks.
		if pruneHistory {
			// Walk DAG chains to delete all blocks including previous versions.
			if err := c.hardDeleteDocumentBlocks(ctx, docIDStr); err != nil {
				return nil, err
			}
		} else {
			// Delete only head blocks referenced by headstore entries.
			headstore := txn.Headstore()
			blockstore := txn.Blockstore()
			hsPrefix := keys.HeadstoreDocKey{DocID: docIDStr}
			hsIter, err := headstore.Iterator(ctx, corekv.IterOptions{
				Prefix:   hsPrefix.Bytes(),
				KeysOnly: true,
			})
			if err != nil {
				return nil, err
			}

			var hsKeys []keys.HeadstoreDocKey
			for {
				hasNext, err := hsIter.Next()
				if err != nil {
					_ = hsIter.Close()
					return nil, err
				}
				if !hasNext {
					break
				}
				key, err := keys.NewHeadstoreDocKey(string(hsIter.Key()))
				if err != nil {
					_ = hsIter.Close()
					return nil, err
				}
				hsKeys = append(hsKeys, key)
			}
			_ = hsIter.Close()

			for _, key := range hsKeys {
				_ = headstore.Delete(ctx, key.Bytes())
				_ = blockstore.DeleteBlock(ctx, key.Cid)
			}
		}

		purgedIDs = append(purgedIDs, docIDStr)
	}

	if err := txn.Commit(); err != nil {
		return nil, err
	}

	return &client.PurgeResult{
		Count:  int64(len(purgedIDs)),
		DocIDs: purgedIDs,
	}, nil
}

// purgeDoc purges one document at a time, each in its own transaction.
func (c *collection) purgeDoc(ctx context.Context, docIDs []string) (*client.PurgeResult, error) {
	purgedIDs := make([]string, 0, len(docIDs))

	for _, docIDStr := range docIDs {
		if ctx.Err() != nil {
			return &client.PurgeResult{
				Count:  int64(len(purgedIDs)),
				DocIDs: purgedIDs,
			}, ctx.Err()
		}

		success, err := c.purgeOneDocWithRetry(ctx, docIDStr)
		if err != nil {
			log.ErrorContextE(ctx, "Purge failed for document, skipping",
				err,
				corelog.String("docID", docIDStr),
				corelog.String("collection", c.Name()))
			continue
		}
		if success {
			purgedIDs = append(purgedIDs, docIDStr)
		}
	}

	return &client.PurgeResult{
		Count:  int64(len(purgedIDs)),
		DocIDs: purgedIDs,
	}, nil
}

// purgeOneDocWithRetry purges a single document with retry logic for conflicts.
func (c *collection) purgeOneDocWithRetry(ctx context.Context, docIDStr string) (bool, error) {
	const maxRetries = 5

	for attempt := 0; attempt < maxRetries; attempt++ {
		if ctx.Err() != nil {
			return false, ctx.Err()
		}

		err := c.purgeOneDoc(ctx, docIDStr)
		if err == nil {
			return true, nil
		}

		if strings.Contains(err.Error(), "transaction conflict") {
			delay := time.Duration(10<<attempt) * time.Millisecond
			if delay > 200*time.Millisecond {
				delay = 200 * time.Millisecond
			}
			time.Sleep(delay)
			continue
		}

		return false, err
	}

	return false, nil
}

// purgeOneDoc purges a single document in its own transaction.
func (c *collection) purgeOneDoc(ctx context.Context, docIDStr string) error {
	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	shortID, err := id.GetShortCollectionID(ctx, c.def.CollectionID)
	if err != nil {
		return err
	}

	// Delete index entries first
	if len(c.Version().Indexes) > 0 {
		docID, err := client.NewDocIDFromString(docIDStr)
		if err != nil {
			return err
		}
		if err := c.deleteIndexedDocWithID(ctx, docID); err != nil {
			return err
		}
	}

	baseKey := keys.DataStoreKey{
		CollectionShortID: shortID,
		DocID:             docIDStr,
	}

	ds := txn.Datastore()

	if err := c.hardDeleteDatastorePrefix(ctx, baseKey.WithValueFlag()); err != nil {
		return err
	}

	_ = ds.Delete(ctx, baseKey.WithPriorityFlag())
	_ = ds.Delete(ctx, baseKey.WithDeletedFlag())

	primaryKey := keys.PrimaryDataStoreKey{
		CollectionShortID: shortID,
		DocID:             docIDStr,
	}
	_ = ds.Delete(ctx, primaryKey)

	if err := c.hardDeleteDocumentBlocks(ctx, docIDStr); err != nil {
		return err
	}

	return txn.Commit()
}
