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

//export PurgeDocuments
func PurgeDocuments(
	nodePtr C.uintptr_t,
	opts C.CollectionOptions,
	docIDsJSON *C.char,
	pruneHistory C.int,
	identityPtr C.uintptr_t,
) C.Result {
	ctx, err := contextWithIdentity(context.Background(), identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	var rawIDs []string
	if err := json.Unmarshal([]byte(C.GoString(docIDsJSON)), &rawIDs); err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	docIDs := make([]client.DocID, len(rawIDs))
	for i, rawID := range rawIDs {
		docIDs[i], err = client.NewDocIDFromString(rawID)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
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

	ctx = attachTxnFromPointer(nodePtr, ctx)
	col, err := getCollection(store, ctx, colOptions)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	err = col.PurgeByDocIDs(
		ctx,
		docIDs,
		pruneHistory != 0,
		options.WithIdentity(options.TruncateCollection(), ident),
	)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(returnGoC(0, "", ""))
}
