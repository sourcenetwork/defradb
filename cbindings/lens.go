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

	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
)

//export LensSet
func LensSet(nodePtr C.uintptr_t, src *C.char, dst *C.char, cfg *C.char) C.Result {
	ctx := context.Background()

	decoder := json.NewDecoder(strings.NewReader(C.GoString(cfg)))
	decoder.DisallowUnknownFields()
	var lensCfg model.Lens
	if err := decoder.Decode(&lensCfg); err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	migrationCfg := client.LensConfig{
		SourceCollectionVersionID:      C.GoString(src),
		DestinationCollectionVersionID: C.GoString(dst),
		Lens:                           lensCfg,
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	lensID, err := store.SetMigration(ctx, migrationCfg)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", lensID))
}

//export LensAdd
func LensAdd(nodePtr C.uintptr_t, identityPtr C.uintptr_t, cfg *C.char) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	decoder := json.NewDecoder(strings.NewReader(C.GoString(cfg)))
	decoder.DisallowUnknownFields()
	var lensCfg model.Lens
	if err := decoder.Decode(&lensCfg); err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	lensID, err := store.AddLens(ctx, lensCfg)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", lensID))
}

//export LensList
func LensList(nodePtr C.uintptr_t, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	lenses, err := store.ListLenses(ctx)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	lensesJSON, err := json.Marshal(lenses)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", string(lensesJSON)))
}
