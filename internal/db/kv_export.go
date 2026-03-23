package db

import (
	"context"
	"encoding/binary"
	"io"
	"strconv"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// ExportFieldMapping returns the field mapping for the given collection.
// The caller should serialize this mapping alongside the KV snapshot so that
// the importing instance can remap short IDs.
func (db *DB) ExportFieldMapping(ctx context.Context, collectionName string) (*client.CollectionFieldMapping, error) {
	ctx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard()

	col, err := db.getCollectionByName(ctx, collectionName)
	if err != nil {
		return nil, err
	}

	shortID, err := id.GetShortCollectionID(ctx, col.CollectionID())
	if err != nil {
		return nil, err
	}

	mapping := &client.CollectionFieldMapping{
		CollectionID:      col.CollectionID(),
		CollectionShortID: shortID,
		FieldIDMapping:    make(map[uint32]string),
	}

	for _, field := range col.Version().Fields {
		if field.FieldID == "" {
			continue
		}
		fieldShortID, err := id.GetShortFieldID(ctx, shortID, field.FieldID)
		if err != nil {
			return nil, err
		}
		mapping.FieldIDMapping[fieldShortID] = field.FieldID
	}

	return mapping, nil
}

// ExportDocKVs writes rootstore KV pairs for the given docIDs to w
// in length-prefixed binary format: [key_len u32 BE][key][value_len u32 BE][value].
// If datastoreOnly is true, headstore and blockstore keys are skipped.
// Returns the number of KV pairs written.
func (db *DB) ExportDocKVs(ctx context.Context, collectionName string, docIDs []string, w io.Writer, datastoreOnly bool) (int, error) {
	// Read-only txn; initializes context caches for short ID lookups.
	ctx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return 0, err
	}
	defer txn.Discard()

	col, err := db.getCollectionByName(ctx, collectionName)
	if err != nil {
		return 0, err
	}

	shortID, err := id.GetShortCollectionID(ctx, col.CollectionID())
	if err != nil {
		return 0, err
	}

	version := col.Version()
	fieldIDs := make([]string, 0, len(version.Fields))
	for _, field := range version.Fields {
		if field.FieldID != "" {
			shortFieldID, err := id.GetShortFieldID(ctx, shortID, field.FieldID)
			if err != nil {
				return 0, err
			}
			fieldIDs = append(fieldIDs, strconv.FormatUint(uint64(shortFieldID), 10))
		}
	}

	rootstore := db.Rootstore()
	count := 0

	for _, docIDStr := range docIDs {
		if ctx.Err() != nil {
			return count, ctx.Err()
		}

		n, err := exportOneDoc(ctx, rootstore, w, shortID, docIDStr, fieldIDs, datastoreOnly)
		if err != nil {
			return count, err
		}
		count += n
	}

	return count, nil
}

// exportOneDoc writes all rootstore KV pairs for a single document.
func exportOneDoc(
	ctx context.Context,
	rootstore corekv.TxnStore,
	w io.Writer,
	shortID uint32,
	docIDStr string,
	fieldIDs []string,
	datastoreOnly bool,
) (int, error) {
	count := 0

	// --- Datastore keys (namespace prefix 'd') ---

	baseKey := keys.DataStoreKey{
		CollectionShortID: shortID,
		DocID:             docIDStr,
	}

	// Value fields: /colID/v/docID/fieldID
	valueKey := baseKey.WithValueFlag()
	for _, fid := range fieldIDs {
		n, err := exportExactKey(ctx, rootstore, w, prependNamespace('d', valueKey.WithFieldID(fid).Bytes()))
		if err != nil {
			return count, err
		}
		count += n
	}

	// Version metadata: /colID/v/docID/v
	n, err := exportExactKey(ctx, rootstore, w, prependNamespace('d', valueKey.WithFieldID(keys.DATASTORE_DOC_VERSION_FIELD_ID).Bytes()))
	if err != nil {
		return count, err
	}
	count += n

	// Value key without field (base value key): /colID/v/docID
	n, err = exportExactKey(ctx, rootstore, w, prependNamespace('d', valueKey.Bytes()))
	if err != nil {
		return count, err
	}
	count += n

	// Priority key: /colID/p/docID
	n, err = exportExactKey(ctx, rootstore, w, prependNamespace('d', baseKey.WithPriorityFlag().Bytes()))
	if err != nil {
		return count, err
	}
	count += n

	// Deleted key: /colID/d/docID
	n, err = exportExactKey(ctx, rootstore, w, prependNamespace('d', baseKey.WithDeletedFlag().Bytes()))
	if err != nil {
		return count, err
	}
	count += n

	// Primary key: /colID/pk/docID
	primaryKey := keys.PrimaryDataStoreKey{
		CollectionShortID: shortID,
		DocID:             docIDStr,
	}
	n, err = exportExactKey(ctx, rootstore, w, prependNamespace('d', primaryKey.Bytes()))
	if err != nil {
		return count, err
	}
	count += n

	if datastoreOnly {
		return count, nil
	}

	// --- Headstore keys (namespace 'h') ---

	hsPrefix := keys.HeadstoreDocKey{DocID: docIDStr}
	hsPrefixBytes := prependNamespace('h', hsPrefix.Bytes())

	hsIter, err := rootstore.Iterator(ctx, corekv.IterOptions{Prefix: hsPrefixBytes})
	if err != nil {
		return count, err
	}

	var cidBytes [][]byte
	for {
		hasNext, err := hsIter.Next()
		if err != nil {
			_ = hsIter.Close()
			return count, err
		}
		if !hasNext {
			break
		}

		key := hsIter.Key()
		value, err := hsIter.Value()
		if err != nil {
			_ = hsIter.Close()
			return count, err
		}

		if err := writeKVEntry(w, key, value); err != nil {
			_ = hsIter.Close()
			return count, err
		}
		count++

		// Extract CID from headstore key for blockstore export.
		hsKeyStr := string(key[1:]) // strip 'h' namespace prefix
		hsKey, err := keys.NewHeadstoreDocKey(hsKeyStr)
		if err == nil && hsKey.Cid.Defined() {
			cidBytes = append(cidBytes, hsKey.Cid.Bytes())
		}
	}
	_ = hsIter.Close()

	// --- Blockstore keys (namespace 'b') ---

	for _, cb := range cidBytes {
		blockPrefix := prependNamespace('b', cb)
		n, err := exportByPrefix(ctx, rootstore, w, blockPrefix)
		if err != nil {
			return count, err
		}
		count += n
	}

	return count, nil
}

// exportExactKey reads a single key from the store and writes it if it exists.
func exportExactKey(ctx context.Context, store corekv.TxnStore, w io.Writer, key []byte) (int, error) {
	value, err := store.Get(ctx, key)
	if err != nil {
		// Key doesn't exist — that's fine, skip it
		return 0, nil
	}

	if err := writeKVEntry(w, key, value); err != nil {
		return 0, err
	}
	return 1, nil
}

// exportByPrefix iterates all keys with the given prefix and writes them.
func exportByPrefix(ctx context.Context, store corekv.TxnStore, w io.Writer, prefix []byte) (int, error) {
	iter, err := store.Iterator(ctx, corekv.IterOptions{Prefix: prefix})
	if err != nil {
		return 0, err
	}
	defer iter.Close() //nolint:errcheck

	count := 0
	for {
		hasNext, err := iter.Next()
		if err != nil {
			return count, err
		}
		if !hasNext {
			break
		}

		key := iter.Key()
		value, err := iter.Value()
		if err != nil {
			return count, err
		}

		if err := writeKVEntry(w, key, value); err != nil {
			return count, err
		}
		count++
	}

	return count, nil
}

// writeKVEntry writes a single key-value pair in binary format.
func writeKVEntry(w io.Writer, key, value []byte) error {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(len(key)))
	if _, err := w.Write(buf[:]); err != nil {
		return err
	}
	if _, err := w.Write(key); err != nil {
		return err
	}
	binary.BigEndian.PutUint32(buf[:], uint32(len(value)))
	if _, err := w.Write(buf[:]); err != nil {
		return err
	}
	if _, err := w.Write(value); err != nil {
		return err
	}
	return nil
}

// prependNamespace prepends a single namespace byte to a key.
func prependNamespace(ns byte, key []byte) []byte {
	result := make([]byte, 1+len(key))
	result[0] = ns
	copy(result[1:], key)
	return result
}
