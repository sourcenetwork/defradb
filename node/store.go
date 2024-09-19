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

// storePurgeFuncs is a map of [StoreType]s to store purge functions.
//
// Is is populated by the `init` functions in the runtime-specific files - this
// allows it's population to be managed by build flags.
var storePurgeFuncs = map[StoreType]func(ctx context.Context, options *StoreOptions) error{}

// StoreOptions contains store configuration values.
type StoreOptions struct {
	store               StoreType
	path                string
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

// WithStorePath sets the store path.
func WithStorePath(path string) StoreOpt {
	return func(o *StoreOptions) {
		o.path = path
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

func purgeStore(ctx context.Context, opts ...StoreOpt) error {
	options := DefaultStoreOptions()
	for _, opt := range opts {
		opt(options)
	}
	purgeFunc, ok := storePurgeFuncs[options.store]
	if ok {
		return purgeFunc(ctx, options)
	}
	return NewErrStoreTypeNotSupported(options.store)
}
