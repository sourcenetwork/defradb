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

import "C"

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

var TxnStore sync.Map

func TransactionCreate(concurrent bool, readOnly bool) GoCResult {
	ctx := context.Background()
	var tx client.Txn
	var err error

	if concurrent {
		tx, err = globalNode.DB.NewConcurrentTxn(ctx, readOnly)
	} else {
		tx, err = globalNode.DB.NewTxn(ctx, readOnly)
	}
	if err != nil {
		return returnGoC(1, fmt.Sprintf(errCreatingTxn, err), "")
	}
	TxnStore.Store(tx.ID(), tx)
	IDstring := strconv.FormatUint(tx.ID(), 10)
	retVal := fmt.Sprintf(`{"id": %s}`, IDstring)
	return returnGoC(0, "", retVal)
}

func TransactionCommit(txnID uint64) GoCResult {
	ctx := context.Background()

	tx, ok := TxnStore.Load(txnID)
	if !ok {
		return returnGoC(1, fmt.Sprintf(errTxnDoesNotExist, txnID), "")
	}
	txn := tx.(datastore.Txn) //nolint:forcetypeassert

	err := txn.Commit(ctx)
	if err != nil {
		return returnGoC(1, fmt.Sprintf(errTxnDoesNotExist, txnID), "")
	}
	TxnStore.Delete(txnID)
	return returnGoC(0, "", "")
}

func TransactionDiscard(txnID uint64) GoCResult {
	ctx := context.Background()

	tx, ok := TxnStore.Load(txnID)
	if !ok {
		return returnGoC(1, fmt.Sprintf(errTxnDoesNotExist, txnID), "")
	}
	txn := tx.(datastore.Txn) //nolint:forcetypeassert

	txn.Discard(ctx)
	TxnStore.Delete(txnID)
	return returnGoC(0, "", "")
}
