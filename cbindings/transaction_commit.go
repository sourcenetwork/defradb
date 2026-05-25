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
	"runtime/cgo"

	"github.com/sourcenetwork/defradb/client"
)

//export CommitTransaction
func CommitTransaction(txnPtr C.uintptr_t) C.Result {
	h := cgo.Handle(txnPtr)
	defer h.Delete()
	txn := h.Value().(client.Txn) //nolint:forcetypeassert

	err := txn.Commit()
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(returnGoC(0, "", ""))
}
