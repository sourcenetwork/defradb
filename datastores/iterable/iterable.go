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

type IterableTxn interface {
	ds.Txn
	Iterable
}

// IterableTxnDatastore is an interface that should be implemented by datastores that
// support iterable transactions allowing proper use of native multi-prefix iteration.
type IterableTxnDatastore interface {
	ds.TxnDatastore

	NewIterableTransaction(ctx context.Context, readOnly bool) (IterableTxn, error)
}

type Iterable interface {
	// Returns an iterator allowing for multi-prefix iteration
	GetIterator(q dsq.Query) (Iterator, error)
}

type Iterator interface {
	// Iterates across the given prefix
	IteratePrefix(ctx context.Context, startPrefix ds.Key, endPrefix ds.Key) (dsq.Results, error)
	Close() error
}
