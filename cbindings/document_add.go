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
	"strings"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/encryption"
	acpIdentity "github.com/sourcenetwork/defradb/internal/identity"
)

//export AddDocument
func AddDocument(
	nodePtr C.uintptr_t,
	json *C.char,
	isEncrypted C.int,
	encryptedFields *C.char,
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

	col, err := getCollection(store, ctx, colOptions)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	var encryptFields []string
	encryptFieldsStr := C.GoString(encryptedFields)
	if encryptFieldsStr != "" {
		for _, f := range strings.Split(encryptFieldsStr, ",") {
			if trimmed := strings.TrimSpace(f); trimmed != "" {
				encryptFields = append(encryptFields, trimmed)
			}
		}
	}
	ctx = encryption.SetContextConfigFromParams(ctx, isEncrypted != 0, encryptFields)

	addOpt := options.WithIdentity(options.AddDocument(), acpIdentity.FromContext(ctx))

	// Determine if JSON is array or object by looking for the first character being [
	jsonString := strings.TrimSpace(C.GoString(json))
	if strings.HasPrefix(jsonString, "[") {
		// Multiple documents
		docs, err := client.NewDocsFromJSON(ctx, []byte(jsonString), col.Version())
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		err = col.AddManyDocuments(ctx, docs, addOpt)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
	} else {
		// Single document
		doc, err := client.NewDocFromJSON(ctx, []byte(jsonString), col.Version())
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		err = col.AddDocument(ctx, doc, addOpt)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
	}
	return returnC(returnGoC(0, "", ""))
}
