// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/events"
	"github.com/sourcenetwork/defradb/internal/planner"
)

// handleSubscription checks for a subscription within the given request and
// starts a new go routine that will return all subscription results on the returned
// channel. If a subscription does not exist on the given request nil will be returned.
func (db *db) handleSubscription(ctx context.Context, r *request.Request) (<-chan client.GQLResult, error) {
	if len(r.Subscription) == 0 || len(r.Subscription[0].Selections) == 0 {
		return nil, nil // This is not a subscription request and we have nothing to do here
	}
	if !db.events.Updates.HasValue() {
		return nil, ErrSubscriptionsNotAllowed
	}
	selections := r.Subscription[0].Selections[0]
	subRequest, ok := selections.(*request.ObjectSubscription)
	if !ok {
		return nil, client.NewErrUnexpectedType[request.ObjectSubscription]("SubscriptionSelection", selections)
	}
	// unsubscribing from this publisher will cause a race condition
	// https://github.com/sourcenetwork/defradb/issues/2687
	pub, err := events.NewPublisher(db.events.Updates.Value(), 5)
	if err != nil {
		return nil, err
	}

	resCh := make(chan client.GQLResult)
	go func() {
		defer close(resCh)

		// listen for events and send to the result channel
		for {
			var evt events.Update
			select {
			case val := <-pub.Event():
				evt = val
			case <-ctx.Done():
				return // context cancelled
			}

			txn, err := db.NewTxn(ctx, false)
			if err != nil {
				log.ErrorContext(ctx, err.Error())
				continue
			}

			ctx := SetContextTxn(ctx, txn)
			identity := GetContextIdentity(ctx)

			p := planner.New(ctx, identity, db.acp, db, txn)
			s := subRequest.ToSelect(evt.DocID, evt.Cid.String())

			result, err := p.RunSubscriptionRequest(ctx, s)
			if err == nil && len(result) == 0 {
				txn.Discard(ctx)
				continue // Don't send anything back to the client if the request yields an empty dataset.
			}
			res := client.GQLResult{
				Data: result,
			}
			if err != nil {
				res.Errors = []error{err}
			}

			select {
			case resCh <- res:
				txn.Discard(ctx)
			case <-ctx.Done():
				txn.Discard(ctx)
				return // context cancelled
			}
		}
	}()

	return resCh, nil
}
