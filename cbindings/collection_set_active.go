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
)

//export SetActiveCollection
func SetActiveCollection(nodePtr C.uintptr_t, opts C.CollectionOptions, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	versionID := C.GoString(opts.version)

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	ctx = attachTxnFromPointer(nodePtr, ctx)

	err = store.SetActiveCollectionVersion(ctx, versionID,
		options.WithIdentity(options.SetActiveCollectionVersion(), acpIdentity.FromContext(ctx)))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}
