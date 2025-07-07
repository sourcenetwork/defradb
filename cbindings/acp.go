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
)

//export acpAddPolicy
func acpAddPolicy(cIdentity *C.char, cPolicy *C.char, cTxnID C.ulonglong) *C.Result {
	ctx := context.Background()
	policy := C.GoString(cPolicy)
	identityStr := C.GoString(cIdentity)

	// Attach the identity to the context
	newctx, err := contextWithIdentity(ctx, identityStr)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Set the transaction
	newctx, err = contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Try to add the policy
	policyResult, err := globalNode.DB.AddDACPolicy(ctx, policy)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	return marshalJSONToCResult(policyResult)
}

//export acpAddRelationship
func acpAddRelationship(cIdentity *C.char,
	cCollection *C.char,
	cDocID *C.char,
	cRelation *C.char,
	cActor *C.char,
	cTxnID C.ulonglong,
) *C.Result {
	ctx := context.Background()
	collectionArg := C.GoString(cCollection)
	docIDArg := C.GoString(cDocID)
	relationArg := C.GoString(cRelation)
	targetActorArg := C.GoString(cActor)
	identityStr := C.GoString(cIdentity)

	// Attach the identity to the context
	newctx, err := contextWithIdentity(ctx, identityStr)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Set the transaction
	newctx, err = contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Add the relationship
	result, err := globalNode.DB.AddDACActorRelationship(ctx, collectionArg, docIDArg, relationArg, targetActorArg)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	return marshalJSONToCResult(result)
}

//export acpDeleteRelationship
func acpDeleteRelationship(
	cIdentity *C.char,
	cCollection *C.char,
	cDocID *C.char,
	cRelation *C.char,
	cActor *C.char,
	cTxnID C.ulonglong,
) *C.Result {
	ctx := context.Background()
	collectionArg := C.GoString(cCollection)
	docIDArg := C.GoString(cDocID)
	relationArg := C.GoString(cRelation)
	targetActorArg := C.GoString(cActor)
	identityStr := C.GoString(cIdentity)

	// Attach the identity to the context
	newctx, err := contextWithIdentity(ctx, identityStr)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Set the transaction
	newctx, err = contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Delete the relationship
	result, err := globalNode.DB.DeleteDACActorRelationship(ctx, collectionArg, docIDArg, relationArg, targetActorArg)
	if err != nil {
		return returnC(1, err.Error(), "")
	}

	return marshalJSONToCResult(result)
}
