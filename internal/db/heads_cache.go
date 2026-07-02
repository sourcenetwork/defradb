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

	"github.com/ipfs/go-cid"
	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/internal/keys"
)

// headsCache maps docShortID → composite head CIDs for that document.
// It is stored in a context and used to avoid redundant headstore scans
// when multiple documents are merged in a single batch (Tier 3B).
type headsCache map[uint64][]cid.Cid

type headsCacheContextKey struct{}

// InitHeadsCacheContext returns a child context that carries a fresh heads cache.
// Call this before prefetching heads for a batch merge.
func InitHeadsCacheContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, headsCacheContextKey{}, make(headsCache))
}

// PrefetchDocHeads scans the headstore for docShortID's composite heads and stores
// them in the context cache.  No-ops if the context has no cache or the doc's heads
// are already cached.  Errors are best-effort: a scan failure leaves the cache empty
// for this docShortID (callers will fall back to a live scan).
func PrefetchDocHeads(ctx context.Context, headstore corekv.ReaderWriter, docShortID uint64) error {
	cache, ok := ctx.Value(headsCacheContextKey{}).(headsCache)
	if !ok {
		return nil // no cache in context, nothing to do
	}
	if _, already := cache[docShortID]; already {
		return nil
	}

	prefix := keys.HeadstoreDocKey{DocShortID: docShortID}.Bytes()
	iter, err := headstore.Iterator(ctx, corekv.IterOptions{Prefix: prefix})
	if err != nil {
		return err
	}

	var cids []cid.Cid
	for {
		hasNext, err := iter.Next()
		if err != nil {
			_ = iter.Close()
			return err
		}
		if !hasNext {
			break
		}
		hk, err := keys.NewHeadstoreDocKey(string(iter.Key()))
		if err != nil {
			continue
		}
		if hk.Cid.Defined() {
			cids = append(cids, hk.Cid)
		}
	}
	if err := iter.Close(); err != nil {
		return err
	}

	cache[docShortID] = cids
	return nil
}

// GetCachedHeads returns the composite head CIDs for docShortID if they were
// previously fetched by PrefetchDocHeads.  Returns false if the cache is absent
// or this docShortID has not been prefetched.
func GetCachedHeads(ctx context.Context, docShortID uint64) ([]cid.Cid, bool) {
	cache, ok := ctx.Value(headsCacheContextKey{}).(headsCache)
	if !ok {
		return nil, false
	}
	cids, ok := cache[docShortID]
	return cids, ok
}
