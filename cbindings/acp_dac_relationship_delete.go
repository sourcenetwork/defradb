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

//export ACPDeleteDACActorRelationship
func ACPDeleteDACActorRelationship(
	nodePtr C.uintptr_t,
	identityPtr C.uintptr_t,
	collection *C.char,
	docID *C.char,
	relation *C.char,
	actor *C.char,
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

	delOpt := options.WithIdentity(options.DeleteDACActorRelationship(), iIdentity.FromContext(ctx))
	result, err := store.DeleteDACActorRelationship(
		ctx,
		C.GoString(collection),
		C.GoString(docID),
		C.GoString(relation),
		C.GoString(actor),
		delOpt,
	)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(result))
}
