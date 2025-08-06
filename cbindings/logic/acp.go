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

import (
	"context"

	"github.com/sourcenetwork/defradb/client"
)

func ACPAddDACPolicy(n int, identityPrivateKey string, policy string, TxnID uint64) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, identityPrivateKey)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = contextWithTransaction(n, ctx, TxnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	policyResult, err := GetNode(n).DB.AddDACPolicy(ctx, policy)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	return marshalJSONToGoCResult(policyResult)
}

func ACPAddDACActorRelationship(n int, identityPrivateKey string,
	collectionArg string,
	docIDArg string,
	relationArg string,
	targetActorArg string,
	TxnID uint64,
) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, identityPrivateKey)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = contextWithTransaction(n, ctx, TxnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	result, err := GetNode(n).DB.AddDACActorRelationship(ctx, collectionArg, docIDArg, relationArg, targetActorArg)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	return marshalJSONToGoCResult(result)
}

func ACPDeleteDACActorRelationship(
	n int,
	identityPrivateKey string,
	collectionArg string,
	docIDArg string,
	relationArg string,
	targetActorArg string,
	TxnID uint64,
) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, identityPrivateKey)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = contextWithTransaction(n, ctx, TxnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	result, err := GetNode(n).DB.DeleteDACActorRelationship(ctx, collectionArg, docIDArg, relationArg, targetActorArg)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	return marshalJSONToGoCResult(result)
}

func ACPNodeDisable(n int, identityPrivateKey string, TxnID uint64) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, identityPrivateKey)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = contextWithTransaction(n, ctx, TxnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	if err := GetNode(n).DB.DisableNAC(ctx); err != nil {
		return returnGoC(1, err.Error(), "")
	}

	return marshalJSONToGoCResult(client.SuccessResponse{Success: true})
}

func ACPNodeReEnable(n int, identityPrivateKey string, TxnID uint64) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, identityPrivateKey)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = contextWithTransaction(n, ctx, TxnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	if err := GetNode(n).DB.ReEnableNAC(ctx); err != nil {
		return returnGoC(1, err.Error(), "")
	}

	return marshalJSONToGoCResult(client.SuccessResponse{Success: true})
}

func ACPNodeRelationshipAdd(
	n int,
	identityPrivateKey string,
	relationArg string,
	targetActorArg string,
	TxnID uint64,
) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, identityPrivateKey)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = contextWithTransaction(n, ctx, TxnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	addNACActorRelationshipResult, err := GetNode(n).DB.AddNACActorRelationship(ctx, relationArg, targetActorArg)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	return marshalJSONToGoCResult(addNACActorRelationshipResult)
}

func ACPNodeRelationshipDelete(
	n int,
	identityPrivateKey string,
	relationArg string,
	targetActorArg string,
	TxnID uint64,
) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, identityPrivateKey)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = contextWithTransaction(n, ctx, TxnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	deleteNACActorRelationshipResult, err := GetNode(n).DB.DeleteNACActorRelationship(ctx, relationArg, targetActorArg)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	return marshalJSONToGoCResult(deleteNACActorRelationshipResult)
}

func ACPNodeStatus(n int, identityPrivateKey string, TxnID uint64) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, identityPrivateKey)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = contextWithTransaction(n, ctx, TxnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	status, err := GetNode(n).DB.GetNACStatus(ctx)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	return marshalJSONToGoCResult(status)
}
