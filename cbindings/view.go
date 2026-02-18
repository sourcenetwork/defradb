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

	"github.com/sourcenetwork/defradb/client/options"
	acpIdentity "github.com/sourcenetwork/defradb/internal/identity"
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

	opts := options.WithIdentity(options.AddView(), acpIdentity.FromContext(ctx))
	transformCIDValue := C.GoString(transformCIDStr)
	if transformCIDValue != "" {
		opts.SetTransformCID(transformCIDValue)
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	defs, err := store.AddView(ctx, C.GoString(query), C.GoString(sdl), opts)
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

	opt := options.WithIdentity(options.RefreshViews(), acpIdentity.FromContext(ctx))
	if versionID != "" {
		opt.SetVersionID(versionID)
	}
	if collectionID != "" {
		opt.SetCollectionID(collectionID)
	}
	if viewName != "" {
		opt.SetCollectionName(viewName)
	}
	if cOptions.getInactive != 0 {
		opt.SetGetInactive(true)
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	err = store.RefreshViews(ctx, opt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}
