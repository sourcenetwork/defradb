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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/immutable"
)

func AddSchema(cSchema *C.char, cTxnID C.ulonglong) *C.Result {
	newSchema := C.GoString(cSchema)
	ctx := context.Background()

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	collectionVersions, err := globalNode.DB.AddSchema(ctx, newSchema)
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrAddingSchema, err), "")
	}
	return marshalJSONToCResult(collectionVersions)
}

func DescribeSchema(cName *C.char, cRoot *C.char, cVersion *C.char, cTxnID C.ulonglong) *C.Result {
	ctx := context.Background()
	options := client.SchemaFetchOptions{}

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Set the configuration options from the passed in parameters
	if cVersion != nil && C.GoString(cVersion) != "" {
		options.ID = immutable.Some(C.GoString(cVersion))
	}
	if cRoot != nil && C.GoString(cRoot) != "" {
		options.Root = immutable.Some(C.GoString(cRoot))
	}
	if cName != nil && C.GoString(cName) != "" {
		options.Name = immutable.Some(C.GoString(cName))
	}

	// Get the schema, and try to convert it to JSON for return
	schemas, err := globalNode.DB.GetSchemas(ctx, options)
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrGettingSchema, err), "")
	}
	return marshalJSONToCResult(schemas)
}

func PatchSchema(cPatch *C.char, cLensConfig *C.char, cSetActive C.int, cTxnID C.ulonglong) *C.Result {
	ctx := context.Background()
	setActive := cSetActive != 0
	patchString := C.GoString(cPatch)
	lensString := C.GoString(cLensConfig)

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	if cPatch == nil || patchString == "" {
		return returnC(1, cerrEmptyPatch, "")
	}

	var migration immutable.Option[model.Lens] = immutable.None[model.Lens]()
	if lensString != "" {
		var lensCfg model.Lens
		decoder := json.NewDecoder(strings.NewReader(lensString))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&lensCfg); err != nil {
			return returnC(1, fmt.Sprintf(cerrInvalidLensConfig, err), "")
		}

		// Only assign Some if Lense length is greater than 0 (and therefore it is not nil)
		if len(lensCfg.Lenses) > 0 {
			migration = immutable.Some(lensCfg)
		}
	}

	err = globalNode.DB.PatchSchema(ctx, patchString, migration, setActive)
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrPatchingSchema, err), "")
	}
	return returnC(0, "", "")
}

func SetActiveSchema(cVersion *C.char, cTxnID C.ulonglong) *C.Result {
	ctx := context.Background()

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	err = globalNode.DB.SetActiveSchemaVersion(ctx, C.GoString(cVersion))
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrSetActiveSchema, err), "")
	}
	return returnC(0, "", "")
}
