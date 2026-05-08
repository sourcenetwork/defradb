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
	"strings"

	"github.com/sourcenetwork/defradb/client/options"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
)

//export DeleteCollection
func DeleteCollection(
	nodePtr C.uintptr_t,
	names *C.char,
	activeOnly C.int,
	identityPtr C.uintptr_t,
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

	opt := options.WithIdentity(options.DeleteCollection(), iIdentity.FromContext(ctx))
	opt.SetActiveOnly(activeOnly != 0)

	var nameList []string
	if joined := C.GoString(names); joined != "" {
		nameList = strings.Split(joined, ",")
	}

	err = store.DeleteCollection(ctx, nameList, opt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}
