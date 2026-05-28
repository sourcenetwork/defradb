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
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
)

//export ExecuteQuery
func ExecuteQuery(
	nodePtr C.uintptr_t,
	query *C.char,
	identityPtr C.uintptr_t,
	operationName *C.char,
	variables *C.char,
) C.Result {
	ctx := context.Background()

	opt := options.ExecRequest()
	opName := C.GoString(operationName)
	if opName != "" {
		opt.SetOperationName(opName)
	}
	varsStr := C.GoString(variables)
	if varsStr != "" {
		var vars map[string]any
		if err := json.Unmarshal([]byte(varsStr), &vars); err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		opt.SetVariables(vars)
	}

	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	ident := iIdentity.FromContext(ctx)
	if ident.HasValue() {
		opt.SetIdentity(ident.Value())
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	ctx = attachTxnFromPointer(nodePtr, ctx)

	ctx, cancelFunc := context.WithCancel(ctx)

	res := store.ExecRequest(ctx, C.GoString(query), opt)
	sub := &Subscription{
		ctxCancel:  cancelFunc,
		resultChan: res.Subscription,
	}
	// The return is either a subscription ID, or a GQL result. The status indicates
	// which: 0 for GQL, 2 for subscription. 1 is not used because this cannot error; the
	// error is part of the GQL result, to be GQL-compliant.
	if res.Subscription != nil {
		id := storeSubscription(sub)
		return returnC(returnGoC(2, "", id))
	}
	return returnC(marshalJSONToGoCResult(res.GQL))
}
