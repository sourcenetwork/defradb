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
)

//export ViewAdd
func ViewAdd(nodePtr C.uintptr_t, query *C.char, sdl *C.char, transformStr *C.char) *C.Result {
	ctx := context.Background()

	var transform immutable.Option[model.Lens]
	lensCfgJson := C.GoString(transformStr)
	if lensCfgJson != "" {
		decoder := json.NewDecoder(strings.NewReader(lensCfgJson))
		decoder.DisallowUnknownFields()
		var lensCfg model.Lens
		if err := decoder.Decode(&lensCfg); err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		transform = immutable.Some(lensCfg)
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	defs, err := store.AddView(ctx, C.GoString(query), C.GoString(sdl), transform)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(defs))
}

//export ViewRefresh
func ViewRefresh(
	nodePtr C.uintptr_t,
	viewNameStr *C.char,
	collectionIDStr *C.char,
	versionIDStr *C.char,
	getInactive C.int,
) *C.Result {
	ctx := context.Background()

	viewName := C.GoString(viewNameStr)
	collectionID := C.GoString(collectionIDStr)
	versionID := C.GoString(versionIDStr)
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
	if getInactive != 0 {
		options.IncludeInactive = immutable.Some(getInactive != 0)
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
