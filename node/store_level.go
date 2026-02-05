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
)

// LevelStore specifies the defradb in leveldb datastore
const LevelStore = StoreType("level")

func init() {
	constructor := func(ctx context.Context, options *StoreOptions) (corekv.TxnStore, error) {
		return leveldb.NewDatastore(options.path, nil)
	}
	purge := func(ctx context.Context, options *StoreOptions) error {
		return nil
	}
	storeConstructors[LevelStore] = constructor
	storePurgeFuncs[LevelStore] = purge
}
