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

	"github.com/sourcenetwork/immutable/enumerable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client"
)

//export LensSet
func LensSet(nodePtr C.uintptr_t, src *C.char, dst *C.char, cfg *C.char) *C.Result {
	ctx := context.Background()

	decoder := json.NewDecoder(strings.NewReader(C.GoString(cfg)))
	decoder.DisallowUnknownFields()
	var lensCfg model.Lens
	if err := decoder.Decode(&lensCfg); err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	migrationCfg := client.LensConfig{
		SourceSchemaVersionID:      C.GoString(src),
		DestinationSchemaVersionID: C.GoString(dst),
		Lens:                       lensCfg,
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	err = store.SetMigration(ctx, migrationCfg)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export LensDown
func LensDown(nodePtr C.uintptr_t, collectionID *C.char, documents *C.char) *C.Result {
	ctx := context.Background()
	srcData := []byte(C.GoString(documents))

	var src []map[string]any
	if err := json.Unmarshal(srcData, &src); err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	out, err := store.LensRegistry().MigrateDown(ctx, enumerable.New(src), C.GoString(collectionID))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	var value []map[string]any
	err = enumerable.ForEach(out, func(item map[string]any) {
		value = append(value, item)
	})
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(value))
}

//export LensUp
func LensUp(nodePtr C.uintptr_t, collectionID *C.char, documents *C.char) *C.Result {
	ctx := context.Background()
	srcData := []byte(C.GoString(documents))

	var src []map[string]any
	if err := json.Unmarshal(srcData, &src); err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	out, err := store.LensRegistry().MigrateUp(ctx, enumerable.New(src), C.GoString(collectionID))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	var value []map[string]any
	err = enumerable.ForEach(out, func(item map[string]any) {
		value = append(value, item)
	})
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(value))
}

//export LensReload
func LensReload(nodePtr C.uintptr_t) *C.Result {
	ctx := context.Background()

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	err = store.LensRegistry().ReloadLenses(ctx)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export LensSetRegistry
func LensSetRegistry(nodePtr C.uintptr_t, collectionID *C.char, cfg *C.char) *C.Result {
	ctx := context.Background()

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

	err = store.LensRegistry().SetMigration(ctx, C.GoString(collectionID), lensCfg)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}
