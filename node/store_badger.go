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
	badgeropts "github.com/dgraph-io/badger/v4/options"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/badger"

	"github.com/sourcenetwork/defradb/client/options"
)

// rawBadgerDB holds a reference to the underlying Badger DB instance
var rawBadgerDB *badgerds.DB

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
		badgerOpts.IndexCacheSize = 100 << 20
		badgerOpts.ValueThreshold = 1 << 10
		badgerOpts.Compression = badgeropts.ZSTD
		badgerOpts.ZSTDCompressionLevel = 1

		if opts.BadgerMemTableSize > 0 {
			badgerOpts.MemTableSize = opts.BadgerMemTableSize
		}
		if opts.BadgerBlockCacheSize > 0 {
			badgerOpts.BlockCacheSize = opts.BadgerBlockCacheSize
		}
		if opts.BadgerIndexCacheSize > 0 {
			badgerOpts.IndexCacheSize = opts.BadgerIndexCacheSize
		}
		if opts.BadgerNumCompactors > 0 {
			badgerOpts.NumCompactors = opts.BadgerNumCompactors
		}
		if opts.BadgerNumLevelZeroTables > 0 {
			badgerOpts.NumLevelZeroTables = opts.BadgerNumLevelZeroTables
		}
		if opts.BadgerNumLevelZeroTablesStall > 0 {
			badgerOpts.NumLevelZeroTablesStall = opts.BadgerNumLevelZeroTablesStall
		}

		badgerOpts.Logger = nil
		db, err := badgerds.Open(badgerOpts)
		if err != nil {
			return nil, err
		}
		rawBadgerDB = db
		return badger.NewDatastoreFrom(db), nil
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
