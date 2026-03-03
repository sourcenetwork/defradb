// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package node

import (
	"context"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/leveldb"

	"github.com/sourcenetwork/defradb/client/options"
)

func init() {
	constructor := func(ctx context.Context, opts *options.NodeStoreOptions) (corekv.TxnStore, error) {
		return leveldb.NewDatastore(opts.Path, nil)
	}
	purge := func(ctx context.Context, opts *options.NodeStoreOptions) error {
		return nil
	}
	storeConstructors[options.NodeStoreType("level")] = constructor
	storePurgeFuncs[options.NodeStoreType("level")] = purge
}
