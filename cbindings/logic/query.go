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

type Subscription struct {
	ctxCancel  context.CancelFunc
	resultChan <-chan client.GQLResult
}

// Using UUID lets us avoid collisions, even if we use this across multiple nodes
func storeSubscription(s Subscription) string {
	id := uuid.NewString()
	subscriptionStore.Store(id, s)
	return id
}

func getSubscription(id string) (*Subscription, bool) {
	val, ok := subscriptionStore.Load(id)
	if !ok {
		return nil, false
	}
	//nolint:forcetypeassert
	return val.(*Subscription), true
}

func removeSubscription(id string) {
	val, ok := subscriptionStore.LoadAndDelete(id)
	if ok {
		//nolint:forcetypeassert
		sub := val.(*Subscription)
		sub.ctxCancel()
	}
}

func PollSubscription(id string) GoCResult {
	sub, ok := getSubscription(id)
	if !ok {
		return returnGoC(1, errInvalidSubscriptionID, "")
	}

	select {
	case msg, ok := <-sub.resultChan:
		if !ok {
			removeSubscription(id)
			return returnGoC(0, "", "")
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
	n int,
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

	ctx, err = contextWithTransaction(n, ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx, cancelFunc := context.WithCancel(ctx)
	res := GetNode(n).DB.ExecRequest(ctx, query, opts...)
	sub := &Subscription{
		ctxCancel:  cancelFunc,
		resultChan: res.Subscription,
	}
	// The return is either a subscription ID, or a GQL result. The status indicates
	// which: 0 for GQL, 2 for subscription. 1 is not used because this cannot error; the
	// error is part of the GQL result, to be GQL-compliant.
	if res.Subscription != nil {
		id := storeSubscription(*sub)
		return returnGoC(2, "", id)
	}
	return marshalJSONToGoCResult(res.GQL)
}
