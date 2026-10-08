// Copyright 2026 Democratized Data Foundation
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
