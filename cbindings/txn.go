// Copyright 2025 Democratized Data Foundation
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
	"github.com/sourcenetwork/defradb/node"
)

//export TransactionCreate
func TransactionCreate(nodePtr C.uintptr_t, isConcurrent C.int, isReadOnly C.int) C.NewTxnResult {
	h := cgo.Handle(nodePtr)
	n := h.Value().(*node.Node) //nolint:forcetypeassert

	var tx client.Txn
	var err error
	if isConcurrent != 0 {
		tx, err = n.DB.NewConcurrentTxn(isReadOnly != 0)
	} else {
		tx, err = n.DB.NewTxn(isReadOnly != 0)
	}
	if err != nil {
		return returnNewTxnResultC(1, err.Error(), nil)
	}

	return returnNewTxnResultC(0, "", tx)
}

//export TransactionCommit
func TransactionCommit(txnPtr C.uintptr_t) C.Result {
	h := cgo.Handle(txnPtr)
	defer h.Delete()
	txn := h.Value().(client.Txn) //nolint:forcetypeassert

	err := txn.Commit()
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(returnGoC(0, "", ""))
}

//export TransactionDiscard
func TransactionDiscard(txnPtr C.uintptr_t) {
	// Avoid panic in the case of a double discard
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()

	h := cgo.Handle(txnPtr)
	txn := h.Value().(client.Txn) //nolint:forcetypeassert
	txn.Discard()
	h.Delete()
}
