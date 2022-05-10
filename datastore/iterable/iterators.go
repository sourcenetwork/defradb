// Copyright 2022 Democratized Data Foundation
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
	badgerds "github.com/ipfs/go-ds-badger2"
)

// implement interface check
var _ Iterator = (*badgerds.BadgerIterator)(nil)

// Doesn't work.
// "github.com/ipfs/go-ds-badger2".Datastore does not
//     implement IterableTxnDatastore (wrong type for NewIterableTransaction method)
//        have NewIterableTransaction(context.Context, bool) (*"github.com/ipfs/go-ds-badger2".txn, error)
//        want NewIterableTransaction(context.Context, bool) (IterableTxn, error)
// var _ IterableTxnDatastore = (*badgerds.Datastore)(nil)
