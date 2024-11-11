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
	"sync"

	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
)

func (db *db) handleMessages(ctx context.Context, sub *event.Subscription) {
	docIdQueue := newMergeQueue()
	schemaRootQueue := newMergeQueue()

	// This is used to ensure we only trigger loadAndPublishP2PCollections and loadAndPublishReplicators
	// once per db instanciation.
	loadOnce := sync.Once{}
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
					col, err := getCollectionFromRootSchema(ctx, db, evt.SchemaRoot)
					if err != nil {
						log.ErrorContextE(
							ctx,
							"Failed to execute merge",
							err,
							corelog.Any("Event", evt))
						return
					}

					if col.Description().IsBranchable {
						// As collection commits link to document composite commits, all events
						// recieved for branchable collections must be processed serially else
						// they may otherwise cause a transaction conflict.
						schemaRootQueue.add(evt.SchemaRoot)
						defer schemaRootQueue.done(evt.SchemaRoot)
					} else {
						// ensure only one merge per docID
						docIdQueue.add(evt.DocID)
						defer docIdQueue.done(evt.DocID)
					}

					// retry the merge process if a conflict occurs
					//
					// conficts occur when a user updates a document
					// while a merge is in progress.
					for i := 0; i < db.MaxTxnRetries(); i++ {
						err = db.executeMerge(ctx, col, evt)
						if errors.Is(err, datastore.ErrTxnConflict) {
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
			case event.PeerInfo:
				db.peerInfo.Store(evt.Info)
				// Load and publish P2P collections and replicators once per db instance start.
				// A Go routine is used to ensure the message handler is not blocked by these potentially
				// long running operations.
				go loadOnce.Do(func() {
					err := db.loadAndPublishP2PCollections(ctx)
					if err != nil {
						log.ErrorContextE(ctx, "Failed to load P2P collections", err)
					}

					err = db.loadAndPublishReplicators(ctx)
					if err != nil {
						log.ErrorContextE(ctx, "Failed to load replicators", err)
					}
				})
			case event.ReplicatorFailure:
				// ReplicatorFailure is a notification that a replicator has failed to replicate a document.
				err := db.handleReplicatorFailure(ctx, evt.PeerID.String(), evt.DocID)
				if err != nil {
					log.ErrorContextE(ctx, "Failed to handle replicator failure", err)
				}
			}
		}
	}
}
