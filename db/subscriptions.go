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
	"fmt"
	"sync"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/events"
	"github.com/sourcenetwork/defradb/planner"
)

type subscriptions struct {
	updateEvt events.Subscription[client.UpdateEvent]
	requests  []*request.ObjectSubscription
	syncLock  sync.Mutex
}

func (db *db) handleClientSubscriptions(ctx context.Context) {
	if db.clientSubscriptions == nil {
		log.Info(ctx, "can't run subscription without adding the option to the db")
		return
	}

	log.Info(ctx, "Starting client subscription handler")
	for evt := range db.clientSubscriptions.updateEvt {
		db.clientSubscriptions.syncLock.Lock()
		if len(db.clientSubscriptions.requests) == 0 {
			db.clientSubscriptions.syncLock.Unlock()
			continue
		}

		txn, err := db.NewTxn(ctx, false)
		if err != nil {
			log.Error(ctx, err.Error())
			db.clientSubscriptions.syncLock.Unlock()
			continue
		}

		planner := planner.New(ctx, db, txn)

		col, err := db.GetCollectionBySchemaID(ctx, evt.SchemaID)
		if err != nil {
			log.Error(ctx, err.Error())
			db.clientSubscriptions.syncLock.Unlock()
			continue
		}

		// keeping track of the active requests
		subs := db.clientSubscriptions.requests[:0]
		for _, objSub := range db.clientSubscriptions.requests {
			if objSub.Stream.IsClosed() {
				continue
			}
			subs = append(subs, objSub)
			if objSub.Schema == col.Name() {
				objSub.CID = client.Some(evt.Cid.String())
				objSub.DocKeys = client.Some([]string{evt.DocKey})
				result, err := planner.RunSubscriptionRequest(ctx, objSub)
				if err != nil {
					objSub.Stream.Write(client.GQLResult{
						Errors: []any{err.Error()},
					})
				}

				// Don't send anything back to the client if the request yields an empty dataset.
				if len(result) == 0 {
					continue
				}

				objSub.Stream.Write(client.GQLResult{
					Data: result,
				})
			}
		}

		// helping the GC
		for i := len(subs); i < len(db.clientSubscriptions.requests); i++ {
			db.clientSubscriptions.requests[i] = nil
		}

		db.clientSubscriptions.requests = subs

		txn.Discard(ctx)
		db.clientSubscriptions.syncLock.Unlock()
	}
}

func (db *db) checkForClientSubsciptions(r *request.Request) (*events.Publisher, error) {
	if len(r.Subscription) > 0 && len(r.Subscription[0].Selections) > 0 {
		s := r.Subscription[0].Selections[0]
		if subRequest, ok := s.(*request.ObjectSubscription); ok {
			if db.clientSubscriptions == nil {
				return nil, errors.New("server does not accept subscriptions")
			}

			stream := events.NewPublisher(make(chan any))
			db.clientSubscriptions.syncLock.Lock()
			subRequest.Stream = stream
			db.clientSubscriptions.requests = append(db.clientSubscriptions.requests, subRequest)
			db.clientSubscriptions.syncLock.Unlock()
			return stream, nil
		}

		return nil, errors.New(fmt.Sprintf("expected ObjectSubscription[client.GQLResult] type but got %T", s))
	}
	return nil, nil
}
