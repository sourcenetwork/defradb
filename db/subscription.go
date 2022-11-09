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

type subscription[T any] struct {
	updateEvt events.Subscription[client.UpdateEvent]
	requests  []*request.ObjectSubscription[T]
	syncLock  sync.Mutex
}

func (db *db) runSubscriptions(ctx context.Context) {
	log.Info(ctx, "Starting subscription runner")
	for evt := range db.streams.updateEvt {
		txn, err := db.NewTxn(ctx, false)
		if err != nil {
			log.Error(ctx, err.Error())
			continue
		}
		db.streams.syncLock.Lock()

		planner := planner.New(ctx, db, txn)

		col, err := db.GetCollectionBySchemaID(ctx, evt.SchemaID)
		if err != nil {
			log.Error(ctx, err.Error())
			continue
		}

		// keeping track of the active requests
		subs := db.streams.requests[:0]
		for _, objSub := range db.streams.requests {
			if objSub.Stream.IsClosed() {
				continue
			}
			subs = append(subs, objSub)
			if objSub.Schema == col.Name() {
				objSub.CID = client.Some(evt.Cid.String())
				objSub.DocKeys = client.Some([]string{evt.DocKey})
				result := planner.RunSubscriptionRequest(ctx, objSub)
				if result.Data == nil {
					continue
				}
				objSub.Stream.Write(result)
			}
		}

		// helping the GC
		for i := len(subs); i < len(db.streams.requests); i++ {
			db.streams.requests[i] = nil
		}

		db.streams.requests = subs

		txn.Discard(ctx)
		db.streams.syncLock.Unlock()
	}
}

func (db *db) checkForSubsciptions(r *request.Request) (*events.Publisher[client.GQLResult], error) {
	if len(r.Subscription) > 0 && len(r.Subscription[0].Selections) > 0 {
		s := r.Subscription[0].Selections[0]
		if subRequest, ok := s.(*request.ObjectSubscription[client.GQLResult]); ok {
			stream := events.NewPublisher(make(chan client.GQLResult))
			db.streams.syncLock.Lock()
			subRequest.Stream = stream
			db.streams.requests = append(db.streams.requests, subRequest)
			db.streams.syncLock.Unlock()
			return stream, nil
		}

		return nil, errors.New(fmt.Sprintf("expected ObjectSubscription[client.GQLResult] type but got %T", s))
	}
	return nil, nil
}
