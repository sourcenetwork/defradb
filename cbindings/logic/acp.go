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

func ACPAddPolicy(n int, identityStr string, policy string, TxnID uint64) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, identityStr)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = contextWithTransaction(n, ctx, TxnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	policyResult, err := GlobalNodes[n].DB.AddDACPolicy(ctx, policy)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	return marshalJSONToGoCResult(policyResult)
}

func ACPAddRelationship(n int, identityStr string,
	collectionArg string,
	docIDArg string,
	relationArg string,
	targetActorArg string,
	TxnID uint64,
) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, identityStr)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = contextWithTransaction(n, ctx, TxnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	result, err := GlobalNodes[n].DB.AddDACActorRelationship(ctx, collectionArg, docIDArg, relationArg, targetActorArg)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	return marshalJSONToGoCResult(result)
}

func ACPDeleteRelationship(
	n int,
	identityStr string,
	collectionArg string,
	docIDArg string,
	relationArg string,
	targetActorArg string,
	TxnID uint64,
) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithIdentity(ctx, identityStr)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = contextWithTransaction(n, ctx, TxnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	result, err := GlobalNodes[n].DB.DeleteDACActorRelationship(ctx, collectionArg, docIDArg, relationArg, targetActorArg)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	return marshalJSONToGoCResult(result)
}
