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
	"encoding/binary"
	"encoding/json"
	"io"
	"strconv"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// kvRemapTable holds the short-ID translations built from a kvFieldMappingFile.
type kvRemapTable struct {
	srcCollectionShortID uint32
	dstCollectionShortID uint32
	fieldShortIDs        map[uint32]uint32 // src local field short ID → dst local field short ID
}

const kvImportBatchSize = 1_000

// ImportRawKVs reads length-prefixed KV pairs from r (as written by ExportDocKVs) and
// writes them to the database rootstore in batches of 1 000.
// Returns the total number of pairs imported.
func (db *DB) ImportRawKVs(ctx context.Context, r io.Reader) (int, error) {
	total := 0
	batch := make(map[string][]byte, kvImportBatchSize)

	flush := func() error {
		ctx, txn, err := ensureContextTxn(ctx, db, false)
		if err != nil {
			return err
		}
		defer txn.Discard()

		for k, v := range batch {
			if err := txn.Rootstore().Set(ctx, []byte(k), v); err != nil {
				return err
			}
		}
		if err := txn.Commit(); err != nil {
			return err
		}
		clear(batch)
		return nil
	}

	for {
		var keyLen uint32
		if err := binary.Read(r, binary.BigEndian, &keyLen); err != nil {
			if err == io.EOF {
				break
			}
			return total, err
		}
		if keyLen == 0 {
			break // EOF sentinel
		}

		key := make([]byte, keyLen)
		if _, err := io.ReadFull(r, key); err != nil {
			return total, err
		}

		var valLen uint32
		if err := binary.Read(r, binary.BigEndian, &valLen); err != nil {
			return total, err
		}

		val := make([]byte, valLen)
		if _, err := io.ReadFull(r, val); err != nil {
			return total, err
		}

		batch[string(key)] = val
		total++

		if len(batch) >= kvImportBatchSize {
			if err := flush(); err != nil {
				return total, err
			}
		}
	}

	if len(batch) > 0 {
		if err := flush(); err != nil {
			return total, err
		}
	}

	return total, nil
}

// ImportRawKVsWithMapping reads length-prefixed KV pairs from r (as written by ExportDocKVs)
// and writes them to the rootstore, remapping datastore collection and field short IDs
// according to mappingJSON (as produced by ExportFieldMapping on the source node).
func (db *DB) ImportRawKVsWithMapping(ctx context.Context, r io.Reader, mappingJSON []byte) (int, error) {
	var mapping kvFieldMappingFile
	if err := json.Unmarshal(mappingJSON, &mapping); err != nil {
		return 0, err
	}

	// Resolve destination short IDs via a single read txn.
	txCtx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return 0, err
	}

	col, err := db.getCollectionByName(txCtx, mapping.CollectionName)
	if err != nil {
		txn.Discard()
		return 0, err
	}
	dstColShortID, err := id.GetUncachedCollectionShortID(txCtx, col.CollectionID(), db.Multistore().Systemstore())
	if err != nil {
		txn.Discard()
		return 0, err
	}

	fieldRemap := make(map[uint32]uint32, len(mapping.Fields))
	for _, f := range mapping.Fields {
		dstShortID, err := id.GetShortFieldID(txCtx, dstColShortID, f.FieldID)
		if err != nil {
			txn.Discard()
			return 0, err
		}
		if dstShortID != 0 {
			fieldRemap[f.ShortID] = dstShortID
		}
	}
	txn.Discard()

	remap := kvRemapTable{
		srcCollectionShortID: mapping.CollectionShortID,
		dstCollectionShortID: dstColShortID,
		fieldShortIDs:        fieldRemap,
	}
	return db.importRawKVsWithRemap(ctx, r, remap)
}

func (db *DB) importRawKVsWithRemap(ctx context.Context, r io.Reader, remap kvRemapTable) (int, error) {
	total := 0
	batch := make(map[string][]byte, kvImportBatchSize)

	flush := func() error {
		ctx, txn, err := ensureContextTxn(ctx, db, false)
		if err != nil {
			return err
		}
		defer txn.Discard()
		for k, v := range batch {
			if err := txn.Rootstore().Set(ctx, []byte(k), v); err != nil {
				return err
			}
		}
		if err := txn.Commit(); err != nil {
			return err
		}
		clear(batch)
		return nil
	}

	for {
		var keyLen uint32
		if err := binary.Read(r, binary.BigEndian, &keyLen); err != nil {
			if err == io.EOF {
				break
			}
			return total, err
		}
		if keyLen == 0 {
			break
		}

		rawKey := make([]byte, keyLen)
		if _, err := io.ReadFull(r, rawKey); err != nil {
			return total, err
		}

		var valLen uint32
		if err := binary.Read(r, binary.BigEndian, &valLen); err != nil {
			return total, err
		}
		val := make([]byte, valLen)
		if _, err := io.ReadFull(r, val); err != nil {
			return total, err
		}

		// Remap datastore keys that belong to the source collection.
		if len(rawKey) > 0 && rawKey[0] == kvNsDatastore {
			if dsKey, err := keys.DecodeDataStoreKey(rawKey[1:]); err == nil &&
				dsKey.CollectionShortID == remap.srcCollectionShortID {
				dsKey.CollectionShortID = remap.dstCollectionShortID
				if srcFieldID, ferr := dsKey.FieldIDAsUint(); ferr == nil {
					if dstFieldID, ok := remap.fieldShortIDs[srcFieldID]; ok {
						dsKey = dsKey.WithFieldID(strconv.Itoa(int(dstFieldID)))
					}
				}
				rawKey = append([]byte{kvNsDatastore}, dsKey.Bytes()...)
			}
		}

		batch[string(rawKey)] = val
		total++

		if len(batch) >= kvImportBatchSize {
			if err := flush(); err != nil {
				return total, err
			}
		}
	}

	if len(batch) > 0 {
		if err := flush(); err != nil {
			return total, err
		}
	}

	return total, nil
}

// RebuildCollectionIndexes drops and recreates all secondary indexes for the named
// collection. Should be called after ImportRawKVs to restore correct index state.
func (db *DB) RebuildCollectionIndexes(ctx context.Context, collectionName string) error {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	col, err := db.getCollectionByName(ctx, collectionName)
	if err != nil {
		return err
	}

	idxResults, err := col.ListIndexes(ctx)
	if err != nil {
		return err
	}

	// Collect existing index descriptions before any modification.
	type savedIndex struct {
		name   string
		fields []client.IndexedFieldDescription
	}
	saved := make([]savedIndex, 0, len(idxResults))
	for _, r := range idxResults {
		desc := r.Description
		saved = append(saved, savedIndex{name: desc.Name, fields: desc.Fields})
	}

	// Drop existing indexes.
	for _, s := range saved {
		if err := col.DeleteIndex(ctx, s.name); err != nil {
			return err
		}
	}

	// Recreate each index (this triggers a full backfill from imported documents).
	for _, s := range saved {
		if _, err := col.NewIndex(ctx, client.NewIndexRequest{
			Name:   s.name,
			Fields: s.fields,
		}); err != nil {
			return err
		}
	}

	return txn.Commit()
}
