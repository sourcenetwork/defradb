// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package iterable

import (
	"context"

	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
)

// implement interface check
var _ IterableTxn = (*iterableTransactionShim)(nil)
var _ Iterator = (*iteratorShim)(nil)

type iterableTransactionShim struct {
	ds.Txn
}

type iteratorShim struct {
	txn     iterableTransactionShim
	q       dsq.Query
	results dsq.Results
}

func NewIterableTransaction(txn ds.Txn) IterableTxn {
	return iterableTransactionShim{
		txn,
	}
}

func (shim iterableTransactionShim) GetIterator(q dsq.Query) (Iterator, error) {
	return &iteratorShim{
		txn: shim,
		q:   q,
	}, nil
}

func (shim *iteratorShim) IteratePrefix(ctx context.Context, prefix ds.Key) (dsq.Results, error) {
	if shim.results != nil {
		err := shim.results.Close()
		if err != nil {
			return nil, err
		}
	}

	query := shim.q
	query.Prefix = prefix.String()
	results, err := shim.txn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	shim.results = results
	return results, nil
}

func (shim *iteratorShim) Close() error {
	if shim.results == nil {
		return nil
	}
	return shim.results.Close()
}
