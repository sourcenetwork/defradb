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
	"sync"

	"github.com/google/uuid"

	"github.com/sourcenetwork/defradb/client"
)

// We cannot return a channel to/from C, so instead we have a map of subscription IDs to
// Subscription objects. Three functions, storeSubscription, getSubscription, and removeSubscription are
// helpers which manage this store behind the scenes, while PollSubscription and CloseSubscription
// are made available to the user for interacting with the subscriptions.

var subscriptionStore sync.Map // map[string]*Subscription

// Subscription is a wrapper for a GraphQL subscription query that also contains a function to
// cancel the context when the subscription is closed. This is used to allow us to avoid leaking
// goroutines.
type Subscription struct {
	ctxCancel  context.CancelFunc
	resultChan <-chan client.GQLResult
}

// Using UUID lets us avoid collisions, even if we use this across multiple nodes
func storeSubscription(s *Subscription) string {
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

// PollSubscription will get the subscription object associcated with an ID, and if
// it exists will see if there's a message in its result channel. If there isn't, it will
// return with status 2, and a blank payload. If there is, it will return with status 0,
// and the payload of the message. If an error occurs, status 1 is returned.
//
//export PollSubscription
func PollSubscription(id *C.char) C.Result {
	subID := C.GoString(id)
	sub, ok := getSubscription(subID)
	if !ok {
		return returnC(returnGoC(1, NewErrInvalidSubscriptionID(subID).Error(), ""))
	}
	select {
	case msg, ok := <-sub.resultChan:
		if !ok {
			removeSubscription(subID)
			return returnC(returnGoC(1, errGettingSubscription, ""))
		}
		return returnC(marshalJSONToGoCResult(msg))
	default:
		return returnC(returnGoC(2, "", ""))
	}
}

//export CloseSubscription
func CloseSubscription(id *C.char) C.Result {
	removeSubscription(C.GoString(id))
	return returnC(returnGoC(0, "", ""))
}

//export ExecuteQuery
func ExecuteQuery(
	nodePtr C.uintptr_t,
	query *C.char,
	identityPtr C.uintptr_t,
	operationName *C.char,
	variables *C.char,
) C.Result {
	ctx := context.Background()
	opts, err := buildRequestOptions(C.GoString(operationName), C.GoString(variables))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	ctx, err = contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	store, err := getStoreFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	ctx, cancelFunc := context.WithCancel(ctx)

	res := store.ExecRequest(ctx, C.GoString(query), opts...)
	sub := &Subscription{
		ctxCancel:  cancelFunc,
		resultChan: res.Subscription,
	}
	// The return is either a subscription ID, or a GQL result. The status indicates
	// which: 0 for GQL, 2 for subscription. 1 is not used because this cannot error; the
	// error is part of the GQL result, to be GQL-compliant.
	if res.Subscription != nil {
		id := storeSubscription(sub)
		return returnC(returnGoC(2, "", id))
	}
	return returnC(marshalJSONToGoCResult(res.GQL))
}
