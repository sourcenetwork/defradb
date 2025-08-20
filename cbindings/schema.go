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
	"fmt"
)

//export AddSchema
func AddSchema(nodePtr C.uintptr_t, schema *C.char, identityPtr C.uintptr_t) *C.Result {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	store := getStoreFromPointer(nodePtr)
	collectionVersions, err := store.AddSchema(ctx, C.GoString(schema))
	if err != nil {
		return returnC(returnGoC(1, fmt.Sprintf(errAddingSchema, err), ""))
	}
	return returnC(marshalJSONToGoCResult(collectionVersions))
}
