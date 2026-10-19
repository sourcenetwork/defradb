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

	defraOpts "github.com/sourcenetwork/defradb/client/options"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
)

//export ListIndexes
func ListIndexes(nodePtr C.uintptr_t, options C.CollectionOptions, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	ctx = attachTxnFromPointer(nodePtr, ctx)

	ident := iIdentity.FromContext(ctx)
	collectionName := C.GoString(options.name)
	switch {
	// Get the indices associated with a given collection
	case collectionName != "":
		col, err := store.GetCollectionByName(ctx, collectionName,
			defraOpts.WithIdentity(defraOpts.GetCollectionByName(), ident))
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		indices, err := col.ListIndexes(ctx,
			defraOpts.WithIdentity(defraOpts.ListCollectionIndexes(), ident))
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		return returnC(marshalJSONToGoCResult(indices))
	// Get all of the indices, because no collection was specified
	default:
		indices, err := store.ListIndexes(ctx,
			defraOpts.WithIdentity(defraOpts.ListIndexes(), ident))
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		return returnC(marshalJSONToGoCResult(indices))
	}
}
