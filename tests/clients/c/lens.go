// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cwrap

/*
#include "defra_structs.h"
*/
import "C"

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable/enumerable"

	"github.com/sourcenetwork/defradb/client"
)

func LensSet(cSrc *C.char, cDst *C.char, cCfg *C.char, cTxnID C.ulonglong) *C.Result {
	ctx := context.Background()
	srcSchemaVersionID := C.GoString(cSrc)
	dstSchemaVersionID := C.GoString(cDst)
	lensCfgJson := C.GoString(cCfg)

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Parse the lens config string into a client.LensConfig
	decoder := json.NewDecoder(strings.NewReader(lensCfgJson))
	decoder.DisallowUnknownFields()
	var lensCfg model.Lens
	if err := decoder.Decode(&lensCfg); err != nil {
		return returnC(1, fmt.Sprintf(cerrInvalidLensConfig, err), "")
	}
	migrationCfg := client.LensConfig{
		SourceSchemaVersionID:      srcSchemaVersionID,
		DestinationSchemaVersionID: dstSchemaVersionID,
		Lens:                       lensCfg,
	}
	err = globalNode.DB.SetMigration(ctx, migrationCfg)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return returnC(0, "", "")
}

func LensDown(cCollectionID *C.char, cDocuments *C.char, cTxnID C.ulonglong) *C.Result {
	ctx := context.Background()
	collectionID := C.GoString(cCollectionID)
	srcData := []byte(C.GoString(cDocuments))

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Decode the input documents
	var src []map[string]any
	if err := json.Unmarshal(srcData, &src); err != nil {
		return returnC(1, err.Error(), "")
	}

	// Call the lens down migration
	out, err := globalNode.DB.LensRegistry().MigrateDown(ctx, enumerable.New(src), collectionID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	// Each reversed document will be appended to a value array
	var value []map[string]any
	err = enumerable.ForEach(out, func(item map[string]any) {
		value = append(value, item)
	})
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	return marshalJSONToCResult(value)
}

func LensUp(cCollectionID *C.char, cDocuments *C.char, cTxnID C.ulonglong) *C.Result {
	ctx := context.Background()
	collectionID := C.GoString(cCollectionID)
	srcData := []byte(C.GoString(cDocuments))

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Decode the input documents
	var src []map[string]any
	if err := json.Unmarshal(srcData, &src); err != nil {
		return returnC(1, err.Error(), "")
	}

	// Call the lens down migration
	out, err := globalNode.DB.LensRegistry().MigrateUp(ctx, enumerable.New(src), collectionID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	// Each reversed document will be appended to a value array
	var value []map[string]any
	err = enumerable.ForEach(out, func(item map[string]any) {
		value = append(value, item)
	})
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	return marshalJSONToCResult(value)
}

func LensReload(cTxnID C.ulonglong) *C.Result {
	ctx := context.Background()

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Reload the lenses and return
	err = globalNode.DB.LensRegistry().ReloadLenses(ctx)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return returnC(0, "", "")
}

func LensSetRegistry(cCollectionID *C.char, cLensCfg *C.char, cTxnID C.ulonglong) *C.Result {
	ctx := context.Background()
	collectionID := C.GoString(cCollectionID)
	lensCfgJSON := C.GoString(cLensCfg)

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Create a model.Lens from the lens configuration
	decoder := json.NewDecoder(strings.NewReader(lensCfgJSON))
	decoder.DisallowUnknownFields()
	var lensCfg model.Lens
	if err := decoder.Decode(&lensCfg); err != nil {
		return returnC(1, fmt.Sprintf(cerrInvalidLensConfig, err), "")
	}

	// Set migration and return
	err = globalNode.DB.LensRegistry().SetMigration(ctx, collectionID, lensCfg)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return returnC(0, "", "")
}
