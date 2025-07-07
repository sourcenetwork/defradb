// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build cgo
// +build cgo

package main

/*
#include "defra_structs.h"
*/
import "C"

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

type transactionContextKey struct{}

var TxnStore sync.Map

//export transactionCreate
func transactionCreate(cIsConcurrent C.int, cIsReadOnly C.int) *C.Result {
	concurrent := cIsConcurrent != 0
	readOnly := cIsReadOnly != 0
	ctx := context.Background()
	var tx client.Txn
	var err error

	// Create a Txn object based on parameters passed in
	if concurrent {
		tx, err = globalNode.DB.NewConcurrentTxn(ctx, readOnly)
	} else {
		tx, err = globalNode.DB.NewTxn(ctx, readOnly)
	}
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrCreatingTxn, err), "")
	}
	// Store the Txn in the store, and return the ID to the user
	TxnStore.Store(tx.ID(), tx)
	IDstring := strconv.FormatUint(tx.ID(), 10)
	retVal := fmt.Sprintf(`{"id": %s}`, IDstring)
	return returnC(0, "", retVal)
}

//export transactionCommit
func transactionCommit(cTxnID C.ulonglong) *C.Result {
	TxnIDu64 := uint64(cTxnID)
	ctx := context.Background()

	// Get the transaction with the associated ID from the store
	tx, ok := TxnStore.Load(TxnIDu64)
	if !ok {
		return returnC(1, fmt.Sprintf(cerrTxnDoesNotExist, cTxnID), "")
	}
	txn := tx.(datastore.Txn) //nolint:forcetypeassert

	// Commit the transaction, and if that doesn't error, remove it from the store
	err := txn.Commit(ctx)
	if err != nil {
		return returnC(1, fmt.Sprintf(cerrTxnDoesNotExist, cTxnID), "")
	}
	TxnStore.Delete(TxnIDu64)
	return returnC(0, "", "")
}

//export transactionDiscard
func transactionDiscard(cTxnID C.ulonglong) *C.Result {
	TxnIDu64 := uint64(cTxnID)
	ctx := context.Background()

	// Get the transaction with the associated ID from the store
	tx, ok := TxnStore.Load(TxnIDu64)
	if !ok {
		return returnC(1, fmt.Sprintf(cerrTxnDoesNotExist, cTxnID), "")
	}
	txn := tx.(datastore.Txn) //nolint:forcetypeassert

	// Discard it, which currently cannot error, and then delete it from the store
	txn.Discard(ctx)
	TxnStore.Delete(TxnIDu64)
	return returnC(0, "", "")
}

// The following function exists to accomodate test functionality with
// the C client wrapper. But it is not exported.

func getTxnHandle(cTxnID C.ulonglong) any {
	TxnIDu64 := uint64(cTxnID)
	val, ok := TxnStore.Load(TxnIDu64)
	if !ok {
		return 0
	}
	return val
}
