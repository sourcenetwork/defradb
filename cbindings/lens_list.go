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

	"github.com/sourcenetwork/defradb/client/options"
	acpIdentity "github.com/sourcenetwork/defradb/internal/identity"
)

//export ListLenses
func ListLenses(nodePtr C.uintptr_t, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	listOpt := options.WithIdentity(options.ListLenses(), acpIdentity.FromContext(ctx))
	lenses, err := store.ListLenses(ctx, listOpt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	lensesJSON, err := json.Marshal(lenses)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", string(lensesJSON)))
}
