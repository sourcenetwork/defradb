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
)

func ACPAddPolicy(identityStr string, policy string, TxnID uint64) GoCResult {
	ctx := context.Background()

	// Attach the identity to the context
	newctx, err := contextWithIdentity(ctx, identityStr)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	ctx = newctx

	// Set the transaction
	newctx, err = contextWithTransaction(ctx, TxnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	ctx = newctx

	// Try to add the policy
	policyResult, err := globalNode.DB.AddDACPolicy(ctx, policy)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	return marshalJSONToGoCResult(policyResult)
}

func ACPAddRelationship(identityStr string,
	collectionArg string,
	docIDArg string,
	relationArg string,
	targetActorArg string,
	TxnID uint64,
) GoCResult {
	ctx := context.Background()

	// Attach the identity to the context
	newctx, err := contextWithIdentity(ctx, identityStr)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	ctx = newctx

	// Set the transaction
	newctx, err = contextWithTransaction(ctx, TxnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	ctx = newctx

	// Add the relationship
	result, err := globalNode.DB.AddDACActorRelationship(ctx, collectionArg, docIDArg, relationArg, targetActorArg)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	return marshalJSONToGoCResult(result)
}

func ACPDeleteRelationship(
	identityStr string,
	collectionArg string,
	docIDArg string,
	relationArg string,
	targetActorArg string,
	TxnID uint64,
) GoCResult {
	ctx := context.Background()

	// Attach the identity to the context
	newctx, err := contextWithIdentity(ctx, identityStr)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	ctx = newctx

	// Set the transaction
	newctx, err = contextWithTransaction(ctx, TxnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	ctx = newctx

	// Delete the relationship
	result, err := globalNode.DB.DeleteDACActorRelationship(ctx, collectionArg, docIDArg, relationArg, targetActorArg)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	return marshalJSONToGoCResult(result)
}
