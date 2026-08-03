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
	"time"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/namespace"
)

// ReclaimOrphanBlocks deletes blocks that were fetched during P2P sync but never
// merged into a document, identified by a to-merge marker older than cutoff. The
// marker is removed atomically when a block's merge commits, so a marker that
// outlives the sync window belongs to no committed document and dropping its block
// is safe; a later reference re-fetches it. History prune never reaches these blocks
// because it walks out from documents, so without this sweep they accumulate for the
// life of the store.
//
// It resumes from startKey (nil begins at the start of the marker index) and scans at
// most scanLimit markers per call, returning the key to resume from next (nil once the
// index has been swept end to end), the number of blocks reclaimed, and the number of
// markers scanned. Markers are collected during the scan and deleted after it, so the
// iterator is never mutated while open.
func ReclaimOrphanBlocks(
	ctx context.Context,
	rootstore corekv.ReaderWriter,
	cutoff time.Time,
	startKey []byte,
	scanLimit int,
) (nextKey []byte, reclaimed int, scanned int, err error) {
	blockNS := namespace.Wrap(rootstore, []byte{blockStoreKey})

	expired, nextKey, scanned, err := collectExpiredMarkers(ctx, blockNS, cutoff, startKey, scanLimit)
	if err != nil {
		return nil, 0, scanned, err
	}

	for _, marker := range expired {
		// A marker key is the to-merge prefix followed by the block's CID.
		cid := marker[1:]
		// Delete the block before the marker: if this is interrupted, the surviving
		// marker keeps the block reclaimable on a later sweep.
		if err := blockNS.Delete(ctx, cid); err != nil {
			return nil, reclaimed, scanned, err
		}
		if err := blockNS.Delete(ctx, marker); err != nil {
			return nil, reclaimed, scanned, err
		}
		reclaimed++
	}
	return nextKey, reclaimed, scanned, nil
}

// collectExpiredMarkers scans the to-merge index from startKey and returns the keys
// of markers older than cutoff (including older single-byte markers, which have no
// timestamp), along with the key to resume from next. Kept keys are copied because
// the iterator reuses its buffers across Next.
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
		if scanned == scanLimit {
			// The current marker is unexamined; resume from it next call.
			return expired, copyKey(iter.Key()), scanned, nil
		}
		var value []byte
		if value, err = iter.Value(); err != nil {
			break
		}
		if t, decoded := toMergeTime(value); !decoded || t.Before(cutoff) {
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

func copyKey(k []byte) []byte {
	c := make([]byte, len(k))
	copy(c, k)
	return c
}
