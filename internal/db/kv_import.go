package db

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// kvImportBatchSize is the number of KV pairs per write transaction.
const kvImportBatchSize = 1000

// ImportRawKVs reads length-prefixed KV pairs from r and writes them directly to the rootstore.
// The format must match what ExportDocKVs produces:
//
//	[key_len uint32 BE][key bytes][value_len uint32 BE][value bytes]
//
// A key_len of 0 signals end-of-stream.
// Returns the number of KV pairs imported.
func (db *DB) ImportRawKVs(ctx context.Context, r io.Reader) (int, error) {
	rootstore := db.Rootstore()
	total := 0
	batchCount := 0

	txn := rootstore.NewTxn(false)

	commitBatch := func() error {
		if err := txn.Commit(); err != nil {
			return fmt.Errorf("commit batch at %d: %w", total, err)
		}
		txn = rootstore.NewTxn(false)
		batchCount = 0
		return nil
	}

	var lenBuf [4]byte

	for {
		if ctx.Err() != nil {
			txn.Discard()
			return total, ctx.Err()
		}

		// Read key length
		if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
			if err == io.EOF {
				break
			}
			txn.Discard()
			return total, fmt.Errorf("read key length: %w", err)
		}
		keyLen := binary.BigEndian.Uint32(lenBuf[:])
		if keyLen == 0 {
			// EOF marker
			break
		}

		// Read key
		key := make([]byte, keyLen)
		if _, err := io.ReadFull(r, key); err != nil {
			txn.Discard()
			return total, fmt.Errorf("read key: %w", err)
		}

		// Read value length
		if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
			txn.Discard()
			return total, fmt.Errorf("read value length: %w", err)
		}
		valueLen := binary.BigEndian.Uint32(lenBuf[:])

		// Read value
		value := make([]byte, valueLen)
		if _, err := io.ReadFull(r, value); err != nil {
			txn.Discard()
			return total, fmt.Errorf("read value: %w", err)
		}

		if err := txn.Set(ctx, key, value); err != nil {
			txn.Discard()
			return total, fmt.Errorf("set kv pair: %w", err)
		}
		total++
		batchCount++

		if batchCount >= kvImportBatchSize {
			if err := commitBatch(); err != nil {
				return total, err
			}
		}
	}

	// Commit remaining
	if batchCount > 0 {
		if err := txn.Commit(); err != nil {
			return total, fmt.Errorf("commit final batch: %w", err)
		}
	} else {
		txn.Discard()
	}

	return total, nil
}

// indexRebuildBatchSize is the number of documents per transaction when rebuilding indexes.
const indexRebuildBatchSize = 500

// RebuildCollectionIndexes drops and rebuilds all indexes for the named collection.
// Must be called after ImportRawKVs since raw KV import skips index entries.
func (db *DB) RebuildCollectionIndexes(ctx context.Context, collectionName string) error {
	// Phase 1: Check if the collection has indexes. If not, return early.
	{
		txnCtx, txn, err := ensureContextTxn(ctx, db, true)
		if err != nil {
			return err
		}
		colIface, err := db.getCollectionByName(txnCtx, collectionName)
		txn.Discard()
		if err != nil {
			return err
		}
		col, ok := colIface.(*collection)
		if !ok {
			return fmt.Errorf("internal: unexpected collection type for %s", collectionName)
		}
		if len(col.indexes) == 0 {
			return nil
		}
	}

	// Phase 2: Remove all existing index entries.
	{
		txnCtx, txn, err := ensureContextTxn(ctx, db, false)
		if err != nil {
			return err
		}
		colIface, err := db.getCollectionByName(txnCtx, collectionName)
		if err != nil {
			txn.Discard()
			return err
		}
		col := colIface.(*collection)
		for _, index := range col.indexes {
			if err := index.RemoveAll(txnCtx); err != nil {
				txn.Discard()
				return fmt.Errorf("remove index %s: %w", index.Name(), err)
			}
		}
		if err := txn.Commit(); err != nil {
			return fmt.Errorf("commit index removal: %w", err)
		}
	}

	// Phase 3: Single-pass scan of all docs, rebuild indexes in batched write txns.
	txnCtx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return err
	}
	defer txn.Discard()

	colIface, err := db.getCollectionByName(txnCtx, collectionName)
	if err != nil {
		return err
	}
	col := colIface.(*collection)

	// Only fetch fields that are part of an index.
	indexedFieldNames := make(map[string]struct{})
	for _, index := range col.indexes {
		for _, field := range index.Description().Fields {
			indexedFieldNames[field.Name] = struct{}{}
		}
	}
	var fields []client.CollectionFieldDescription
	for _, field := range col.Version().Fields {
		if _, ok := indexedFieldNames[field.Name]; ok {
			fields = append(fields, field)
		}
	}

	// Single fetcher for the entire collection scan.
	readTxn := datastore.CtxMustGetTxn(txnCtx)
	df := col.newFetcher(txnCtx)
	err = df.Init(
		txnCtx,
		identity.FromContext(txnCtx),
		readTxn,
		col.db.nodeACP,
		col.db.documentACP,
		immutable.None[client.IndexDescription](),
		col,
		fields,
		nil,
		nil,
		nil,
		false,
	)
	if err != nil {
		_ = df.Close()
		return fmt.Errorf("init fetcher: %w", err)
	}

	shortID, err := id.GetShortCollectionID(txnCtx, col.Version().CollectionID)
	if err != nil {
		_ = df.Close()
		return err
	}

	prefix := keys.DataStoreKey{CollectionShortID: shortID}
	if err := df.Start(txnCtx, prefix); err != nil {
		_ = df.Close()
		return fmt.Errorf("start fetcher: %w", err)
	}

	var batch []*client.Document
	batchNum := 0

	for {
		encodedDoc, _, err := df.FetchNext(txnCtx)
		if err != nil {
			_ = df.Close()
			return fmt.Errorf("fetch doc: %w", err)
		}
		if encodedDoc == nil {
			break
		}

		doc, err := fetcher.Decode(txnCtx, encodedDoc, col.Version())
		if err != nil {
			_ = df.Close()
			return fmt.Errorf("decode doc: %w", err)
		}

		batch = append(batch, doc)

		if len(batch) >= indexRebuildBatchSize {
			if err := saveIndexBatch(ctx, db, collectionName, batch); err != nil {
				_ = df.Close()
				return fmt.Errorf("save batch %d: %w", batchNum, err)
			}
			batch = batch[:0]
			batchNum++
		}
	}
	_ = df.Close()

	// Process remaining docs
	if len(batch) > 0 {
		if err := saveIndexBatch(ctx, db, collectionName, batch); err != nil {
			return fmt.Errorf("save final batch: %w", err)
		}
	}

	return nil
}

// saveIndexBatch creates a write transaction and saves index entries for a batch of documents.
func saveIndexBatch(ctx context.Context, db *DB, collectionName string, docs []*client.Document) error {
	txnCtx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}

	colIface, err := db.getCollectionByName(txnCtx, collectionName)
	if err != nil {
		txn.Discard()
		return err
	}
	col := colIface.(*collection)

	for _, doc := range docs {
		for _, index := range col.indexes {
			if err := index.Save(txnCtx, doc); err != nil {
				txn.Discard()
				return fmt.Errorf("save index for %s: %w", doc.ID().String(), err)
			}
		}
	}

	return txn.Commit()
}
