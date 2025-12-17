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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
)

//export ViewAdd
func ViewAdd(nodePtr C.uintptr_t,
	query *C.char,
	sdl *C.char,
	transformCIDStr *C.char,
	identityPtr C.uintptr_t,
) C.Result {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	var transformCID immutable.Option[string]
	transformCIDValue := C.GoString(transformCIDStr)
	if transformCIDValue != "" {
		transformCID = immutable.Some(transformCIDValue)
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	defs, err := store.AddView(ctx, C.GoString(query), C.GoString(sdl), transformCID)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(defs))
}

//export ViewRefresh
func ViewRefresh(nodePtr C.uintptr_t,
	cOptions C.CollectionOptions,
	identityPtr C.uintptr_t,
) C.Result {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	viewName := C.GoString(cOptions.name)
	collectionID := C.GoString(cOptions.collectionID)
	versionID := C.GoString(cOptions.version)

	options := client.CollectionFetchOptions{}
	if versionID != "" {
		options.VersionID = immutable.Some(versionID)
	}
	if collectionID != "" {
		options.CollectionID = immutable.Some(collectionID)
	}
	if viewName != "" {
		options.Name = immutable.Some(viewName)
	}
	if cOptions.getInactive != 0 {
		options.IncludeInactive = immutable.Some(true)
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	err = store.RefreshViews(ctx, options)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}
