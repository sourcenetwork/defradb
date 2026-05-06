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
	"encoding/json"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	acpIdentity "github.com/sourcenetwork/defradb/internal/identity"
)

//export DeleteDocument
func DeleteDocument(nodePtr C.uintptr_t,
	docIDStr *C.char,
	filterStr *C.char,
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

	docID := C.GoString(docIDStr)
	filter := C.GoString(filterStr)
	switch {
	case docID != "":
		ID, err := client.NewDocIDFromString(docID)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		_, err = col.DeleteDocument(ctx, ID, options.WithIdentity(options.DeleteDocument(), ident))
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		return returnC(returnGoC(0, "", ""))
	case filter != "":
		var filterValue any
		if err := json.Unmarshal([]byte(filter), &filterValue); err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		deleteOpt := options.WithIdentity(options.DeleteDocumentsWithFilter(), ident)
		res, err := col.DeleteDocumentsWithFilter(ctx, filterValue, deleteOpt)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		jsonBytes, err := json.Marshal(res)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		return returnC(returnGoC(0, "", string(jsonBytes)))
	default:
		return returnC(returnGoC(1, errNoDocIDOrFilter, ""))
	}
}
