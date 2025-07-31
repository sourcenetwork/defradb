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

var (
	txnStoreMapMu sync.RWMutex
	TxnStoreMap   = make(map[int]*sync.Map) // nodeID -> (txnID -> Txn)
)

func TransactionCreate(n int, concurrent bool, readOnly bool) GoCResult {
	ctx := context.Background()
	var tx client.Txn
	var err error

	if concurrent {
		tx, err = GlobalNodes[n].DB.NewConcurrentTxn(ctx, readOnly)
	} else {
		tx, err = GlobalNodes[n].DB.NewTxn(ctx, readOnly)
	}
	if err != nil {
		return returnGoC(1, fmt.Sprintf(errCreatingTxn, err), "")
	}

	txnStoreMapMu.Lock()
	if TxnStoreMap[n] == nil {
		TxnStoreMap[n] = &sync.Map{}
	}
	txnStore := TxnStoreMap[n]
	txnStoreMapMu.Unlock()

	txnStore.Store(tx.ID(), tx)

	IDstring := strconv.FormatUint(tx.ID(), 10)
	retVal := fmt.Sprintf(`{"id": %s}`, IDstring)
	return returnGoC(0, "", retVal)
}

func TransactionCommit(n int, txnID uint64) GoCResult {
	ctx := context.Background()

	txnStoreMapMu.RLock()
	txnStore, ok := TxnStoreMap[n]
	txnStoreMapMu.RUnlock()
	if !ok {
		return returnGoC(1, fmt.Sprintf(errTxnDoesNotExist, txnID), "")
	}

	tx, ok := txnStore.Load(txnID)
	if !ok {
		return returnGoC(1, fmt.Sprintf(errTxnDoesNotExist, txnID), "")
	}

	txn := tx.(datastore.Txn) //nolint:forcetypeassert
	err := txn.Commit(ctx)
	if err != nil {
		return returnGoC(1, fmt.Sprintf(errTxnDoesNotExist, txnID), "")
	}

	txnStore.Delete(txnID)
	return returnGoC(0, "", "")
}

func TransactionDiscard(n int, txnID uint64) GoCResult {
	ctx := context.Background()

	txnStoreMapMu.RLock()
	txnStore, ok := TxnStoreMap[n]
	txnStoreMapMu.RUnlock()
	if !ok {
		return returnGoC(1, fmt.Sprintf(errTxnDoesNotExist, txnID), "")
	}

	tx, ok := txnStore.Load(txnID)
	if !ok {
		return returnGoC(1, fmt.Sprintf(errTxnDoesNotExist, txnID), "")
	}

	txn := tx.(datastore.Txn) //nolint:forcetypeassert
	txn.Discard(ctx)

	txnStore.Delete(txnID)
	return returnGoC(0, "", "")
}
