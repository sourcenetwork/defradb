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
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
)

//export DeleteEncryptedIndex
func DeleteEncryptedIndex(
	nodePtr C.uintptr_t, collectionName *C.char, fieldName *C.char, identityPtr C.uintptr_t,
) C.Result {
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

	getColOpt := options.WithIdentity(options.GetCollectionByName(), iIdentity.FromContext(ctx))
	col, err := store.GetCollectionByName(ctx, C.GoString(collectionName), getColOpt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	deleteOpt := options.WithIdentity(options.DeleteEncryptedIndex(), iIdentity.FromContext(ctx))
	err = col.DeleteEncryptedIndex(ctx, C.GoString(fieldName), deleteOpt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}
