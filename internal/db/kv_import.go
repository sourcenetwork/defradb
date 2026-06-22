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
	"io"

	"github.com/sourcenetwork/defradb/client"
)

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
