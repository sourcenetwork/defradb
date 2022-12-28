// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package datastore

import (
	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/datastore/iterable"
)

// AsDSReaderWriter returns a DSReaderWriter from the given datastore.
func AsDSReaderWriter(store ds.Datastore) DSReaderWriter {
	switch typedStore := store.(type) {
	case iterable.IterableDatastore:
		return typedStore
	default:
		return shim{
			store,
			iterable.NewIterable(store),
		}
	}
}

type shim struct {
	ds.Datastore
	iterable.Iterable
}

var _ DSReaderWriter = (*shim)(nil)
