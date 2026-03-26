// Copyright 2025 Democratized Data Foundation
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
	"encoding/json"
	"strings"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	acpIdentity "github.com/sourcenetwork/defradb/internal/identity"
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

//export DescribeCollection
func DescribeCollection(nodePtr C.uintptr_t, opts C.CollectionOptions, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	colOptions := parseCollectionOptionsToGetCollectionsOptions(opts)
	ident := acpIdentity.FromContext(ctx)
	if ident.HasValue() {
		colOptions.SetIdentity(ident.Value())
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	ctx = attachTxnFromPointer(nodePtr, ctx)

	cols, err := store.GetCollections(ctx, colOptions)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	colDesc := make([]client.CollectionVersion, len(cols))
	for i, col := range cols {
		colDesc[i] = col.Version()
	}

	return returnC(marshalJSONToGoCResult(colDesc))
}

//export PatchCollection
func PatchCollection(nodePtr C.uintptr_t,
	patch *C.char, lensConfig *C.char,
	identityPtr C.uintptr_t,
) C.Result {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	var migration immutable.Option[model.Lens] = immutable.None[model.Lens]()
	lensString := C.GoString(lensConfig)
	if lensString != "" {
		var lensCfg model.Lens
		decoder := json.NewDecoder(strings.NewReader(lensString))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&lensCfg); err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}

		// Length being greater than 0 also means it is not nil, so no need to check
		if len(lensCfg.Lenses) > 0 {
			migration = immutable.Some(lensCfg)
		}
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	ctx = attachTxnFromPointer(nodePtr, ctx)

	err = store.PatchCollection(ctx, C.GoString(patch), migration,
		options.WithIdentity(options.PatchCollection(), acpIdentity.FromContext(ctx)))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export SetActiveCollection
func SetActiveCollection(nodePtr C.uintptr_t, opts C.CollectionOptions, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	versionID := C.GoString(opts.version)

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	ctx = attachTxnFromPointer(nodePtr, ctx)

	err = store.SetActiveCollectionVersion(ctx, versionID,
		options.WithIdentity(options.SetActiveCollectionVersion(), acpIdentity.FromContext(ctx)))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export TruncateCollection
func TruncateCollection(
	nodePtr C.uintptr_t,
	opts C.CollectionOptions,
	identityPtr C.uintptr_t,
) C.Result {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	colOptions := parseCollectionOptionsToGetCollectionsOptions(opts)
	ident := acpIdentity.FromContext(ctx)
	if ident.HasValue() {
		colOptions.SetIdentity(ident.Value())
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	ctx = attachTxnFromPointer(nodePtr, ctx)

	col, err := getCollection(store, ctx, colOptions)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	err = col.Truncate(ctx, options.WithIdentity(options.TruncateCollection(), ident))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(returnGoC(0, "", ""))
}
