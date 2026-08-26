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

	"github.com/sourcenetwork/defradb/client/options"
	acpIdentity "github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/internal/utils"
)

//export TruncateCollection
func TruncateCollection(
	nodePtr C.uintptr_t,
	opts C.CollectionOptions,
	identityPtr C.uintptr_t,
) C.Result {
	return truncateCollection(nodePtr, opts, identityPtr, nil, false)
}

// TruncateCollectionWithFilter preserves TruncateCollection's v1 C ABI.
//
// Deprecated: This compatibility function will be removed in v2, when TruncateCollection
// accepts filtered-truncate options.
//
//export TruncateCollectionWithFilter
func TruncateCollectionWithFilter(
	nodePtr C.uintptr_t,
	opts C.CollectionOptions,
	identityPtr C.uintptr_t,
	filterJSON *C.char,
	pruneHistory C.int,
) C.Result {
	if filterJSON == nil {
		return returnC(returnGoC(1, "filter is required", ""))
	}
	filter, err := utils.DecodeJSONFilter([]byte(C.GoString(filterJSON)))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	// JSON null must not fall through to an unfiltered collection truncate.
	if filter == nil {
		return returnC(returnGoC(1, "filter cannot be null", ""))
	}
	return truncateCollection(nodePtr, opts, identityPtr, filter, pruneHistory != 0)
}

func truncateCollection(
	nodePtr C.uintptr_t,
	opts C.CollectionOptions,
	identityPtr C.uintptr_t,
	filter any,
	pruneHistory bool,
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

	ctx = attachTxnFromPointer(nodePtr, ctx)

	col, err := getCollection(store, ctx, colOptions)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	truncateOpts := options.WithIdentity(options.TruncateCollection(), ident)
	if filter != nil {
		truncateOpts.SetFilter(filter).SetPruneHistory(pruneHistory)
	}
	err = col.Truncate(ctx, truncateOpts)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(returnGoC(0, "", ""))
}
