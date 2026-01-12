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
	"os"
	"path/filepath"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/client/options"
)

// TODO: remove all types for backward compatibility
// StoreType is an alias for options.NodeStoreType for backward compatibility.
type StoreType = options.NodeStoreType

const (
	// DefaultStore is the default store type.
	DefaultStore = options.NodeDefaultStore
)

// storeConstructors is a map of [StoreType]s to store constructors.
//
// Is is populated by the `init` functions in the runtime-specific files - this
// allows it's population to be managed by build flags.
var storeConstructors = map[StoreType]func(
	ctx context.Context,
	opts *options.NodeStoreOptions,
) (corekv.TxnStore, error){}

// storePurgeFuncs is a map of [StoreType]s to store purge functions.
//
// Is is populated by the `init` functions in the runtime-specific files - this
// allows it's population to be managed by build flags.
var storePurgeFuncs = map[StoreType]func(ctx context.Context, opts *options.NodeStoreOptions) error{}

// StoreOptions is an alias for options.NodeStoreOptions for backward compatibility.
type StoreOptions = options.NodeStoreOptions

// GetDefaultStorePath is a helper function that returns '$HOME/.defradb', but which
// relies on Go to handle the platform-specific path resolution.
func GetDefaultStorePath() string {
	home, err := os.UserHomeDir()
	// This should never error on any major platform. But if it does, as a fallback,
	// we will leave the root directory path blank.
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".defradb")
}

// DefaultStoreOptions returns new options with default values.
func DefaultStoreOptions() *options.NodeStoreOptions {
	return options.NodeStore().SetPath(GetDefaultStorePath())
}

// NewStore returns a new store with the given options.
func NewStore(ctx context.Context, opts ...*options.NodeStoreOptions) (corekv.TxnStore, bool, error) {
	var opt *options.NodeStoreOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	if opt == nil {
		opt = DefaultStoreOptions()
	}

	var isValueSizeLimited bool
	if opt.BadgerInMemory {
		isValueSizeLimited = true
	}

	storeConstructor, ok := storeConstructors[opt.Store]
	if ok {
		store, err := storeConstructor(ctx, opt)
		return store, isValueSizeLimited, err
	}

	return nil, false, NewErrStoreTypeNotSupported(opt.Store)
}

func purgeStore(ctx context.Context, opts *options.NodeStoreOptions) error {
	if opts == nil {
		opts = DefaultStoreOptions()
	}
	purgeFunc, ok := storePurgeFuncs[opts.Store]
	if ok {
		return purgeFunc(ctx, opts)
	}
	return NewErrStoreTypeNotSupported(opts.Store)
}
