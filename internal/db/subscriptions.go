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
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/extensions"
	"github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/internal/planner"
)

type subscriptionSelector interface {
	ToSubscriptionSelect(docID, cid string) request.Selection
	CheckCIDFilter(cid string) bool
	CheckDocIDFilter(docID string) bool
	CheckCollectionFilter(collectionID string) bool
}

// targetCollectionSetter is implemented by subscription selectors whose
// target collection can be resolved at subscribe-time. The subscription
// handler uses this to inject the root CollectionID so CheckCollectionFilter
// can drop events from other collections without opening a transaction.
type targetCollectionSetter interface {
	SetTargetCollectionID(id string)
}

// handleSubscription checks for a subscription within the given request and
// starts a new go routine that will return all subscription results on the returned
// channel. If a subscription does not exist on the given request nil will be returned.
func (db *DB) handleSubscription(ctx context.Context, r *request.Request) (<-chan client.GQLResult, error) {
	if len(r.Subscription) == 0 || len(r.Subscription[0].Selections) == 0 {
		return nil, nil // This is not a subscription request and we have nothing to do here
	}
	subRequest, ok := r.Subscription[0].Selections[0].(subscriptionSelector)
	if !ok {
		return nil, client.NewErrUnexpectedType[request.Selection]("SubscriptionSelection", subRequest)
	}

	// Resolve the target collection's root CollectionID once at subscribe-time
	// so the event loop can reject events from other collections without
	// opening a transaction. If the target isn't a request.Select we skip the
	// stamping; CheckCollectionFilter will default to "allow" for such cases.
	if setter, ok := subRequest.(targetCollectionSetter); ok {
		selection, ok := r.Subscription[0].Selections[0].(*request.Select)
		if !ok {
			return nil, client.NewErrUnexpectedType[request.Selection]("SubscriptionSelection", subRequest)
		}
		col, err := db.GetCollectionByName(ctx, selection.Name)
		if err != nil {
			return nil, err
		}
		setter.SetTargetCollectionID(col.Version().CollectionID)
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
			// Skip events that do not pass the subscription's collection, docID,
			// and cid filters. This is an optimization to avoid running the
			// selection planner and related query logic when we know the event
			// will not be relevant to the subscription.
			if !subRequest.CheckCollectionFilter(evt.CollectionID) ||
				!subRequest.CheckDocIDFilter(evt.DocID) ||
				!subRequest.CheckCIDFilter(evt.Cid.String()) {
				continue
			}
			txn, err := db.NewTxn(false)
			if err != nil {
				log.ErrorContextE(ctx, "Failed to create transaction for subscription", err)
				continue
			}
			ctx := InitContext(ctx, txn)

			// One accumulator per event, not per subscription. A shared one would
			// make each event repeat every warning before it.
			ctx = extensions.WithAccumulator(ctx)

			p := planner.New(
				ctx,
				identity.FromContext(ctx),
				db.nodeACP,
				db.documentACP,
				db,
				db.p2p,
				db.getLensStore(ctx),
				db.collectionRepository,
			)
			s := subRequest.ToSubscriptionSelect(evt.DocID, evt.Cid.String())

			result, err := p.RunSelection(ctx, s)
			if err == nil && len(result) == 0 {
				txn.Discard()
				continue // Don't send anything back to the client if the request yields an empty dataset.
			}

			res := client.GQLResult{}

			// This approach will only support return types that are []map[string]any
			// (ie docs) for results. So top level aggregates, or other top level fields
			// that we would want to add to subscriptions that don't return
			// docs currently will not work.
			for op, data := range result {
				resultSlice, ok := data.([]map[string]any)
				if !ok {
					res.Errors = append(res.Errors, ErrBadDocsResultType)
				}

				if len(resultSlice) == 0 {
					delete(result, op)
				}
			}

			// ignore incorrect CID for DocID error. This is specific to
			// subscription API. Only the DocID is externally configurable for
			// this API, but the CID comes from the event, which means theres a
			// high likely hood of CID/DocID mismatch, so we need to ignore it
			// to falsely report errors to the subscription.
			if err != nil && !errors.Is(err, planner.ErrIncorrectOrMissingCID) {
				res.Errors = append(res.Errors, err)
			}

			// now that weve filtered empty result sets, lets recheck. If the result is
			// empty AND there is no error to forward, drop. Otherwise (e.g. RunSelection
			// returned (empty, err)) deliver the GQLResult so the subscriber sees the
			// error instead of timing out.
			if len(result) == 0 && len(res.Errors) == 0 {
				txn.Discard()
				continue
			}
			res.Data = result
			res.Extensions = extensions.Collect(ctx)

			select {
			case <-ctx.Done():
				txn.Discard()
				return // context cancelled
			case resCh <- res:
				txn.Discard()
			}
		}
	}()

	return resCh, nil
}
