// Copyright 2024 Democratized Data Foundation
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

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
)

func (db *DB) handleMessages(ctx context.Context, sub event.Subscription) {
	docIDQueue := newMergeQueue()
	schemaRootQueue := newMergeQueue()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-sub.Message():
			if !ok {
				return
			}
			switch evt := msg.Data.(type) {
			case event.Merge:
				go func() {
					col, err := getCollectionFromCollectionID(ctx, db, evt.CollectionID)
					if err != nil {
						log.ErrorContextE(
							ctx,
							"Failed to execute merge",
							err,
							corelog.Any("Event", evt))
						return
					}

					if col.Version().IsBranchable {
						// As collection commits link to document composite commits, all events
						// recieved for branchable collections must be processed serially else
						// they may otherwise cause a transaction conflict.
						schemaRootQueue.add(evt.CollectionID)
						defer schemaRootQueue.done(evt.CollectionID)
					} else {
						// ensure only one merge per docID
						docIDQueue.add(evt.DocID)
						defer docIDQueue.done(evt.DocID)
					}

					// retry the merge process if a conflict occurs
					//
					// conficts occur when a user updates a document
					// while a merge is in progress.
					for i := 0; i < db.MaxTxnRetries(); i++ {
						err = db.executeMerge(ctx, col, evt)
						if errors.Is(err, corekv.ErrTxnConflict) {
							continue // retry merge
						}
						break // merge success or error
					}

					if err != nil {
						log.ErrorContextE(
							ctx,
							"Failed to execute merge",
							err,
							corelog.Any("Event", evt))
					}
				}()
			}
		}
	}
}
