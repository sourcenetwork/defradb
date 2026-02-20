// Copyright 2024 Democratized Data Foundation
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

	badgerds "github.com/dgraph-io/badger/v4"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/badger"

	"github.com/sourcenetwork/defradb/client/options"
)

func init() {
	constructor := func(ctx context.Context, opts *options.NodeStoreOptions) (corekv.TxnStore, error) {
		var path string
		if !opts.BadgerInMemory {
			// Badger will error if we give it a path and set `InMemory` to true
			path = opts.Path
		}

		badgerOpts := badgerds.DefaultOptions(path)
		badgerOpts.InMemory = opts.BadgerInMemory
		badgerOpts.ValueLogFileSize = opts.BadgerFileSize
		badgerOpts.EncryptionKey = opts.BadgerEncryptionKey

		if len(opts.BadgerEncryptionKey) > 0 {
			// Having a cache improves the performance.
			// Otherwise, your reads would be very slow while encryption is enabled.
			// https://dgraph.io/docs/badger/get-started/#encryption-mode
			badgerOpts.IndexCacheSize = 100 << 20
		}

		return badger.NewDatastore(path, badgerOpts)
	}
	purge := func(ctx context.Context, opts *options.NodeStoreOptions) error {
		store, err := constructor(ctx, opts)
		if err != nil {
			return err
		}
		err = store.(corekv.Dropable).DropAll()
		if err != nil {
			return err
		}
		return store.Close()
	}

	storeConstructors[options.NodeBadgerStore] = constructor
	storePurgeFuncs[options.NodeBadgerStore] = purge

	storeConstructors[options.NodeDefaultStore] = constructor
	storePurgeFuncs[options.NodeDefaultStore] = purge
}
