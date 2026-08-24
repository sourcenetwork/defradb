// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package datastore

import (
	"context"
	"errors"
	"time"

	"github.com/ipfs/go-cid"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/namespace"
	"github.com/sourcenetwork/defradb/internal/db/blockowner"
)

// orphanDeleteBatchSize is how many markers share one delete transaction. A merge
// that touches any marker in a batch aborts the whole batch, so this is kept small
// enough that redoing one on a later pass is cheap, while still amortising the commit
// across many blocks.
const orphanDeleteBatchSize = 256

// SweepResult reports what one call to ReclaimOrphanBlocks did.
type SweepResult struct {
	// NextKey is where the next call resumes, nil once the index has been swept end to
	// end. It is only meaningful when the call returned no error.
	NextKey   []byte
	Scanned   int
	Reclaimed int
	// Repaired counts markers cleared from blocks a document still owns, which is a
	// marker that outlived the merge that claimed the block.
	Repaired int
	// Conflicts counts batches abandoned because a merge committed against a marker they
	// held. Their blocks stay marked and are reconsidered on the next full pass.
	Conflicts int
	// Unparsed counts markers whose key did not decode as a CID. No block can be identified
	// from such a key, so nothing is deleted for it and it is left in place.
	Unparsed int
}

// ReclaimOrphanBlocks deletes blocks that were fetched during P2P sync but never merged
// into a document, identified by a to-merge marker older than cutoff. History prune
// never reaches these blocks because it walks out from documents, so without this sweep
// they accumulate for the life of the store.
//
// It resumes from startKey, nil beginning at the start of the marker index, and scans at
// most scanLimit markers per call. The scan only nominates candidates: it collects them
// without mutating the iterator, then reclaimBatch re-decides each one under the
// transaction that deletes it.
//
// Only markers carrying a timestamp are eligible. A marker without one was written before
// the index recorded fetch times, by a store that also predates the block-owner edges this
// sweep reads, so neither the age check nor the ownership check holds for it and its block
// is left in place.
func ReclaimOrphanBlocks(
	ctx context.Context,
	rootstore corekv.TxnReaderWriter,
	cutoff time.Time,
	startKey []byte,
	scanLimit int,
) (SweepResult, error) {
	blockNS := namespace.Wrap(rootstore, []byte{blockStoreKey})

	candidates, nextKey, scanned, err := collectExpiredMarkers(ctx, blockNS, cutoff, startKey, scanLimit)
	if err != nil {
		return SweepResult{Scanned: scanned}, err
	}
	result := SweepResult{NextKey: nextKey, Scanned: scanned}

	for start := 0; start < len(candidates); start += orphanDeleteBatchSize {
		// The loop commits one batch at a time and does not stop on its own, so
		// cancellation is checked here rather than only between sweeps.
		if err := ctx.Err(); err != nil {
			return result, err
		}
		end := min(start+orphanDeleteBatchSize, len(candidates))
		if err := reclaimBatch(ctx, rootstore, cutoff, candidates[start:end], &result); err != nil {
			return result, err
		}
	}
	return result, nil
}

// reclaimBatch deletes one batch of orphaned blocks in a single transaction, re-reading
// each marker inside it so the decision to delete and the delete itself are atomic, and
// accumulates what it did into result.
//
// The re-read is what makes the sweep safe, and it rests on one invariant: a merge that
// takes ownership of a block clears that block's to-merge marker in the same transaction.
// Reading the marker here puts it in the transaction's read set, so that merge's write to
// the same key makes this commit conflict and the batch is abandoned with nothing deleted.
// The markers survive for the next full pass.
func reclaimBatch(
	ctx context.Context,
	rootstore corekv.TxnReaderWriter,
	cutoff time.Time,
	markers [][]byte,
	result *SweepResult,
) error {
	txn := rootstore.NewTxn(false)
	defer txn.Discard()

	// corekv takes the transaction from the context, so the stores below join it rather
	// than committing per call.
	txnCtx := corekv.SetCtxTxn(ctx, txn)
	blockNS := namespace.Wrap(rootstore, []byte{blockStoreKey})
	systemNS := SystemstoreFrom(rootstore)

	reclaimed, repaired := 0, 0
	for _, marker := range markers {
		value, err := blockNS.Get(txnCtx, marker)
		if errors.Is(err, corekv.ErrNotFound) {
			// A merge cleared the marker after the scan nominated it.
			continue
		}
		if err != nil {
			return err
		}
		if t, decoded := toMergeTime(value); !decoded || !t.Before(cutoff) {
			// Untimestamped, or re-written since the scan.
			continue
		}

		// A marker key is the to-merge prefix followed by the block's CID. A chunked store
		// appends a suffix to every key it writes, so the CID is read off the front rather
		// than from the whole remainder, and what follows it says whether it is chunked.
		cidLen, blockCID, err := cid.CidFromBytes(marker[1:])
		if err != nil {
			// Not a CID, so ownership cannot be checked. Leave it rather than guess.
			result.Unparsed++
			continue
		}
		chunked := len(marker)-1 > cidLen

		owned, err := blockowner.HasOwners(txnCtx, systemNS, blockCID)
		if err != nil {
			return err
		}
		if owned {
			// A committed document owns the block, so the marker is stale rather than the
			// block being garbage. Ownership decides that, not the marker's presence.
			// Drop the marker, keep the block.
			if err := deleteKey(txnCtx, blockNS, marker[:1+cidLen], chunked); err != nil {
				return err
			}
			repaired++
			continue
		}

		// Both deletes land in one commit, so a crash leaves the block and its marker
		// either both present, and reclaimable again on a later pass, or both gone.
		if err := deleteKey(txnCtx, blockNS, blockCID.Bytes(), chunked); err != nil {
			return err
		}
		if err := deleteKey(txnCtx, blockNS, marker[:1+cidLen], chunked); err != nil {
			return err
		}
		reclaimed++
	}

	if err := txn.Commit(); err != nil {
		if errors.Is(err, corekv.ErrTxnConflict) {
			result.Conflicts++
			return nil
		}
		return err
	}
	result.Reclaimed += reclaimed
	result.Repaired += repaired
	return nil
}

// collectExpiredMarkers scans the to-merge index from startKey and returns the keys of
// markers that carry a timestamp older than cutoff, along with the key to resume from
// next. These are candidates only; each is re-checked under a transaction before anything
// is deleted. Kept keys are copied because the iterator reuses its buffers across Next.
func collectExpiredMarkers(
	ctx context.Context,
	blockNS corekv.ReaderWriter,
	cutoff time.Time,
	startKey []byte,
	scanLimit int,
) (expired [][]byte, nextKey []byte, scanned int, err error) {
	iter, err := blockNS.Iterator(ctx, corekv.IterOptions{Prefix: []byte{toMergeIndexPrefix}})
	if err != nil {
		return nil, nil, 0, err
	}
	defer func() {
		if cerr := iter.Close(); cerr != nil && err == nil {
			expired, nextKey, err = nil, nil, cerr
		}
	}()

	ok, err := seekMarker(iter, startKey)
	for ok && err == nil {
		// The scan walks up to scanLimit markers and the iterator does not stop on its
		// own, so cancellation is checked here rather than only between sweeps.
		if cerr := ctx.Err(); cerr != nil {
			return nil, nil, scanned, cerr
		}
		if scanned == scanLimit {
			// The current marker is unexamined; resume from it next call.
			return expired, copyKey(iter.Key()), scanned, nil
		}
		var value []byte
		if value, err = iter.Value(); err != nil {
			break
		}
		if t, decoded := toMergeTime(value); decoded && t.Before(cutoff) {
			expired = append(expired, copyKey(iter.Key()))
		}
		scanned++
		ok, err = iter.Next()
	}
	if err != nil {
		return nil, nil, scanned, err
	}
	return expired, nil, scanned, nil
}

// seekMarker positions the iterator at the first marker to examine: the beginning of
// the index when startKey is nil, otherwise startKey, or the next marker if startKey
// was deleted since the previous call.
func seekMarker(iter corekv.Iterator, startKey []byte) (bool, error) {
	if startKey == nil {
		return iter.Next()
	}
	return iter.Seek(startKey)
}

// deleteKey removes the value stored under key. A chunked store splits one value across several
// keys sharing it as a prefix, so those are gathered before any delete: the memory store deadlocks
// if it is written to while an iterator is open.
func deleteKey(ctx context.Context, store corekv.ReaderWriter, key []byte, chunked bool) error {
	if !chunked {
		return store.Delete(ctx, key)
	}

	iter, err := store.Iterator(ctx, corekv.IterOptions{Prefix: key, KeysOnly: true})
	if err != nil {
		return err
	}

	var keys [][]byte
	for {
		hasNext, err := iter.Next()
		if err != nil {
			return errors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}
		keys = append(keys, copyKey(iter.Key()))
	}
	if err := iter.Close(); err != nil {
		return err
	}

	for _, k := range keys {
		if err := store.Delete(ctx, k); err != nil {
			return err
		}
	}
	return nil
}

func copyKey(k []byte) []byte {
	c := make([]byte, len(k))
	copy(c, k)
	return c
}
