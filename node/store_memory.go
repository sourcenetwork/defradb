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
	"github.com/sourcenetwork/defradb/datastore/memory"
)

// MemoryStore specifies the defradb in memory datastore
const MemoryStore = StoreType("memory")

func init() {
	constructor := func(ctx context.Context, options *StoreOptions) (datastore.Rootstore, error) {
		return memory.NewDatastore(ctx), nil
	}
	purge := func(ctx context.Context, options *StoreOptions) error {
		return nil
	}
	// don't override the default constructor if previously set
	if _, ok := storeConstructors[DefaultStore]; !ok {
		storeConstructors[DefaultStore] = constructor
		storePurgeFuncs[DefaultStore] = purge
	}
	storeConstructors[MemoryStore] = constructor
	storePurgeFuncs[MemoryStore] = purge
}
