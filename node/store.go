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
	"github.com/sourcenetwork/defradb/internal/utils"
)

// storeConstructors is a map of store types to store constructors.
//
// Is is populated by the `init` functions in the runtime-specific files - this
// allows it's population to be managed by build flags.
var storeConstructors = map[options.NodeStoreType]func(
	ctx context.Context,
	opts *options.NodeStoreOptions,
) (corekv.TxnStore, error){}

// storePurgeFuncs is a map of store types to store purge functions.
//
// Is is populated by the `init` functions in the runtime-specific files - this
// allows it's population to be managed by build flags.
var storePurgeFuncs = map[options.NodeStoreType]func(
	ctx context.Context,
	opts *options.NodeStoreOptions,
) error{}

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

// NewStore returns a new store with the given options.
func NewStore(
	ctx context.Context, opts ...options.Enumerable[options.NodeStoreOptions],
) (corekv.TxnStore, bool, error) {
	o := utils.NewOptions(opts...)
	var isValueSizeLimited bool
	if o.BadgerInMemory {
		isValueSizeLimited = true
	}

	storeConstructor, ok := storeConstructors[o.Store]
	if ok {
		store, err := storeConstructor(ctx, o)
		return store, isValueSizeLimited, err
	}

	return nil, false, NewErrStoreTypeNotSupported(o.Store)
}

func purgeStore(ctx context.Context, opts *options.NodeStoreOptions) error {
	purgeFunc, ok := storePurgeFuncs[opts.Store]
	if ok {
		return purgeFunc(ctx, opts)
	}
	return NewErrStoreTypeNotSupported(opts.Store)
}
