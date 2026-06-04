// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cbindings

/*
#include <stdlib.h>
#include "defra_structs.h"
*/
import "C"

import (
	"context"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
)

// parseCollectionOptionsToGetCollectionsOptions is a helper function that converts
// a C.CollectionOptions struct into a GetCollectionsOptions
func parseCollectionOptionsToGetCollectionsOptions(
	opts C.CollectionOptions,
) *options.GetCollectionsOptionsBuilder {
	versionID := C.GoString(opts.version)
	collectionID := C.GoString(opts.collectionID)
	name := C.GoString(opts.name)
	getInactive := opts.getInactive != 0
	opt := options.GetCollections()
	if versionID != "" {
		opt.SetVersionID(versionID)
	}
	if collectionID != "" {
		opt.SetCollectionID(collectionID)
	}
	if name != "" {
		opt.SetCollectionName(name)
	}
	if getInactive {
		opt.SetGetInactive(getInactive)
	}
	return opt
}

// getCollection is a helper function wrapping DB.GetCollections, and ensuring
// that only one collection matches the criteria
func getCollection(
	store client.Store,
	ctx context.Context,
	builder options.Enumerable[options.GetCollectionsOptions],
) (client.Collection, error) {
	cols, err := store.GetCollections(ctx, builder)
	if err != nil {
		return nil, err
	}

	// Only one collection should match the criteria
	if len(cols) == 0 {
		return nil, client.ErrCollectionNotFound
	}
	if len(cols) > 1 {
		return nil, NewErrAmbiguousCollection()
	}
	return cols[0], nil
}
