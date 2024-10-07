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
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-datastore/query"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/merkle/clock"
)

const (
	retryLoopInterval = 2 * time.Second
	retryTimeout      = 10 * time.Second
)

func (db *db) SetReplicator(ctx context.Context, rep client.Replicator) error {
	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	if err := rep.Info.ID.Validate(); err != nil {
		return err
	}

	peerInfo := peer.AddrInfo{}
	if info := db.peerInfo.Load(); info != nil {
		peerInfo = info.(peer.AddrInfo)
	}
	if rep.Info.ID == peerInfo.ID {
		return ErrSelfTargetForReplicator
	}

	ctx = SetContextTxn(ctx, txn)

	storedRep := client.Replicator{}
	storedSchemas := make(map[string]struct{})
	repKey := core.NewReplicatorKey(rep.Info.ID.String())
	hasOldRep, err := txn.Peerstore().Has(ctx, repKey.ToDS())
	if err != nil {
		return err
	}
	if hasOldRep {
		repBytes, err := txn.Peerstore().Get(ctx, repKey.ToDS())
		if err != nil {
			return err
		}
		err = json.Unmarshal(repBytes, &storedRep)
		if err != nil {
			return err
		}
		for _, schema := range storedRep.Schemas {
			storedSchemas[schema] = struct{}{}
		}
	} else {
		storedRep.Info = rep.Info
		storedRep.LastStatusChange = time.Now()
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

			collections = append(collections, col)
		}

	default:
		collections, err = db.GetCollections(ctx, client.CollectionFetchOptions{})
		if err != nil {
			return NewErrReplicatorCollections(err)
		}
	}

	if db.acp.HasValue() && !db.acp.Value().SupportsP2P() {
		for _, col := range collections {
			if col.Description().Policy.HasValue() {
				return ErrReplicatorColHasPolicy
			}
		}
	}

	addedCols := []client.Collection{}
	for _, col := range collections {
		if _, ok := storedSchemas[col.SchemaRoot()]; !ok {
			storedSchemas[col.SchemaRoot()] = struct{}{}
			addedCols = append(addedCols, col)
			storedRep.Schemas = append(storedRep.Schemas, col.SchemaRoot())
		}
	}

	// persist replicator to the datastore
	newRepBytes, err := json.Marshal(storedRep)
	if err != nil {
		return err
	}

	err = txn.Peerstore().Put(ctx, repKey.ToDS(), newRepBytes)
	if err != nil {
		return err
	}

	txn.OnSuccess(func() {
		db.events.Publish(event.NewMessage(event.ReplicatorName, event.Replicator{
			Info:    rep.Info,
			Schemas: storedSchemas,
			Docs:    db.getDocsHeads(context.Background(), addedCols),
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
					docID.WithFieldID(core.COMPOSITE_NAMESPACE).ToHeadStoreKey(),
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

	storedRep := client.Replicator{}
	storedSchemas := make(map[string]struct{})
	repKey := core.NewReplicatorKey(rep.Info.ID.String())
	hasOldRep, err := txn.Peerstore().Has(ctx, repKey.ToDS())
	if err != nil {
		return err
	}
	if !hasOldRep {
		return ErrReplicatorNotFound
	}
	repBytes, err := txn.Peerstore().Get(ctx, repKey.ToDS())
	if err != nil {
		return err
	}
	err = json.Unmarshal(repBytes, &storedRep)
	if err != nil {
		return err
	}
	for _, schema := range storedRep.Schemas {
		storedSchemas[schema] = struct{}{}
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
		_, err = txn.Peerstore().Get(ctx, key.ToDS())
		if err != nil {
			return err
		}
	} else {
		storedSchemas = make(map[string]struct{})
	}

	for _, col := range collections {
		delete(storedSchemas, col.SchemaRoot())
	}
	// Update the list of schemas for this replicator prior to persisting.
	storedRep.Schemas = []string{}
	for schema := range storedSchemas {
		storedRep.Schemas = append(storedRep.Schemas, schema)
	}

	// Persist the replicator to the store, deleting it if no schemas remain
	key := core.NewReplicatorKey(rep.Info.ID.String())
	if len(rep.Schemas) == 0 {
		err := txn.Peerstore().Delete(ctx, key.ToDS())
		if err != nil {
			return err
		}
	} else {
		repBytes, err := json.Marshal(rep)
		if err != nil {
			return err
		}
		err = txn.Peerstore().Put(ctx, key.ToDS(), repBytes)
		if err != nil {
			return err
		}
	}

	txn.OnSuccess(func() {
		db.events.Publish(event.NewMessage(event.ReplicatorName, event.Replicator{
			Info:    rep.Info,
			Schemas: storedSchemas,
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
	query := query.Query{
		Prefix: core.NewReplicatorKey("").ToString(),
	}
	results, err := txn.Peerstore().Query(ctx, query)
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

func (db *db) loadAndPublishReplicators(ctx context.Context) error {
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

// retryStatus is used to communicate if the retry was successful or not.
type retryStatus struct {
	PeerID  string
	Success bool
}

// handleReplicatorRetries manages retries for failed replication attempts.
func (db *db) handleReplicatorRetries(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case r := <-db.retryChan:
			err := db.handleReplicatorFailure(ctx, r)
			if err != nil {
				log.ErrorContextE(ctx, "Failed to handle replicator failure", err)
			}

		case r := <-db.retryDone:
			err := db.handleCompletedReplicatorRetry(ctx, r)
			if err != nil {
				log.ErrorContextE(ctx, "Failed to handle completed replicator retry", err)
			}

		case <-time.After(retryLoopInterval):
			db.retryReplicators(ctx)
		}
	}
}

func (db *db) handleReplicatorFailure(ctx context.Context, r event.ReplicatorFailure) error {
	err := db.updateReplicatorStatus(ctx, r.PeerID.String(), false)
	if err != nil {
		return err
	}
	err = db.createIfNotExistsReplicatorRetry(ctx, r.PeerID.String())
	if err != nil {
		return err
	}
	docIDKey := core.NewReplicatorRetryDocIDKey(r.PeerID.String(), r.DocID)
	err = db.Peerstore().Put(ctx, docIDKey.ToDS(), []byte{})
	if err != nil {
		return err
	}
	return nil
}

func (db *db) handleCompletedReplicatorRetry(ctx context.Context, r retryStatus) error {
	if r.Success {
		done, err := db.deleteReplicatorRetryIfNoMoreDocs(ctx, r.PeerID)
		if err != nil {
			return err
		}
		if done {
			err := db.updateReplicatorStatus(ctx, r.PeerID, true)
			if err != nil {
				return err
			}
		}
	} else {
		err := db.setReplicatorNextRetry(ctx, r.PeerID)
		if err != nil {
			return err
		}
	}
	return nil
}

// updateReplicatorStatus updates the status of a replicator in the peerstore.
func (db *db) updateReplicatorStatus(ctx context.Context, peerID string, active bool) error {
	key := core.NewReplicatorKey(peerID)
	repBytes, err := db.Peerstore().Get(ctx, key.ToDS())
	if err != nil {
		return err
	}
	rep := client.Replicator{}
	err = json.Unmarshal(repBytes, &rep)
	if err != nil {
		return err
	}
	switch active {
	case true:
		rep.Status = client.ReplicatorStatusActive
		if rep.Status == client.ReplicatorStatusInactive {
			rep.LastStatusChange = time.Time{}
		}
	case false:
		rep.Status = client.ReplicatorStatusInactive
		if rep.Status == client.ReplicatorStatusActive {
			rep.LastStatusChange = time.Now()
		}
	}
	b, err := json.Marshal(rep)
	if err != nil {
		return err
	}
	return db.Peerstore().Put(ctx, key.ToDS(), b)
}

type retryInfo struct {
	NextRetry  time.Time
	NumRetries int
	Retrying   bool
}

func (db *db) createIfNotExistsReplicatorRetry(ctx context.Context, peerID string) error {
	key := core.NewReplicatorRetryIDKey(peerID)
	exists, err := db.Peerstore().Has(ctx, key.ToDS())
	if err != nil {
		return err
	}
	if !exists {
		r := retryInfo{
			NextRetry:  time.Now().Add(db.retryIntervals[0]),
			NumRetries: 0,
		}
		b, err := cbor.Marshal(r)
		if err != nil {
			return err
		}
		err = db.Peerstore().Put(ctx, key.ToDS(), b)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (db *db) retryReplicators(ctx context.Context) {
	q := query.Query{
		Prefix: core.REPLICATOR_RETRY_ID,
	}
	results, err := db.Peerstore().Query(ctx, q)
	if err != nil {
		log.ErrorContextE(ctx, "Failed to query replicator retries", err)
		return
	}
	defer closeQueryResults(results)
	now := time.Now()
	for result := range results.Next() {
		rInfo := retryInfo{}
		err = cbor.Unmarshal(result.Value, &rInfo)
		if err != nil {
			log.ErrorContextE(ctx, "Failed to unmarshal replicator retry info", err)
			continue
		}
		// If the next retry time has passed and the replicator is not already retrying.
		if now.After(rInfo.NextRetry) && !rInfo.Retrying {
			key, err := core.NewReplicatorRetryIDKeyFromString(result.Key)
			if err != nil {
				log.ErrorContextE(ctx, "Failed to parse replicator retry ID key", err)
				continue
			}

			// The replicator might have been deleted by the time we reach this point.
			// If it no longer exists, we delete the retry key and all retry docs.
			exists, err := db.Peerstore().Has(ctx, core.NewReplicatorKey(key.PeerID).ToDS())
			if err != nil {
				log.ErrorContextE(ctx, "Failed to check if replicator exists", err)
				continue
			}
			if !exists {
				err = db.deleteReplicatorRetryAndDocs(ctx, key.PeerID)
				if err != nil {
					log.ErrorContextE(ctx, "Failed to delete replicator retry and docs", err)
				}
				continue
			}

			err = db.setReplicatorAsRetrying(ctx, key, rInfo)
			if err != nil {
				log.ErrorContextE(ctx, "Failed to set replicator as retrying", err)
				continue
			}
			go db.retryReplicator(ctx, key.PeerID)
		}
	}
}

func (db *db) setReplicatorAsRetrying(ctx context.Context, key core.ReplicatorRetryIDKey, rInfo retryInfo) error {
	rInfo.Retrying = true
	rInfo.NumRetries++
	b, err := cbor.Marshal(rInfo)
	if err != nil {
		return err
	}
	return db.Peerstore().Put(ctx, key.ToDS(), b)
}

func (db *db) setReplicatorNextRetry(ctx context.Context, peerID string) error {
	key := core.NewReplicatorRetryIDKey(peerID)
	b, err := db.Peerstore().Get(ctx, key.ToDS())
	if err != nil {
		return err
	}
	rInfo := retryInfo{}
	err = cbor.Unmarshal(b, &rInfo)
	if err != nil {
		return err
	}
	if rInfo.NumRetries >= len(db.retryIntervals) {
		rInfo.NextRetry = time.Now().Add(db.retryIntervals[len(db.retryIntervals)-1])
	} else {
		rInfo.NextRetry = time.Now().Add(db.retryIntervals[rInfo.NumRetries])
	}
	rInfo.Retrying = false
	b, err = cbor.Marshal(rInfo)
	if err != nil {
		return err
	}
	return db.Peerstore().Put(ctx, key.ToDS(), b)
}

func (db *db) retryReplicator(ctx context.Context, peerID string) {
	log.InfoContext(ctx, "Retrying replicator", corelog.String("PeerID", peerID))
	key := core.NewReplicatorRetryDocIDKey(peerID, "")
	q := query.Query{
		Prefix: key.ToString(),
	}
	results, err := db.Peerstore().Query(ctx, q)
	if err != nil {
		log.ErrorContextE(ctx, "Failed to query retry docs", err)
		return
	}
	defer closeQueryResults(results)
	for result := range results.Next() {
		select {
		case <-ctx.Done():
			return
		default:
		}
		key, err := core.NewReplicatorRetryDocIDKeyFromString(result.Key)
		if err != nil {
			log.ErrorContextE(ctx, "Failed to parse retry doc key", err)
			continue
		}
		err = db.retryDoc(ctx, key.DocID)
		if err != nil {
			log.ErrorContextE(ctx, "Failed to retry doc", err)
			db.retryDone <- retryStatus{
				PeerID:  peerID,
				Success: false,
			}
			// if one doc fails, stop retrying the rest and just wait for the next retry
			return
		}
		err = db.Peerstore().Delete(ctx, key.ToDS())
		if err != nil {
			log.ErrorContextE(ctx, "Failed to delete retry docID", err)
		}
	}
	db.retryDone <- retryStatus{
		PeerID:  peerID,
		Success: true,
	}
}

func (db *db) retryDoc(ctx context.Context, docID string) error {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)
	headStoreKey := core.HeadStoreKey{
		DocID:   docID,
		FieldID: core.COMPOSITE_NAMESPACE,
	}
	headset := clock.NewHeadSet(txn.Headstore(), headStoreKey)
	cids, _, err := headset.List(ctx)
	if err != nil {
		return err
	}

	for _, c := range cids {
		select {
		case <-ctx.Done():
			return ErrContextDone
		default:
		}
		rawblk, err := db.Blockstore().Get(ctx, c)
		if err != nil {
			return err
		}
		blk, err := coreblock.GetFromBytes(rawblk.RawData())
		if err != nil {
			return err
		}
		schema, err := db.getSchemaByVersionID(ctx, blk.Delta.GetSchemaVersionID())
		if err != nil {
			return err
		}
		successChan := make(chan bool)
		updateEvent := event.Update{
			DocID:      docID,
			Cid:        c,
			SchemaRoot: schema.Root,
			Block:      rawblk.RawData(),
			IsRetry:    true,
			Success:    successChan,
		}
		db.events.Publish(event.NewMessage(event.UpdateName, updateEvent))

		select {
		case success := <-successChan:
			if !success {
				return errors.New("pushlog failed")
			}
		case <-time.After(retryTimeout):
			return ErrTimeoutDocRetry
		}
	}
	return nil
}

// deleteReplicatorRetryIfNoMoreDocs deletes the replicator retry key if there are no more docs to retry.
// It returns true if there are no more docs to retry, false otherwise.
func (db *db) deleteReplicatorRetryIfNoMoreDocs(ctx context.Context, peerID string) (bool, error) {
	key := core.NewReplicatorRetryDocIDKey(peerID, "")
	q := query.Query{
		Prefix:   key.ToString(),
		KeysOnly: true,
	}
	results, err := db.Peerstore().Query(ctx, q)
	if err != nil {
		return false, err
	}
	defer closeQueryResults(results)
	entries, err := results.Rest()
	if err != nil {
		return false, err
	}
	if len(entries) == 0 {
		key := core.NewReplicatorRetryIDKey(peerID)
		return true, db.Peerstore().Delete(ctx, key.ToDS())
	}
	// If we there are still docs to retry, we run the retry right away.
	go db.retryReplicator(ctx, peerID)
	return false, nil
}

// deleteReplicatorRetryAndDocs deletes the replicator retry and all retry docs.
func (db *db) deleteReplicatorRetryAndDocs(ctx context.Context, peerID string) error {
	key := core.NewReplicatorRetryIDKey(peerID)
	err := db.Peerstore().Delete(ctx, key.ToDS())
	if err != nil {
		return err
	}
	docKey := core.NewReplicatorRetryDocIDKey(peerID, "")
	q := query.Query{
		Prefix:   docKey.ToString(),
		KeysOnly: true,
	}
	results, err := db.Peerstore().Query(ctx, q)
	if err != nil {
		return err
	}
	defer closeQueryResults(results)
	for result := range results.Next() {
		err = db.Peerstore().Delete(ctx, core.NewReplicatorRetryDocIDKey(peerID, result.Key).ToDS())
		if err != nil {
			return err
		}
	}
	return nil
}

func closeQueryResults(results query.Results) {
	err := results.Close()
	if err != nil {
		log.ErrorE("Failed to close query results", err)
	}
}
