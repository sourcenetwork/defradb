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
)

//export GetNodeIdentity
func GetNodeIdentity(nodePtr C.uintptr_t) C.Result {
	ctx := context.Background()
	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	identity, err := store.GetNodeIdentity(ctx)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	if identity.HasValue() {
		return returnC(marshalJSONToGoCResult(identity.Value()))
	}
	return returnC(returnGoC(0, "", "Node has no identity assigned to it."))
}
