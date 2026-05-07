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

	"github.com/sourcenetwork/defradb/node"
)

//export CreateTransaction
func CreateTransaction(nodePtr C.uintptr_t, isReadOnly C.int) C.NewTxnResult {
	h := cgo.Handle(nodePtr)
	n := h.Value().(*node.Node) //nolint:forcetypeassert

	tx, err := n.DB.NewTxn(isReadOnly != 0)
	if err != nil {
		return returnNewTxnResultC(1, err.Error(), nil)
	}

	return returnNewTxnResultC(0, "", tx)
}
