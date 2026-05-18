// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package fetcher

import (
	"context"
	"errors"

	"github.com/ipfs/go-cid"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// HeadFetcher is a utility to incrementally fetch all the MerkleCRDT heads of a given doc/field.
type HeadFetcher struct {
	kvIters   []corekv.Iterator
	iterIndex int
}

// Start starts/initializes the fetcher, performing all the work it can do outside
// of the main iteration loop/funcs.
//
// prefix - Optional. The headstore prefix to scan across.  If None, only collection
// and document commit heads will be scanned.
func (hf *HeadFetcher) Start(
	ctx context.Context,
	prefix immutable.Option[keys.HeadstoreKey],
) error {
	txn := datastore.CtxMustGetTxn(ctx)

	if len(hf.kvIters) > 0 {
		var firstErr error
		for _, iter := range hf.kvIters {
			if err := iter.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		if firstErr != nil {
			return firstErr
		}
	}

	hf.kvIters = nil
	hf.iterIndex = 0

	if prefix.HasValue() {
		iter, err := txn.Headstore().Iterator(ctx, corekv.IterOptions{
			Prefix: prefix.Value().Bytes(),
		})
		if err != nil {
			return NewErrCreateHeadIterator(err)
		}
		hf.kvIters = []corekv.Iterator{iter}
		return nil
	}

	// When no specific prefix is provided, explicitly scan only the
	// collection commits ("/c") and document commits ("/d") prefixes,
	// excluding schema-definition keys ("/f", "/g", "/s").
	// This avoids relying on lexicographic ordering of key prefixes.
	colIter, err := txn.Headstore().Iterator(ctx, corekv.IterOptions{
		Prefix: []byte(keys.HEADSTORE_COL),
	})
	if err != nil {
		return NewErrCreateHeadIterator(err)
	}

	docIter, err := txn.Headstore().Iterator(ctx, corekv.IterOptions{
		Prefix: []byte(keys.HEADSTORE_DOC),
	})
	if err != nil {
		return errors.Join(NewErrCreateHeadIterator(err), colIter.Close())
	}

	hf.kvIters = []corekv.Iterator{colIter, docIter}
	return nil
}

func (hf *HeadFetcher) FetchNext() (*cid.Cid, error) {
	for hf.iterIndex < len(hf.kvIters) {
		hasValue, err := hf.kvIters[hf.iterIndex].Next()
		if err != nil {
			return nil, NewErrIterateHeads(err)
		}
		if !hasValue {
			hf.iterIndex++
			continue
		}

		headStoreKey, err := keys.NewHeadstoreKey(string(hf.kvIters[hf.iterIndex].Key()))
		if err != nil {
			return nil, NewErrParseHeadKey(err)
		}

		cid := headStoreKey.GetCid()
		return &cid, nil
	}

	return nil, nil
}

func (hf *HeadFetcher) Close() error {
	var firstErr error
	for _, iter := range hf.kvIters {
		if err := iter.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}
