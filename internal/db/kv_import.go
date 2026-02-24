package db

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"

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

// kvRemapTable holds precomputed remapping from source to host short IDs.
type kvRemapTable struct {
	// collectionRemap maps source collection short ID → host collection short ID.
	collectionRemap map[uint32]uint32
	// fieldRemap maps source collection short ID → (source field short ID → host field short ID).
	fieldRemap map[uint32]map[uint32]uint32
}

// ImportRawKVsWithMapping reads length-prefixed KV pairs from r, remaps
// collection and field short IDs using the provided source mappings, and writes
// them to the rootstore. This enables importing KV snapshots across DefraDB
// instances that may have different short ID assignments.
//
// If sourceMappings is nil or empty, behaves identically to ImportRawKVs.
func (db *DB) ImportRawKVsWithMapping(
	ctx context.Context,
	r io.Reader,
	sourceMappings []*client.CollectionFieldMapping,
) (int, error) {
	if len(sourceMappings) == 0 {
		return db.ImportRawKVs(ctx, r)
	}

	remap, err := db.buildRemapTable(ctx, sourceMappings)
	if err != nil {
		return 0, fmt.Errorf("build remap table: %w", err)
	}

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

		if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
			if err == io.EOF {
				break
			}
			txn.Discard()
			return total, fmt.Errorf("read key length: %w", err)
		}
		keyLen := binary.BigEndian.Uint32(lenBuf[:])
		if keyLen == 0 {
			break
		}

		key := make([]byte, keyLen)
		if _, err := io.ReadFull(r, key); err != nil {
			txn.Discard()
			return total, fmt.Errorf("read key: %w", err)
		}

		if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
			txn.Discard()
			return total, fmt.Errorf("read value length: %w", err)
		}
		valueLen := binary.BigEndian.Uint32(lenBuf[:])

		value := make([]byte, valueLen)
		if _, err := io.ReadFull(r, value); err != nil {
			txn.Discard()
			return total, fmt.Errorf("read value: %w", err)
		}

		// Remap short IDs in 'd' namespace keys.
		key = remapKey(key, remap)

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

	if batchCount > 0 {
		if err := txn.Commit(); err != nil {
			return total, fmt.Errorf("commit final batch: %w", err)
		}
	} else {
		txn.Discard()
	}

	return total, nil
}

// buildRemapTable constructs the short ID remap table by looking up the host's
// short IDs for the same collection/field CIDs referenced in the source mappings.
func (db *DB) buildRemapTable(ctx context.Context, sourceMappings []*client.CollectionFieldMapping) (*kvRemapTable, error) {
	ctx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard()

	remap := &kvRemapTable{
		collectionRemap: make(map[uint32]uint32),
		fieldRemap:      make(map[uint32]map[uint32]uint32),
	}

	for _, m := range sourceMappings {
		hostColShortID, err := id.GetShortCollectionID(ctx, m.CollectionID)
		if err != nil {
			return nil, fmt.Errorf("get host collection short ID for %s: %w", m.CollectionID, err)
		}
		remap.collectionRemap[m.CollectionShortID] = hostColShortID

		fieldMap := make(map[uint32]uint32)
		for srcFieldShortID, fieldCID := range m.FieldIDMapping {
			hostFieldShortID, err := id.GetShortFieldID(ctx, hostColShortID, fieldCID)
			if err != nil {
				return nil, fmt.Errorf("get host field short ID for %s: %w", fieldCID, err)
			}
			fieldMap[srcFieldShortID] = hostFieldShortID
		}
		remap.fieldRemap[m.CollectionShortID] = fieldMap
	}

	return remap, nil
}

// remapKey rewrites collection and field short IDs in a 'd' namespace key.
// Headstore ('h') and blockstore ('b') keys are returned unchanged.
func remapKey(key []byte, remap *kvRemapTable) []byte {
	if len(key) < 2 || key[0] != 'd' {
		return key
	}

	inner := key[1:] // strip namespace byte

	// All datastore keys start with '/'
	if len(inner) == 0 || inner[0] != '/' {
		return key
	}

	// Detect key type by first byte after '/'.
	// Uvarint-encoded keys (DataStoreKey) have first byte >= 0x80 (intZero=136).
	// Decimal-string keys (PrimaryDataStoreKey) have ASCII digits (0x30-0x39).
	afterSlash := inner[1:]
	if len(afterSlash) == 0 {
		return key
	}

	if afterSlash[0] >= 0x80 {
		// DataStoreKey with uvarint-encoded collection short ID.
		return remapDataStoreKey(key, inner, remap)
	}
	if afterSlash[0] >= '0' && afterSlash[0] <= '9' {
		// PrimaryDataStoreKey with decimal-string collection short ID.
		return remapPrimaryDataStoreKey(key, inner, remap)
	}

	return key
}

// remapDataStoreKey remaps a DataStoreKey by parsing and re-encoding with host IDs.
// inner is the key bytes with the 'd' namespace stripped.
func remapDataStoreKey(original, inner []byte, remap *kvRemapTable) []byte {
	parsed, err := keys.DecodeDataStoreKey(inner)
	if err != nil {
		return original
	}

	hostColID, ok := remap.collectionRemap[parsed.CollectionShortID]
	if !ok {
		return original
	}

	srcColID := parsed.CollectionShortID
	parsed.CollectionShortID = hostColID

	// Remap numeric field short ID if present.
	if parsed.FieldID != "" && parsed.FieldID != keys.DATASTORE_DOC_VERSION_FIELD_ID {
		srcFieldID, err := strconv.ParseUint(parsed.FieldID, 10, 32)
		if err == nil {
			if fieldMap, ok := remap.fieldRemap[srcColID]; ok {
				if hostFieldID, ok := fieldMap[uint32(srcFieldID)]; ok {
					parsed.FieldID = strconv.FormatUint(uint64(hostFieldID), 10)
				}
			}
		}
	}

	return prependNamespace('d', keys.EncodeDataStoreKey(&parsed))
}

// remapPrimaryDataStoreKey remaps a PrimaryDataStoreKey (decimal-string format).
// inner is the key bytes with the 'd' namespace stripped.
func remapPrimaryDataStoreKey(original, inner []byte, remap *kvRemapTable) []byte {
	// Format: /<decimal_colShortID>/pk/<docID>
	s := string(inner)
	if !strings.HasPrefix(s, "/") {
		return original
	}
	s = s[1:] // strip leading '/'

	// Find the collection short ID (decimal digits before "/pk")
	pkIdx := strings.Index(s, "/pk")
	if pkIdx < 0 {
		return original
	}

	srcColStr := s[:pkIdx]
	srcColID, err := strconv.ParseUint(srcColStr, 10, 32)
	if err != nil {
		return original
	}

	hostColID, ok := remap.collectionRemap[uint32(srcColID)]
	if !ok {
		return original
	}

	// Reconstruct: /<hostColID>/pk/<docID>
	remainder := s[pkIdx:] // "/pk/..." or "/pk"
	remapped := "/" + strconv.FormatUint(uint64(hostColID), 10) + remainder
	return append([]byte{'d'}, []byte(remapped)...)
}

// indexRebuildBatchSize is the number of documents per transaction when rebuilding indexes.
const indexRebuildBatchSize = 2000

// RebuildCollectionIndexes drops and rebuilds all indexes for the named collection.
// Must be called after ImportRawKVs since raw KV import skips index entries.
func (db *DB) RebuildCollectionIndexes(ctx context.Context, collectionName string) error {
	// Get collection and check if it has indexes.
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

	// Remove all existing index entries.
	{
		writeTxnCtx, writeTxn, err := ensureContextTxn(ctx, db, false)
		if err != nil {
			return err
		}
		for _, index := range col.indexes {
			if err := index.RemoveAll(writeTxnCtx); err != nil {
				writeTxn.Discard()
				return fmt.Errorf("remove index %s: %w", index.Name(), err)
			}
		}
		if err := writeTxn.Commit(); err != nil {
			return fmt.Errorf("commit index removal: %w", err)
		}
	}

	// Single-pass scan of all docs, rebuild indexes in batched write txns.
	// Reuse the same collection/index objects — they are transaction-agnostic
	// (index.Save reads the transaction from context).
	txnCtx, txn, err = ensureContextTxn(ctx, db, true)
	if err != nil {
		return err
	}
	defer txn.Discard()

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
			if err := saveIndexBatch(ctx, db, col.indexes, batch); err != nil {
				_ = df.Close()
				return fmt.Errorf("save batch %d: %w", batchNum, err)
			}
			batch = batch[:0]
			batchNum++
		}
	}
	_ = df.Close()

	if len(batch) > 0 {
		if err := saveIndexBatch(ctx, db, col.indexes, batch); err != nil {
			return fmt.Errorf("save final batch: %w", err)
		}
	}

	return nil
}

// saveIndexBatch creates a write transaction and saves index entries for a batch of documents.
// Reuses the provided index objects to avoid re-fetching the collection each batch.
func saveIndexBatch(ctx context.Context, db *DB, indexes []CollectionIndex, docs []*client.Document) error {
	txnCtx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}

	for _, doc := range docs {
		for _, index := range indexes {
			if err := index.Save(txnCtx, doc); err != nil {
				txn.Discard()
				return fmt.Errorf("save index for %s: %w", doc.ID().String(), err)
			}
		}
	}

	return txn.Commit()
}
