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
	"strings"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/client/options"
	acpIdentity "github.com/sourcenetwork/defradb/internal/identity"
)

//export PatchCollection
func PatchCollection(nodePtr C.uintptr_t,
	patch *C.char, lensConfig *C.char,
	identityPtr C.uintptr_t,
) C.Result {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	var migration immutable.Option[model.Lens] = immutable.None[model.Lens]()
	lensString := C.GoString(lensConfig)
	if lensString != "" {
		var lensCfg model.Lens
		decoder := json.NewDecoder(strings.NewReader(lensString))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&lensCfg); err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}

		// Length being greater than 0 also means it is not nil, so no need to check
		if len(lensCfg.Lenses) > 0 {
			migration = immutable.Some(lensCfg)
		}
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	ctx = attachTxnFromPointer(nodePtr, ctx)

	err = store.PatchCollection(ctx, C.GoString(patch), migration,
		options.WithIdentity(options.PatchCollection(), acpIdentity.FromContext(ctx)))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}
