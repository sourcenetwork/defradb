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
	"context"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
)

//export ACPAddDACPolicy
func ACPAddDACPolicy(nodePtr C.uintptr_t, identityPtr C.uintptr_t, policy *C.char) C.Result {
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

	opt := options.WithIdentity(options.AddDACPolicy(), iIdentity.FromContext(ctx))
	policyResult, err := store.AddDACPolicy(ctx, C.GoString(policy), opt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(policyResult))
}

//export ACPAddDACActorRelationship
func ACPAddDACActorRelationship(
	nodePtr C.uintptr_t,
	identityPtr C.uintptr_t,
	collection *C.char,
	docID *C.char,
	relation *C.char,
	actor *C.char,
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

	addOpt := options.WithIdentity(options.AddDACActorRelationship(), iIdentity.FromContext(ctx))
	result, err := store.AddDACActorRelationship(
		ctx,
		C.GoString(collection),
		C.GoString(docID),
		C.GoString(relation),
		C.GoString(actor),
		addOpt,
	)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(result))
}

//export ACPDeleteDACActorRelationship
func ACPDeleteDACActorRelationship(
	nodePtr C.uintptr_t,
	identityPtr C.uintptr_t,
	collection *C.char,
	docID *C.char,
	relation *C.char,
	actor *C.char,
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

	delOpt := options.WithIdentity(options.DeleteDACActorRelationship(), iIdentity.FromContext(ctx))
	result, err := store.DeleteDACActorRelationship(
		ctx,
		C.GoString(collection),
		C.GoString(docID),
		C.GoString(relation),
		C.GoString(actor),
		delOpt,
	)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(result))
}

//export ACPDisableNAC
func ACPDisableNAC(nodePtr C.uintptr_t, identityPtr C.uintptr_t) C.Result {
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

	disableOpt := options.WithIdentity(options.DisableNAC(), iIdentity.FromContext(ctx))
	if err := store.DisableNAC(ctx, disableOpt); err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(client.SuccessResponse{Success: true}))
}

//export ACPReEnableNAC
func ACPReEnableNAC(nodePtr C.uintptr_t, identityPtr C.uintptr_t) C.Result {
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

	reEnableOpt := options.WithIdentity(options.ReEnableNAC(), iIdentity.FromContext(ctx))
	if err := store.ReEnableNAC(ctx, reEnableOpt); err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(client.SuccessResponse{Success: true}))
}

//export ACPAddNACActorRelationship
func ACPAddNACActorRelationship(
	nodePtr C.uintptr_t,
	identityPtr C.uintptr_t,
	relation *C.char,
	actor *C.char,
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

	addNACOpt := options.WithIdentity(options.AddNACActorRelationship(), iIdentity.FromContext(ctx))
	addNACActorRelationshipResult, err := store.AddNACActorRelationship(
		ctx,
		C.GoString(relation),
		C.GoString(actor),
		addNACOpt,
	)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(addNACActorRelationshipResult))
}

//export ACPDeleteNACActorRelationship
func ACPDeleteNACActorRelationship(
	nodePtr C.uintptr_t,
	identityPtr C.uintptr_t,
	relation *C.char,
	actor *C.char,
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

	delNACOpt := options.WithIdentity(options.DeleteNACActorRelationship(), iIdentity.FromContext(ctx))
	deleteNACActorRelationshipResult, err := store.DeleteNACActorRelationship(
		ctx,
		C.GoString(relation),
		C.GoString(actor),
		delNACOpt,
	)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(deleteNACActorRelationshipResult))
}

//export ACPGetNACStatus
func ACPGetNACStatus(nodePtr C.uintptr_t, identityPtr C.uintptr_t) C.Result {
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

	getNACOpt := options.WithIdentity(options.GetNACStatus(), iIdentity.FromContext(ctx))
	status, err := store.GetNACStatus(ctx, getNACOpt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(status))
}
