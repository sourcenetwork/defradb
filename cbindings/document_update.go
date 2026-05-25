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

//export UpdateDocument
func UpdateDocument(
	nodePtr C.uintptr_t,
	docIDStr *C.char,
	filterStr *C.char,
	updaterStr *C.char,
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
	updater := C.GoString(updaterStr)
	switch {
	// Update by filter
	case filter != "":
		var filterValue any
		if err := json.Unmarshal([]byte(filter), &filterValue); err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		res, err := col.UpdateDocumentsWithFilter(ctx, filterValue, updater,
			options.WithIdentity(options.UpdateDocumentsWithFilter(), ident))
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		jsonBytes, err := json.Marshal(res)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		return returnC(returnGoC(0, "", string(jsonBytes)))

	// Update by docID
	case docID != "":
		newDocID, err := client.NewDocIDFromString(docID)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		doc, err := col.GetDocument(ctx, newDocID,
			options.WithIdentity(options.GetDocument().SetShowDeleted(true), ident))
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		if err := doc.SetWithJSON(ctx, []byte(updater)); err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		err = col.UpdateDocument(ctx, doc, options.WithIdentity(options.UpdateDocument(), ident))
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		return returnC(returnGoC(0, "", ""))
	default:
		return returnC(returnGoC(1, errNoDocIDOrFilter, ""))
	}
}
