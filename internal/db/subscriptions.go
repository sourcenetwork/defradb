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

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/planner"
)

// handleSubscription checks for a subscription within the given request and
// starts a new go routine that will return all subscription results on the returned
// channel. If a subscription does not exist on the given request nil will be returned.
func (db *db) handleSubscription(ctx context.Context, r *request.Request) (<-chan client.GQLResult, error) {
	if len(r.Subscription) == 0 || len(r.Subscription[0].Selections) == 0 {
		return nil, nil // This is not a subscription request and we have nothing to do here
	}
	selections := r.Subscription[0].Selections[0]
	subRequest, ok := selections.(*request.ObjectSubscription)
	if !ok {
		return nil, client.NewErrUnexpectedType[request.ObjectSubscription]("SubscriptionSelection", selections)
	}
	sub, err := db.events.Subscribe(event.UpdateName)
	if err != nil {
		return nil, err
	}
	resCh := make(chan client.GQLResult)
	go func() {
		defer func() {
			db.events.Unsubscribe(sub)
			close(resCh)
		}()

		// listen for events and send to the result channel
		for {
			var evt event.Update
			select {
			case <-ctx.Done():
				return // context cancelled
			case val, ok := <-sub.Message():
				if !ok {
					return // channel closed
				}
				evt, ok = val.Data.(event.Update)
				if !ok {
					continue // invalid event value
				}
			}

			txn, err := db.NewTxn(ctx, false)
			if err != nil {
				log.ErrorContext(ctx, err.Error())
				continue
			}

			ctx := SetContextTxn(ctx, txn)

			p := planner.New(ctx, identity.FromContext(ctx), db.acp, db, txn)
			s := subRequest.ToSelect(evt.DocID, evt.Cid.String())

			result, err := p.RunSelection(ctx, s)
			if err == nil && len(result) == 0 {
				txn.Discard(ctx)
				continue // Don't send anything back to the client if the request yields an empty dataset.
			}
			res := client.GQLResult{}
			if err != nil {
				res.Errors = append(res.Errors, err)
			}
			res.Data = result

			select {
			case <-ctx.Done():
				txn.Discard(ctx)
				return // context cancelled
			case resCh <- res:
				txn.Discard(ctx)
			}
		}
	}()

	return resCh, nil
}
