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
	"runtime/cgo"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/node"
)

//export ACPAddDACPolicy
func ACPAddDACPolicy(nodePtr C.uintptr_t, identity *C.char, policy *C.char) *C.Result {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, C.GoString(identity))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	h := cgo.Handle(nodePtr)
	node := h.Value().(*node.Node)
	policyResult, err := node.DB.AddDACPolicy(ctx, C.GoString(policy))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(policyResult))
}

//export ACPAddDACActorRelationship
func ACPAddDACActorRelationship(
	nodePtr C.uintptr_t,
	identity *C.char,
	collection *C.char,
	docID *C.char,
	relation *C.char,
	actor *C.char,
) *C.Result {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, C.GoString(identity))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	h := cgo.Handle(nodePtr)
	node := h.Value().(*node.Node)
	result, err := node.DB.AddDACActorRelationship(
		ctx,
		C.GoString(collection),
		C.GoString(docID),
		C.GoString(relation),
		C.GoString(actor),
	)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(result))
}

//export ACPDeleteDACActorRelationship
func ACPDeleteDACActorRelationship(
	nodePtr C.uintptr_t,
	identity *C.char,
	collection *C.char,
	docID *C.char,
	relation *C.char,
	actor *C.char,
) *C.Result {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, C.GoString(identity))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	h := cgo.Handle(nodePtr)
	node := h.Value().(*node.Node)
	result, err := node.DB.DeleteDACActorRelationship(
		ctx,
		C.GoString(collection),
		C.GoString(docID),
		C.GoString(relation),
		C.GoString(actor),
	)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(result))
}

//export ACPDisableNAC
func ACPDisableNAC(nodePtr C.uintptr_t, identity *C.char) *C.Result {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, C.GoString(identity))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	h := cgo.Handle(nodePtr)
	node := h.Value().(*node.Node)
	if err := node.DB.DisableNAC(ctx); err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(client.SuccessResponse{Success: true}))
}

//export ACPReEnableNAC
func ACPReEnableNAC(nodePtr C.uintptr_t, identity *C.char) *C.Result {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, C.GoString(identity))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	h := cgo.Handle(nodePtr)
	node := h.Value().(*node.Node)
	if err := node.DB.ReEnableNAC(ctx); err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(client.SuccessResponse{Success: true}))
}

//export ACPAddNACActorRelationship
func ACPAddNACActorRelationship(
	nodePtr C.uintptr_t,
	identity *C.char,
	relation *C.char,
	actor *C.char,
) *C.Result {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, C.GoString(identity))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	h := cgo.Handle(nodePtr)
	node := h.Value().(*node.Node)
	addNACActorRelationshipResult, err := node.DB.AddNACActorRelationship(
		ctx,
		C.GoString(relation),
		C.GoString(actor),
	)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(addNACActorRelationshipResult))
}

//export ACPDeleteNACActorRelationship
func ACPDeleteNACActorRelationship(
	nodePtr C.uintptr_t,
	identity *C.char,
	relation *C.char,
	actor *C.char,
) *C.Result {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, C.GoString(identity))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	h := cgo.Handle(nodePtr)
	node := h.Value().(*node.Node)
	deleteNACActorRelationshipResult, err := node.DB.DeleteNACActorRelationship(
		ctx,
		C.GoString(relation),
		C.GoString(actor),
	)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(deleteNACActorRelationshipResult))
}

//export ACPGetNACStatus
func ACPGetNACStatus(nodePtr C.uintptr_t, identity *C.char) *C.Result {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, C.GoString(identity))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	h := cgo.Handle(nodePtr)
	node := h.Value().(*node.Node)
	status, err := node.DB.GetNACStatus(ctx)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(status))
}
