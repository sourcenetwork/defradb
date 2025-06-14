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
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	dbErrors "github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/keys"
)

const (
	// retryLoopInterval is the interval at which the retry handler checks for
	// replicators that are due for a retry.
	retryLoopInterval = 2 * time.Second
	// retryTimeout is the timeout for a single doc retry.
	retryTimeout = 10 * time.Second
)

func (db *DB) SetReplicator(ctx context.Context, rep client.ReplicatorParams) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

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

	ctx = InitContext(ctx, txn)

	storedRep := client.Replicator{}
	storedSchemas := make(map[string]struct{})
	repKey := keys.NewReplicatorKey(rep.Info.ID.String())
	hasOldRep, err := txn.Peerstore().Has(ctx, repKey.Bytes())
	if err != nil {
		return err
	}
	if hasOldRep {
		repBytes, err := txn.Peerstore().Get(ctx, repKey.Bytes())
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
	case len(rep.Collections) > 0:
		// if specific collections are chosen get them by name
		for _, name := range rep.Collections {
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

	if db.documentACP.HasValue() && !db.documentACP.Value().SupportsP2P() {
		for _, col := range collections {
			if col.Version().Policy.HasValue() {
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

	err = txn.Peerstore().Set(ctx, repKey.Bytes(), newRepBytes)
	if err != nil {
		return err
	}

	txn.OnSuccess(func() {
		// This is a node specific action which means the actor is the node itself.
		ctx := identity.WithContext(context.Background(), db.nodeIdentity)
		db.events.Publish(event.NewMessage(event.ReplicatorName, event.Replicator{
			Info:    rep.Info,
			Schemas: storedSchemas,
			Docs:    db.getDocsHeads(ctx, addedCols),
		}))
	})

	return txn.Commit(ctx)
}

func (db *DB) getDocsHeads(
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
		ctx = InitContext(ctx, txn)
		for _, col := range cols {
			keysCh, err := col.GetAllDocIDs(ctx)
			if err != nil {
				log.ErrorContextE(
					ctx,
					"Failed to get all docIDs",
					NewErrReplicatorDocID(err, dbErrors.NewKV("Collection", col.Name())),
				)
				continue
			}
			for docIDResult := range keysCh {
				if docIDResult.Err != nil {
					log.ErrorContextE(ctx, "Key channel error", docIDResult.Err)
					continue
				}
				docID := keys.DataStoreKeyFromDocID(docIDResult.ID)
				headset := coreblock.NewHeadSet(
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
						DocID:        docIDResult.ID.String(),
						Cid:          c,
						CollectionID: col.Version().CollectionID,
						Block:        blk.RawData(),
					}
				}
			}
		}
	}()

	return updateChan
}

func (db *DB) DeleteReplicator(ctx context.Context, rep client.ReplicatorParams) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	if err := rep.Info.ID.Validate(); err != nil {
		return err
	}

	// set transaction for all operations
	ctx = InitContext(ctx, txn)

	storedRep := client.Replicator{}
	storedSchemas := make(map[string]struct{})
	repKey := keys.NewReplicatorKey(rep.Info.ID.String())
	hasOldRep, err := txn.Peerstore().Has(ctx, repKey.Bytes())
	if err != nil {
		return err
	}
	if !hasOldRep {
		return ErrReplicatorNotFound
	}
	repBytes, err := txn.Peerstore().Get(ctx, repKey.Bytes())
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
	if len(rep.Collections) > 0 {
		// if specific collections are chosen get them by name
		for _, name := range rep.Collections {
			col, err := db.GetCollectionByName(ctx, name)
			if err != nil {
				return NewErrReplicatorCollections(err)
			}
			collections = append(collections, col)
		}
		// make sure the replicator exists in the datastore
		key := keys.NewReplicatorKey(rep.Info.ID.String())
		_, err = txn.Peerstore().Get(ctx, key.Bytes())
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
	key := keys.NewReplicatorKey(rep.Info.ID.String())
	if len(rep.Collections) == 0 {
		err := txn.Peerstore().Delete(ctx, key.Bytes())
		if err != nil {
			return err
		}
	} else {
		repBytes, err := json.Marshal(rep)
		if err != nil {
			return err
		}
		err = txn.Peerstore().Set(ctx, key.Bytes(), repBytes)
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

func (db *DB) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	txn, err := db.NewTxn(ctx, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	_, reps, err := datastore.DeserializePrefix[client.Replicator](
		ctx,
		keys.NewReplicatorKey("").Bytes(),
		txn.Peerstore(),
	)

	return reps, err
}

func (db *DB) loadAndPublishReplicators(ctx context.Context) error {
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

// handleReplicatorRetries manages retries for failed replication attempts.
func (db *DB) handleReplicatorRetries(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case <-time.After(retryLoopInterval):
			db.retryReplicators(ctx)
		}
	}
}

func (db *DB) handleReplicatorFailure(ctx context.Context, peerID, docID string) error {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)
	err = updateReplicatorStatus(ctx, txn, peerID, false)
	if err != nil {
		return err
	}
	err = createIfNotExistsReplicatorRetry(ctx, txn, peerID, db.retryIntervals)
	if err != nil {
		return err
	}
	docIDKey := keys.NewReplicatorRetryDocIDKey(peerID, docID)
	err = txn.Peerstore().Set(ctx, docIDKey.Bytes(), []byte{})
	if err != nil {
		return err
	}
	return txn.Commit(ctx)
}

func (db *DB) handleCompletedReplicatorRetry(ctx context.Context, peerID string, success bool) error {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)
	var done bool
	if success {
		done, err = deleteReplicatorRetryIfNoMoreDocs(ctx, txn, peerID)
		if err != nil {
			return err
		}
		if done {
			err := updateReplicatorStatus(ctx, txn, peerID, true)
			if err != nil {
				return err
			}
		} else {
			// If there are more docs to retry, set the next retry time to be immediate.
			err := setReplicatorNextRetry(ctx, txn, peerID, []time.Duration{0})
			if err != nil {
				return err
			}
		}
	} else {
		err := setReplicatorNextRetry(ctx, txn, peerID, db.retryIntervals)
		if err != nil {
			return err
		}
	}
	return txn.Commit(ctx)
}

// updateReplicatorStatus updates the status of a replicator in the peerstore.
func updateReplicatorStatus(
	ctx context.Context,
	txn datastore.Txn,
	peerID string,
	active bool,
) error {
	key := keys.NewReplicatorKey(peerID)
	repBytes, err := txn.Peerstore().Get(ctx, key.Bytes())
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
		if rep.Status == client.ReplicatorStatusInactive {
			rep.LastStatusChange = time.Time{}
		}
		rep.Status = client.ReplicatorStatusActive
	case false:
		if rep.Status == client.ReplicatorStatusActive {
			rep.LastStatusChange = time.Now()
		}
		rep.Status = client.ReplicatorStatusInactive
	}
	b, err := json.Marshal(rep)
	if err != nil {
		return err
	}
	return txn.Peerstore().Set(ctx, key.Bytes(), b)
}

type retryInfo struct {
	NextRetry  time.Time
	NumRetries int
	Retrying   bool
}

func createIfNotExistsReplicatorRetry(
	ctx context.Context,
	txn datastore.Txn,
	peerID string,
	retryIntervals []time.Duration,
) error {
	key := keys.NewReplicatorRetryIDKey(peerID)
	exists, err := txn.Peerstore().Has(ctx, key.Bytes())
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	r := retryInfo{
		NextRetry:  time.Now().Add(retryIntervals[0]),
		NumRetries: 0,
	}
	b, err := cbor.Marshal(r)
	if err != nil {
		return err
	}
	err = txn.Peerstore().Set(ctx, key.Bytes(), b)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) retryReplicators(ctx context.Context) {
	iter, err := db.Peerstore().Iterator(ctx, corekv.IterOptions{
		Prefix: []byte(keys.REPLICATOR_RETRY_ID),
	})
	if err != nil {
		log.ErrorContextE(ctx, "Failed iterate replicator retry ID keys", err)
	}
	defer closeQueryResults(iter)

	now := time.Now()
	for {
		hasNext, err := iter.Next()
		if err != nil {
			log.ErrorContextE(ctx, "Failed to get next replicator retry ID key", err)
			break
		}
		if !hasNext {
			break
		}

		key, err := keys.NewReplicatorRetryIDKeyFromString(string(iter.Key()))
		if err != nil {
			log.ErrorContextE(ctx, "Failed to parse replicator retry ID key", err)
			continue
		}

		value, err := iter.Value()
		if err != nil {
			log.ErrorContextE(ctx, "Failed to get replicator retry value", err)
			continue
		}

		rInfo := retryInfo{}
		err = cbor.Unmarshal(value, &rInfo)
		if err != nil {
			log.ErrorContextE(ctx, "Failed to unmarshal replicator retry info", err)
			// If we can't unmarshal the retry info, we delete the retry key and all related retry docs.
			err = db.deleteReplicatorRetryAndDocs(ctx, key.PeerID)
			if err != nil {
				log.ErrorContextE(ctx, "Failed to delete replicator retry and docs", err)
			}
			continue
		}
		// If the next retry time has passed and the replicator is not already retrying.
		if now.After(rInfo.NextRetry) && !rInfo.Retrying {
			// The replicator might have been deleted by the time we reach this point.
			// If it no longer exists, we delete the retry key and all retry docs.
			exists, err := db.Peerstore().Has(ctx, keys.NewReplicatorKey(key.PeerID).Bytes())
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

func (db *DB) setReplicatorAsRetrying(ctx context.Context, key keys.ReplicatorRetryIDKey, rInfo retryInfo) error {
	rInfo.Retrying = true
	rInfo.NumRetries++
	b, err := cbor.Marshal(rInfo)
	if err != nil {
		return err
	}
	return db.Peerstore().Set(ctx, key.Bytes(), b)
}

func setReplicatorNextRetry(
	ctx context.Context,
	txn datastore.Txn,
	peerID string,
	retryIntervals []time.Duration,
) error {
	key := keys.NewReplicatorRetryIDKey(peerID)
	b, err := txn.Peerstore().Get(ctx, key.Bytes())
	if err != nil {
		return err
	}
	rInfo := retryInfo{}
	err = cbor.Unmarshal(b, &rInfo)
	if err != nil {
		return err
	}
	if rInfo.NumRetries >= len(retryIntervals) {
		rInfo.NextRetry = time.Now().Add(retryIntervals[len(retryIntervals)-1])
	} else {
		rInfo.NextRetry = time.Now().Add(retryIntervals[rInfo.NumRetries])
	}
	rInfo.Retrying = false
	b, err = cbor.Marshal(rInfo)
	if err != nil {
		return err
	}
	return txn.Peerstore().Set(ctx, key.Bytes(), b)
}

// retryReplicator retries all unsycned docs for a replicator.
//
// The retry process is as follows:
// 1. Query the retry docs for the replicator.
// 2. For each doc, retry the doc.
// 3. If the doc is successfully retried, delete the retry doc.
// 4. If the doc fails to retry, stop retrying the rest of the docs and wait for the next retry.
// 5. If all docs are successfully retried, delete the replicator retry.
// 6. If there are more docs to retry, set the next retry time to be immediate.
//
// All action within this function are done outside a transaction to always get the most recent data
// and post updates as soon as possible. Because of the asyncronous nature of the retryDoc step, there
// would be a high chance of unnecessary transaction conflicts.
func (db *DB) retryReplicator(ctx context.Context, peerID string) {
	log.InfoContext(ctx, "Retrying replicator", corelog.String("PeerID", peerID))

	iter, err := db.Peerstore().Iterator(ctx, corekv.IterOptions{
		Prefix:   keys.NewReplicatorRetryDocIDKey(peerID, "").Bytes(),
		KeysOnly: true,
	})
	if err != nil {
		log.ErrorContextE(ctx, "Failed iterate replicator retry docID keys", err)
	}
	defer closeQueryResults(iter)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		hasNext, err := iter.Next()
		if err != nil {
			log.ErrorContextE(ctx, "Failed to get next replicator retry docID key", err)
			break
		}
		if !hasNext {
			break
		}

		key, err := keys.NewReplicatorRetryDocIDKeyFromString(string(iter.Key()))
		if err != nil {
			log.ErrorContextE(ctx, "Failed to parse retry doc key", err)
			continue
		}
		err = db.retryDoc(ctx, key.DocID)
		if err != nil {
			log.ErrorContextE(ctx, "Failed to retry doc", err)
			err = db.handleCompletedReplicatorRetry(ctx, peerID, false)
			if err != nil {
				log.ErrorContextE(ctx, "Failed to handle completed replicator retry", err)
			}
			// if one doc fails, stop retrying the rest and just wait for the next retry
			return
		}
		err = db.Peerstore().Delete(ctx, key.Bytes())
		if err != nil {
			log.ErrorContextE(ctx, "Failed to delete retry docID", err)
		}
	}

	err = db.handleCompletedReplicatorRetry(ctx, peerID, true)
	if err != nil {
		log.ErrorContextE(ctx, "Failed to handle completed replicator retry", err)
	}
}

func (db *DB) retryDoc(ctx context.Context, docID string) error {
	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	headsIterator, err := NewHeadBlocksIteratorFromTxn(ctx, txn, docID)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ErrContextDone
		default:
		}

		hasValue, err := headsIterator.Next()
		if err != nil {
			return err
		}
		if !hasValue {
			break
		}

		col, err := db.getCollectionByID(ctx, headsIterator.CurrentBlock().Delta.GetSchemaVersionID())
		if err != nil {
			return err
		}
		successChan := make(chan bool)
		defer close(successChan)
		updateEvent := event.Update{
			DocID:        docID,
			Cid:          headsIterator.CurrentCid(),
			CollectionID: col.Version().CollectionID,
			Block:        headsIterator.CurrentRawBlock(),
			IsRetry:      true,
			// Because the retry is done in a separate goroutine but the retry handling process should be synchronous,
			// we use a channel to block while waiting for the success status of the retry.
			Success: successChan,
		}
		db.events.Publish(event.NewMessage(event.UpdateName, updateEvent))

		select {
		case success := <-successChan:
			if !success {
				return ErrFailedToRetryDoc
			}
		case <-time.After(retryTimeout):
			return ErrTimeoutDocRetry
		}
	}
	return nil
}

// deleteReplicatorRetryIfNoMoreDocs deletes the replicator retry key if there are no more docs to retry.
// It returns true if there are no more docs to retry, false otherwise.
func deleteReplicatorRetryIfNoMoreDocs(
	ctx context.Context,
	txn datastore.Txn,
	peerID string,
) (bool, error) {
	entries, err := datastore.FetchKeysForPrefix(
		ctx,
		keys.NewReplicatorRetryDocIDKey(peerID, "").Bytes(),
		txn.Peerstore(),
	)
	if err != nil {
		return false, err
	}

	if len(entries) == 0 {
		key := keys.NewReplicatorRetryIDKey(peerID)
		return true, txn.Peerstore().Delete(ctx, key.Bytes())
	}
	return false, nil
}

// deleteReplicatorRetryAndDocs deletes the replicator retry and all retry docs.
func (db *DB) deleteReplicatorRetryAndDocs(ctx context.Context, peerID string) error {
	key := keys.NewReplicatorRetryIDKey(peerID)
	err := db.Peerstore().Delete(ctx, key.Bytes())
	if err != nil {
		return err
	}

	iter, err := db.Peerstore().Iterator(ctx, corekv.IterOptions{
		Prefix:   keys.NewReplicatorRetryDocIDKey(peerID, "").Bytes(),
		KeysOnly: true,
	})
	if err != nil {
		return err
	}

	for {
		hasNext, err := iter.Next()
		if err != nil {
			return errors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}

		err = db.Peerstore().Delete(ctx, keys.NewReplicatorRetryDocIDKey(peerID, string(iter.Key())).Bytes())
		if err != nil {
			return errors.Join(err, iter.Close())
		}
	}

	return iter.Close()
}

func closeQueryResults(iter corekv.Iterator) {
	err := iter.Close()
	if err != nil {
		log.ErrorE("Failed to close query results", err)
	}
}
