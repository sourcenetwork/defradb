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

	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/datastore/badger/v4"
	"github.com/sourcenetwork/defradb/datastore/memory"
)

// StoreOptions contains store configuration values.
type StoreOptions struct {
	path             string
	inMemory         bool
	defraStore       bool
	valueLogFileSize int64
	encryptionKey    []byte
}

// DefaultStoreOptions returns new options with default values.
func DefaultStoreOptions() *StoreOptions {
	return &StoreOptions{
		inMemory:         false,
		valueLogFileSize: 1 << 30,
	}
}

// StoreOpt is a function for setting configuration values.
type StoreOpt func(*StoreOptions)

// WithInMemory sets the in memory flag.
func WithInMemory(inMemory bool) StoreOpt {
	return func(o *StoreOptions) {
		o.inMemory = inMemory
	}
}

// WithDefraStore sets the defra store flag.
//
// Setting this to true will result in the defra node being created with
// the a custom defra implementation of the rootstore instead of badger.
func WithDefraStore(defraStore bool) StoreOpt {
	return func(o *StoreOptions) {
		o.defraStore = defraStore
	}
}

// WithPath sets the datastore path.
func WithPath(path string) StoreOpt {
	return func(o *StoreOptions) {
		o.path = path
	}
}

// WithValueLogFileSize sets the badger value log file size.
func WithValueLogFileSize(size int64) StoreOpt {
	return func(o *StoreOptions) {
		o.valueLogFileSize = size
	}
}

// WithEncryptionKey sets the badger encryption key.
func WithEncryptionKey(encryptionKey []byte) StoreOpt {
	return func(o *StoreOptions) {
		o.encryptionKey = encryptionKey
	}
}

// NewStore returns a new store with the given options.
func NewStore(ctx context.Context, opts ...StoreOpt) (datastore.Rootstore, error) {
	options := DefaultStoreOptions()
	for _, opt := range opts {
		opt(options)
	}

	if options.defraStore {
		return memory.NewDatastore(ctx), nil
	}

	badgerOpts := badger.DefaultOptions
	badgerOpts.InMemory = options.inMemory
	badgerOpts.ValueLogFileSize = options.valueLogFileSize
	badgerOpts.EncryptionKey = options.encryptionKey

	if len(options.encryptionKey) > 0 {
		// Having a cache improves the performance.
		// Otherwise, your reads would be very slow while encryption is enabled.
		// https://dgraph.io/docs/badger/get-started/#encryption-mode
		badgerOpts.IndexCacheSize = 100 << 20
	}

	return badger.NewDatastore(options.path, &badgerOpts)
}
