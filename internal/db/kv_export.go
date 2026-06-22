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

	"github.com/ipfs/go-cid"
	"github.com/sourcenetwork/corekv"

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// Namespace bytes that are prepended to keys in the KV wire format.
const (
	kvNsDatastore  = byte('d')
	kvNsHeadstore  = byte('h')
	kvNsBlockstore = byte('b')
)

// ExportDocKVs writes rootstore KV pairs for the given docIDs in the named collection
// to w in length-prefixed binary format.  If datastoreOnly is true, headstore and
// blockstore entries are omitted.  Returns the total number of pairs written.
//
// Wire format per entry:
//
//	[key_len  uint32 big-endian]
//	[key_bytes = namespace_byte + key_from_namespaced_store]
//	[val_len  uint32 big-endian]
//	[val_bytes]
//
// Terminated by a uint32(0) sentinel.
func (db *DB) ExportDocKVs(
	ctx context.Context,
	collectionName string,
	docIDs []string,
	w io.Writer,
	datastoreOnly bool,
) (int, error) {
	ctx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return 0, err
	}
	defer txn.Discard()

	col, err := db.getCollectionByName(ctx, collectionName)
	if err != nil {
		return 0, err
	}

	shortID, err := id.GetUncachedShortCollectionID(ctx, col.CollectionID(), db.Multistore().Systemstore())
	if err != nil {
		return 0, err
	}

	total := 0
	for _, docID := range docIDs {
		n, err := exportOneDoc(ctx, txn, shortID, docID, w, datastoreOnly)
		if err != nil {
			return total, err
		}
		total += n
	}

	// EOF sentinel: key_len == 0
	return total, binary.Write(w, binary.BigEndian, uint32(0))
}

func exportOneDoc(
	ctx context.Context,
	txn *Txn,
	collectionShortID uint32,
	docID string,
	w io.Writer,
	datastoreOnly bool,
) (int, error) {
	total := 0

	// -- Data store --
	dsPrefix := keys.DataStoreKey{CollectionShortID: collectionShortID, DocID: docID}.Bytes()
	n, err := exportByPrefixRaw(ctx, datastore.NewUnsafeDatastore(txn.Rootstore()), dsPrefix, kvNsDatastore, w)
	if err != nil {
		return total, err
	}
	total += n

	if datastoreOnly {
		return total, nil
	}

	// -- Head store --
	headPrefix := keys.HeadstoreDocKey{DocID: docID}.Bytes()
	n, err = exportByPrefixRaw(ctx, txn.Headstore(), headPrefix, kvNsHeadstore, w)
	if err != nil {
		return total, err
	}
	total += n

	// -- Block store --
	n, err = exportBlocksForDoc(ctx, txn, docID, w)
	if err != nil {
		return total, err
	}
	total += n

	return total, nil
}

// exportByPrefixRaw iterates a corekv.ReaderWriter with the given prefix and writes
// each entry to w prepended with the namespace byte.
func exportByPrefixRaw(
	ctx context.Context,
	store corekv.ReaderWriter,
	prefix []byte,
	ns byte,
	w io.Writer,
) (int, error) {
	iter, err := store.Iterator(ctx, corekv.IterOptions{Prefix: prefix})
	if err != nil {
		return 0, err
	}

	count := 0
	for {
		hasNext, err := iter.Next()
		if err != nil {
			return count, err
		}
		if !hasNext {
			break
		}

		k := iter.Key()
		v, err := iter.Value()
		if err != nil {
			return count, err
		}

		if err := writeKVEntry(w, ns, k, v); err != nil {
			return count, err
		}
		count++
	}

	return count, iter.Close()
}

// exportBlocksForDoc collects all block CIDs for the document by scanning its headstore
// entries, then exports each block recursively following DAG links.
func exportBlocksForDoc(ctx context.Context, txn *Txn, docID string, w io.Writer) (int, error) {
	// Collect head CIDs from headstore.
	headPrefix := keys.HeadstoreDocKey{DocID: docID}.Bytes()
	iter, err := txn.Headstore().Iterator(ctx, corekv.IterOptions{Prefix: headPrefix})
	if err != nil {
		return 0, err
	}

	var headCIDs []cid.Cid
	for {
		hasNext, err := iter.Next()
		if err != nil {
			return 0, err
		}
		if !hasNext {
			break
		}

		hk, err := keys.NewHeadstoreDocKey(string(iter.Key()))
		if err != nil {
			continue
		}
		if hk.Cid.Defined() {
			headCIDs = append(headCIDs, hk.Cid)
		}
	}
	if err := iter.Close(); err != nil {
		return 0, err
	}

	visited := make(map[string]struct{})
	total := 0
	for _, c := range headCIDs {
		n, err := exportBlockDAG(ctx, txn.Blockstore(), c, visited, w)
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}

// exportBlockDAG recursively exports a block and all its linked blocks.
func exportBlockDAG(
	ctx context.Context,
	bstore datastore.Blockstore,
	c cid.Cid,
	visited map[string]struct{},
	w io.Writer,
) (int, error) {
	cidStr := c.String()
	if _, seen := visited[cidStr]; seen {
		return 0, nil
	}
	visited[cidStr] = struct{}{}

	blk, err := bstore.Get(ctx, c)
	if err != nil {
		return 0, err
	}

	rawData := blk.RawData()

	// The corekv blockstore stores a block at key = cid.Bytes()
	if err := writeKVEntry(w, kvNsBlockstore, c.Bytes(), rawData); err != nil {
		return 0, err
	}

	// Decode to follow links (skip if not a defradb block).
	decoded, err := coreblock.GetFromBytes(rawData)
	if err != nil {
		return 1, nil
	}

	count := 1
	for _, link := range decoded.Links {
		n, err := exportBlockDAG(ctx, bstore, link.Link.Cid, visited, w)
		if err != nil {
			return count, err
		}
		count += n
	}

	return count, nil
}

// writeKVEntry serialises one key-value entry in the wire format.
// The stored key is: [ns] + key.
func writeKVEntry(w io.Writer, ns byte, key, value []byte) error {
	fullKey := append([]byte{ns}, key...) //nolint:gocritic

	if err := binary.Write(w, binary.BigEndian, uint32(len(fullKey))); err != nil {
		return err
	}
	if _, err := w.Write(fullKey); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint32(len(value))); err != nil {
		return err
	}
	_, err := w.Write(value)
	return err
}
