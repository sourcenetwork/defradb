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
)

// BadgerStore specifies the badger datastore
const BadgerStore = StoreType("badger")

func init() {
	constructor := func(ctx context.Context, options *StoreOptions) (corekv.TxnStore, error) {
		var path string
		if !options.badgerInMemory {
			// Badger will error if we give it a path and set `InMemory` to true
			path = options.path
		}

		badgerOpts := badgerds.DefaultOptions(path)
		badgerOpts.InMemory = options.badgerInMemory
		badgerOpts.ValueLogFileSize = options.badgerFileSize
		badgerOpts.EncryptionKey = options.badgerEncryptionKey

		badgerOpts.NumCompactors = 4
		badgerOpts.NumLevelZeroTables = 30
		badgerOpts.NumLevelZeroTablesStall = 60
		badgerOpts.MemTableSize = 512 << 20
		badgerOpts.BlockCacheSize = 4 << 30
		badgerOpts.ValueThreshold = 1 << 10
		badgerOpts.IndexCacheSize = 512 << 20
		badgerOpts.Compression = badgeropts.None

		return badger.NewDatastore(path, badgerOpts)
	}
	purge := func(ctx context.Context, options *StoreOptions) error {
		store, err := constructor(ctx, options)
		if err != nil {
			return err
		}
		err = store.(corekv.Dropable).DropAll()
		if err != nil {
			return err
		}
		return store.Close()
	}

	storeConstructors[BadgerStore] = constructor
	storePurgeFuncs[BadgerStore] = purge

	storeConstructors[DefaultStore] = constructor
	storePurgeFuncs[DefaultStore] = purge
}

// WithBadgerInMemory sets the badger in memory option.
func WithBadgerInMemory(enable bool) StoreOpt {
	return func(o *StoreOptions) {
		o.badgerInMemory = enable
	}
}

// WithBadgerFileSize sets the badger value log file size.
func WithBadgerFileSize(size int64) StoreOpt {
	return func(o *StoreOptions) {
		o.badgerFileSize = size
	}
}

// WithBadgerEncryptionKey sets the badger encryption key.
func WithBadgerEncryptionKey(encryptionKey []byte) StoreOpt {
	return func(o *StoreOptions) {
		o.badgerEncryptionKey = encryptionKey
	}
}
