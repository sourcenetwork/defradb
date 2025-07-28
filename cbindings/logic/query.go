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
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/sourcenetwork/defradb/client"
)

// We cannot return a channel to/from C, so instead we have a map of subscription IDs to
// RequestResults. Three functions, storeSubscription, getSubscription, and removeSubscription are
// helpers which manage this store behind the scenes, while PollSubscription and CloseSubscription
// are made available to the user for interacting with the subscriptions.

var subscriptionStore sync.Map // map[string]*client.RequestResult

func storeSubscription(res *client.RequestResult) string {
	id := uuid.NewString()
	subscriptionStore.Store(id, res)
	return id
}

func getSubscription(id string) (*client.RequestResult, bool) {
	val, ok := subscriptionStore.Load(id)
	if !ok {
		return nil, false
	}
	//nolint:forcetypeassert
	return val.(*client.RequestResult), true
}

func removeSubscription(id string) {
	subscriptionStore.Delete(id)
}

func PollSubscription(id string) GoCResult {
	res, ok := getSubscription(id)
	if !ok {
		return returnGoC(1, errInvalidSubscriptionID, "")
	}

	select {
	case msg, ok := <-res.Subscription:
		if !ok {
			removeSubscription(id)
			return returnGoC(0, "", "") // closed
		}
		return marshalJSONToGoCResult(msg)
	case <-time.After(time.Second):
		return returnGoC(1, errTimeoutSubscription, "")
	}
}

func CloseSubscription(id string) GoCResult {
	removeSubscription(id)
	return returnGoC(0, "", "")
}

func ExecuteQuery(
	query string,
	identity string,
	txnID uint64,
	operationName string,
	variables string,
) GoCResult {
	ctx := context.Background()
	opts, err := buildRequestOptions(operationName, variables)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = contextWithIdentity(ctx, identity)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, err = contextWithTransaction(ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	res := globalNode.DB.ExecRequest(ctx, query, opts...)

	// The return is either a subscription ID, or a GQL result. The status indicates
	// which: 0 for GQL, 2 for subscription. 1 is not used because this cannot error; the
	// error is part of the GQL result, to be GQL-compliant.
	if res.Subscription != nil {
		id := storeSubscription(res)
		return returnGoC(2, "", id)
	}
	return marshalJSONToGoCResult(res.GQL)
}
