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

	"github.com/ipfs/go-cid"
	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/internal/db/txnctx"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// HeadFetcher is a utility to incrementally fetch all the MerkleCRDT heads of a given doc/field.
type HeadFetcher struct {
	kvIter corekv.Iterator
}

// Start starts/initializes the fetcher, performing all the work it can do outside
// of the main iteration loop/funcs.
//
// prefix - Optional. The headstore prefix to scan across.  If None, the entire
// headstore will be scanned - for example, in order to fetch document and collection
// heads.
func (hf *HeadFetcher) Start(
	ctx context.Context,
	prefix immutable.Option[keys.HeadstoreKey],
) error {
	txn := txnctx.MustGet(ctx)

	var prefixBytes []byte
	if prefix.HasValue() {
		prefixBytes = prefix.Value().Bytes()
	}

	if hf.kvIter != nil {
		if err := hf.kvIter.Close(); err != nil {
			return err
		}
	}

	iter, err := txn.Headstore().Iterator(ctx, corekv.IterOptions{
		Prefix: prefixBytes,
	})
	if err != nil {
		return err
	}

	hf.kvIter = iter
	return nil
}

func (hf *HeadFetcher) FetchNext() (*cid.Cid, error) {
	hasValue, err := hf.kvIter.Next()
	if err != nil || !hasValue {
		return nil, err
	}

	headStoreKey, err := keys.NewHeadstoreKey(string(hf.kvIter.Key()))
	if err != nil {
		return nil, err
	}

	cid := headStoreKey.GetCid()
	return &cid, nil
}

func (hf *HeadFetcher) Close() error {
	if hf.kvIter == nil {
		return nil
	}

	return hf.kvIter.Close()
}
