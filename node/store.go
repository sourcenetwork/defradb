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
)

type StoreType string

const (
	// The Go-enum default StoreType.
	//
	// The actual store type that this resolves to depends on the build target.
	DefaultStore StoreType = ""
)

// storeConstructors is a map of [StoreType]s to store constructors.
//
// Is is populated by the `init` functions in the runtime-specific files - this
// allows it's population to be managed by build flags.
var storeConstructors = map[StoreType]func(ctx context.Context, options *StoreOptions) (datastore.Rootstore, error){}

// StoreOptions contains store configuration values.
type StoreOptions struct {
	store               StoreType
	badgerPath          string
	badgerFileSize      int64
	badgerEncryptionKey []byte
	badgerInMemory      bool
}

// DefaultStoreOptions returns new options with default values.
func DefaultStoreOptions() *StoreOptions {
	return &StoreOptions{
		badgerInMemory: false,
		badgerFileSize: 1 << 30,
	}
}

// StoreOpt is a function for setting configuration values.
type StoreOpt func(*StoreOptions)

// WithStoreType sets the store type to use.
func WithStoreType(store StoreType) StoreOpt {
	return func(o *StoreOptions) {
		o.store = store
	}
}

// WithBadgerInMemory sets the badger in memory option.
func WithBadgerInMemory(enable bool) StoreOpt {
	return func(o *StoreOptions) {
		o.badgerInMemory = enable
	}
}

// WithBadgerPath sets the badger datastore path.
func WithBadgerPath(path string) StoreOpt {
	return func(o *StoreOptions) {
		o.badgerPath = path
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

// NewStore returns a new store with the given options.
func NewStore(ctx context.Context, opts ...StoreOpt) (datastore.Rootstore, error) {
	options := DefaultStoreOptions()
	for _, opt := range opts {
		opt(options)
	}
	storeConstructor, ok := storeConstructors[options.store]
	if ok {
		return storeConstructor(ctx, options)
	}
	return nil, NewErrStoreTypeNotSupported(options.store)
}
