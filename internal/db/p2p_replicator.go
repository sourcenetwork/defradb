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
	"encoding/json"

	dsq "github.com/ipfs/go-datastore/query"

	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/merkle/clock"
)

func (db *db) SetReplicator(ctx context.Context, rep client.Replicator) error {
	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	db.peerMutex.RLock()
	if db.peerInfo.HasValue() && rep.Info.ID == db.peerInfo.Value().ID {
		db.peerMutex.RUnlock()
		return ErrSelfTargetForReplicator
	}
	db.peerMutex.RUnlock()
	if err := rep.Info.ID.Validate(); err != nil {
		return err
	}

	// TODO-ACP: Support ACP <> P2P - https://github.com/sourcenetwork/defradb/issues/2366
	// ctx = db.SetContextIdentity(ctx, identity)
	ctx = SetContextTxn(ctx, txn)

	oldSchemas := make(map[string]struct{})
	repKey := core.NewReplicatorKey(rep.Info.ID.String())
	hasOldRep, err := txn.Systemstore().Has(ctx, repKey.ToDS())
	if err != nil {
		return err
	}
	if hasOldRep {
		repBytes, err := txn.Systemstore().Get(ctx, repKey.ToDS())
		if err != nil {
			return err
		}
		oldRep := client.Replicator{}
		err = json.Unmarshal(repBytes, &oldRep)
		if err != nil {
			return err
		}
		for _, schema := range oldRep.Schemas {
			oldSchemas[schema] = struct{}{}
		}
	}

	var collections []client.Collection
	switch {
	case len(rep.Schemas) > 0:
		// if specific collections are chosen get them by name
		for _, name := range rep.Schemas {
			col, err := db.GetCollectionByName(ctx, name)
			if err != nil {
				return NewErrReplicatorCollections(err)
			}

			if col.Description().Policy.HasValue() {
				return ErrReplicatorColHasPolicy
			}

			collections = append(collections, col)
		}

	default:
		// default to all collections (unless a collection contains a policy).
		// TODO-ACP: default to all collections after resolving https://github.com/sourcenetwork/defradb/issues/2366
		allCollections, err := db.GetCollections(ctx, client.CollectionFetchOptions{})
		if err != nil {
			return NewErrReplicatorCollections(err)
		}

		for _, col := range allCollections {
			// Can not default to all collections if any collection has a policy.
			// TODO-ACP: remove this check/loop after https://github.com/sourcenetwork/defradb/issues/2366
			if col.Description().Policy.HasValue() {
				return ErrReplicatorSomeColsHavePolicy
			}
		}
		collections = allCollections
	}

	rep.Schemas = nil
	schemaMap := make(map[string]struct{})
	addedCols := []client.Collection{}
	for _, col := range collections {
		rep.Schemas = append(rep.Schemas, col.SchemaRoot())
		schemaMap[col.SchemaRoot()] = struct{}{}
		if _, ok := oldSchemas[col.SchemaRoot()]; !ok {
			addedCols = append(addedCols, col)
		}
	}

	// persist replicator to the datastore
	newRepBytes, err := json.Marshal(rep)
	if err != nil {
		return err
	}

	err = txn.Systemstore().Put(ctx, repKey.ToDS(), newRepBytes)
	if err != nil {
		return err
	}

	txn.OnSuccess(func() {
		db.events.Publish(event.NewMessage(event.ReplicatorName, event.Replicator{
			Info:    rep.Info,
			Schemas: schemaMap,
			Docs:    db.getDocsHeads(ctx, addedCols),
		}))
	})

	return txn.Commit(ctx)
}

func (db *db) getDocsHeads(
	ctx context.Context,
	cols []client.Collection,
) <-chan event.Update {
	updateChan := make(chan event.Update)
	go func() {
		defer close(updateChan)
		txn, err := db.NewTxn(ctx, true)
		if err != nil {
			log.ErrorContextE(ctx, "Failed to get transaction", err)
			return
		}
		defer txn.Discard(ctx)
		ctx = SetContextTxn(ctx, txn)
		for _, col := range cols {
			keysCh, err := col.GetAllDocIDs(ctx)
			if err != nil {
				log.ErrorContextE(
					ctx,
					"Failed to get all docIDs",
					NewErrReplicatorDocID(err, errors.NewKV("Collection", col.Name().Value())),
				)
				continue
			}
			for docIDResult := range keysCh {
				if docIDResult.Err != nil {
					log.ErrorContextE(ctx, "Key channel error", docIDResult.Err)
					continue
				}
				docID := core.DataStoreKeyFromDocID(docIDResult.ID)
				headset := clock.NewHeadSet(
					txn.Headstore(),
					docID.WithFieldId(core.COMPOSITE_NAMESPACE).ToHeadStoreKey(),
				)
				cids, _, err := headset.List(ctx)
				if err != nil {
					log.ErrorContextE(
						ctx,
						"Failed to get heads",
						err,
						corelog.String("DocID", docIDResult.ID.String()),
						corelog.Any("Collection", col.Name()))
					continue
				}
				// loop over heads, get block, make the required logs, and send
				for _, c := range cids {
					blk, err := txn.Blockstore().Get(ctx, c)
					if err != nil {
						log.ErrorContextE(ctx, "Failed to get block", err,
							corelog.Any("CID", c),
							corelog.Any("Collection", col.Name()))
						continue
					}

					updateChan <- event.Update{
						DocID:      docIDResult.ID.String(),
						Cid:        c,
						SchemaRoot: col.SchemaRoot(),
						Block:      blk.RawData(),
					}
				}
			}
		}
	}()

	return updateChan
}

func (db *db) DeleteReplicator(ctx context.Context, rep client.Replicator) error {
	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	if err := rep.Info.ID.Validate(); err != nil {
		return err
	}

	// set transaction for all operations
	ctx = SetContextTxn(ctx, txn)

	oldSchemas := make(map[string]struct{})
	repKey := core.NewReplicatorKey(rep.Info.ID.String())
	hasOldRep, err := txn.Systemstore().Has(ctx, repKey.ToDS())
	if err != nil {
		return err
	}
	if hasOldRep {
		repBytes, err := txn.Systemstore().Get(ctx, repKey.ToDS())
		if err != nil {
			return err
		}
		oldRep := client.Replicator{}
		err = json.Unmarshal(repBytes, &oldRep)
		if err != nil {
			return err
		}
		for _, schema := range oldRep.Schemas {
			oldSchemas[schema] = struct{}{}
		}
	}

	var collections []client.Collection
	if len(rep.Schemas) > 0 {
		// if specific collections are chosen get them by name
		for _, name := range rep.Schemas {
			col, err := db.GetCollectionByName(ctx, name)
			if err != nil {
				return NewErrReplicatorCollections(err)
			}
			collections = append(collections, col)
		}
		// make sure the replicator exists in the datastore
		key := core.NewReplicatorKey(rep.Info.ID.String())
		_, err = txn.Systemstore().Get(ctx, key.ToDS())
		if err != nil {
			return err
		}
	} else {
		oldSchemas = make(map[string]struct{})
	}

	rep.Schemas = nil
	for _, col := range collections {
		delete(oldSchemas, col.SchemaRoot())
	}
	for schema := range oldSchemas {
		rep.Schemas = append(rep.Schemas, schema)
	}

	// persist the replicator to the store, deleting it if no schemas remain
	key := core.NewReplicatorKey(rep.Info.ID.String())
	if len(rep.Schemas) == 0 {
		err := txn.Systemstore().Delete(ctx, key.ToDS())
		if err != nil {
			return err
		}
	} else {
		repBytes, err := json.Marshal(rep)
		if err != nil {
			return err
		}
		err = txn.Systemstore().Put(ctx, key.ToDS(), repBytes)
		if err != nil {
			return err
		}
	}

	txn.OnSuccess(func() {
		db.events.Publish(event.NewMessage(event.ReplicatorName, event.Replicator{
			Info:    rep.Info,
			Schemas: oldSchemas,
		}))
	})

	return txn.Commit(ctx)
}

func (db *db) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	txn, err := db.NewTxn(ctx, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	// create collection system prefix query
	query := dsq.Query{
		Prefix: core.NewReplicatorKey("").ToString(),
	}
	results, err := txn.Systemstore().Query(ctx, query)
	if err != nil {
		return nil, err
	}

	var reps []client.Replicator
	for result := range results.Next() {
		var rep client.Replicator
		if err = json.Unmarshal(result.Value, &rep); err != nil {
			return nil, err
		}
		reps = append(reps, rep)
	}
	return reps, nil
}

func (db *db) loadReplicators(ctx context.Context) error {
	replicators, err := db.GetAllReplicators(ctx)
	if err != nil {
		return err
	}

	for _, rep := range replicators {
		schemaMap := make(map[string]struct{})
		for _, schema := range rep.Schemas {
			schemaMap[schema] = struct{}{}
		}
		db.events.Publish(event.NewMessage(event.ReplicatorName, event.Replicator{
			Info:    rep.Info,
			Schemas: schemaMap,
		}))
	}
	return nil
}
