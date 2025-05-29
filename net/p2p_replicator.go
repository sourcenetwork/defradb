// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package net

import (
	"context"
	"encoding/json"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/keys"
)

const (
	// retryLoopInterval is the interval at which the retry handler checks for
	// replicators that are due for a retry.
	retryLoopInterval = 2 * time.Second
)

func (p *Peer) SetReplicator(ctx context.Context, repInfo peer.AddrInfo, collections ...string) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	ctx, txn := datastore.EnsureContextTxn(ctx, p.db.Rootstore(), false)
	defer txn.Discard(ctx)

	if err := repInfo.ID.Validate(); err != nil {
		return err
	}

	peerInfo := p.PeerInfo()
	if repInfo.ID == peerInfo.ID {
		return ErrSelfTargetForReplicator
	}

	repKey := keys.NewReplicatorKey(repInfo.ID.String())
	hasOldRep, err := datastore.PeerstoreFrom(txn.Store()).Has(ctx, repKey.Bytes())
	if err != nil {
		return err
	}

	storedRep := client.Replicator{}
	storedCollectionIDs := make(map[string]struct{})
	if hasOldRep {
		repBytes, err := datastore.PeerstoreFrom(txn.Store()).Get(ctx, repKey.Bytes())
		if err != nil {
			return err
		}
		err = json.Unmarshal(repBytes, &storedRep)
		if err != nil {
			return err
		}
		for _, id := range storedRep.CollectionIDs {
			storedCollectionIDs[id] = struct{}{}
		}
	} else {
		storedRep.Info = repInfo
		storedRep.LastStatusChange = time.Now()
	}

	var fetchedCollections []client.Collection
	switch {
	case len(collections) > 0:
		// if specific collections are chosen get them by name
		for _, name := range collections {
			cols, err := p.db.GetCollections(ctx, client.CollectionFetchOptions{Name: immutable.Some(name)})
			if err != nil {
				return NewErrReplicatorCollections(err)
			}
			if len(cols) == 0 {
				return ErrReplicatorCollections
			}

			fetchedCollections = append(fetchedCollections, cols[0])
		}

	default:
		fetchedCollections, err = p.db.GetCollections(ctx, client.CollectionFetchOptions{})
		if err != nil {
			return NewErrReplicatorCollections(err)
		}
	}

	addedCols := []client.Collection{}
	for _, col := range fetchedCollections {
		if _, ok := storedCollectionIDs[col.SchemaRoot()]; !ok {
			storedCollectionIDs[col.SchemaRoot()] = struct{}{}
			addedCols = append(addedCols, col)
			storedRep.CollectionIDs = append(storedRep.CollectionIDs, col.SchemaRoot())
		}
	}

	// persist replicator to the datastore
	newRepBytes, err := json.Marshal(storedRep)
	if err != nil {
		return err
	}

	err = datastore.PeerstoreFrom(txn.Store()).Set(ctx, repKey.Bytes(), newRepBytes)
	if err != nil {
		return err
	}

	txn.OnSuccessAsync(func() {
		p.server.updateReplicators(repInfo, storedCollectionIDs)
		for _, col := range addedCols {
			err := p.pushHeadsForAllDocs(context.Background(), col, repInfo.ID)
			if err != nil {
				log.ErrorE(
					"Failed push heads for all docs",
					err,
					corelog.Any("Collection", col.Name()),
				)
			}
		}
		p.bus.Publish(event.NewMessage(event.ReplicatorCompletedName, nil))
	})

	return txn.Commit(ctx)
}

// pushHeadsForAllDocs gets all the docID for the given collection and sends them to get
// pushed to the given peer.
func (p *Peer) pushHeadsForAllDocs(ctx context.Context, col client.Collection, peerID peer.ID) error {
	ctx, _ = datastore.EnsureContextTxn(ctx, p.db.Rootstore(), true)
	docIDChan, err := col.GetAllDocIDs(ctx)
	if err != nil {
		return err
	}
	for docIDResult := range docIDChan {
		if docIDResult.Err != nil {
			return docIDResult.Err
		}
		docID := docIDResult.ID.String()
		err := p.pushHeadsForDoc(ctx, docID, col.Version().CollectionID, peerID)
		if err != nil {
			return err
		}
	}
	return nil
}

// pushHeadsForDoc gets the all the head blocks for a given docID and pushes them
// to the given peer.
func (p *Peer) pushHeadsForDoc(ctx context.Context, docID, collectionID string, peerID peer.ID) error {
	heads, err := p.getHeads(ctx, docID)
	if err != nil {
		return err
	}
	for _, head := range heads {
		rawblock, err := head.block.Marshal()
		if err != nil {
			return err
		}
		update := event.Update{
			DocID:        docID,
			Cid:          head.cid,
			CollectionID: collectionID,
			Block:        rawblock,
		}
		if err := p.server.pushLog(update, peerID); err != nil {
			log.ErrorE(
				"Failed to push doc heads. Handling replicator failure",
				err,
				corelog.Any("DocID", docID),
			)
			err := p.handleReplicatorFailure(ctx, peerID.String(), docID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Peer) DeleteReplicator(ctx context.Context, repInfo peer.AddrInfo, collections ...string) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	ctx, txn := datastore.EnsureContextTxn(ctx, p.db.Rootstore(), false)
	defer txn.Discard(ctx)

	if err := repInfo.ID.Validate(); err != nil {
		return err
	}

	storedRep := client.Replicator{}
	storedCollectionIDs := make(map[string]struct{})
	repKey := keys.NewReplicatorKey(repInfo.ID.String())
	hasOldRep, err := datastore.PeerstoreFrom(txn.Store()).Has(ctx, repKey.Bytes())
	if err != nil {
		return err
	}
	if !hasOldRep {
		return ErrReplicatorNotFound
	}
	repBytes, err := datastore.PeerstoreFrom(txn.Store()).Get(ctx, repKey.Bytes())
	if err != nil {
		return err
	}
	err = json.Unmarshal(repBytes, &storedRep)
	if err != nil {
		return err
	}
	for _, id := range storedRep.CollectionIDs {
		storedCollectionIDs[id] = struct{}{}
	}
	if len(collections) > 0 {
		// if specific collections are chosen get them by name
		for _, name := range collections {
			cols, err := p.db.GetCollections(ctx, client.CollectionFetchOptions{Name: immutable.Some(name)})
			if err != nil {
				return NewErrReplicatorCollections(err)
			}
			if len(cols) == 0 {
				return ErrReplicatorCollections
			}
			delete(storedCollectionIDs, cols[0].SchemaRoot())
		}
	} else {
		storedCollectionIDs = make(map[string]struct{})
	}

	// Update the list of schemas for this replicator prior to persisting.
	storedRep.CollectionIDs = []string{}
	for id := range storedCollectionIDs {
		storedRep.CollectionIDs = append(storedRep.CollectionIDs, id)
	}

	// Persist the replicator to the store, deleting it if no collection remain
	key := keys.NewReplicatorKey(repInfo.ID.String())
	if len(storedRep.CollectionIDs) == 0 {
		err := datastore.PeerstoreFrom(txn.Store()).Delete(ctx, key.Bytes())
		if err != nil {
			return err
		}
	} else {
		repBytes, err := json.Marshal(storedRep)
		if err != nil {
			return err
		}
		err = datastore.PeerstoreFrom(txn.Store()).Set(ctx, key.Bytes(), repBytes)
		if err != nil {
			return err
		}
	}

	txn.OnSuccess(func() {
		p.server.updateReplicators(repInfo, storedCollectionIDs)
		p.bus.Publish(event.NewMessage(event.ReplicatorCompletedName, nil))
	})

	return txn.Commit(ctx)
}

func (p *Peer) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	ctx, txn := datastore.EnsureContextTxn(ctx, p.db.Rootstore(), false)
	defer txn.Discard(ctx)

	_, reps, err := datastore.DeserializePrefix[client.Replicator](
		ctx,
		keys.NewReplicatorKey("").Bytes(),
		datastore.PeerstoreFrom(txn.Store()),
	)

	return reps, err
}

func (p *Peer) loadAndPublishReplicators(ctx context.Context) error {
	replicators, err := p.GetAllReplicators(ctx)
	if err != nil {
		return err
	}

	for _, rep := range replicators {
		storedCollectionIDs := make(map[string]struct{})
		for _, id := range rep.CollectionIDs {
			storedCollectionIDs[id] = struct{}{}
		}
		p.server.updateReplicators(rep.Info, storedCollectionIDs)
	}
	return nil
}

// handleReplicatorRetries manages retries for failed replication attempts.
func (p *Peer) handleReplicatorRetries(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case <-time.After(retryLoopInterval):
			p.retryReplicators(ctx)
		}
	}
}

func (p *Peer) handleReplicatorFailure(ctx context.Context, peerID, docID string) error {
	ctx, txn := datastore.EnsureContextTxn(ctx, p.db.Rootstore(), false)
	defer txn.Discard(ctx)

	err := updateReplicatorStatus(ctx, txn, peerID, false)
	if err != nil {
		return err
	}
	err = createIfNotExistsReplicatorRetry(ctx, txn, peerID, p.retryIntervals)
	if err != nil {
		return err
	}
	docIDKey := keys.NewReplicatorRetryDocIDKey(peerID, docID)
	err = datastore.PeerstoreFrom(txn.Store()).Set(ctx, docIDKey.Bytes(), []byte{})
	if err != nil {
		return err
	}
	return txn.Commit(ctx)
}

func (p *Peer) handleCompletedReplicatorRetry(ctx context.Context, peerID string, success bool) error {
	ctx, txn := datastore.EnsureContextTxn(ctx, p.db.Rootstore(), false)
	defer txn.Discard(ctx)

	if success {
		done, err := deleteReplicatorRetryIfNoMoreDocs(ctx, txn, peerID)
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
		err := setReplicatorNextRetry(ctx, txn, peerID, p.retryIntervals)
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
	repBytes, err := datastore.PeerstoreFrom(txn.Store()).Get(ctx, key.Bytes())
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
	return datastore.PeerstoreFrom(txn.Store()).Set(ctx, key.Bytes(), b)
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
	exists, err := datastore.PeerstoreFrom(txn.Store()).Has(ctx, key.Bytes())
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
	err = datastore.PeerstoreFrom(txn.Store()).Set(ctx, key.Bytes(), b)
	if err != nil {
		return err
	}
	return nil
}

func (p *Peer) retryReplicators(ctx context.Context) {
	iter, err := datastore.PeerstoreFrom(p.db.Rootstore()).Iterator(ctx, corekv.IterOptions{
		Prefix: []byte(keys.REPLICATOR_RETRY_ID),
	})
	if err != nil {
		if errors.Is(err, corekv.ErrDBClosed) {
			return
		}
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
			err = p.deleteReplicatorRetryAndDocs(ctx, key.PeerID)
			if err != nil {
				log.ErrorContextE(ctx, "Failed to delete replicator retry and docs", err)
			}
			continue
		}
		// If the next retry time has passed and the replicator is not already retrying.
		if now.After(rInfo.NextRetry) && !rInfo.Retrying {
			// The replicator might have been deleted by the time we reach this point.
			// If it no longer exists, we delete the retry key and all retry docs.
			exists, err := datastore.PeerstoreFrom(p.db.Rootstore()).Has(ctx, keys.NewReplicatorKey(key.PeerID).Bytes())
			if err != nil {
				log.ErrorContextE(ctx, "Failed to check if replicator exists", err)
				continue
			}
			if !exists {
				err = p.deleteReplicatorRetryAndDocs(ctx, key.PeerID)
				if err != nil {
					log.ErrorContextE(ctx, "Failed to delete replicator retry and docs", err)
				}
				continue
			}

			err = p.setReplicatorAsRetrying(ctx, key, rInfo)
			if err != nil {
				log.ErrorContextE(ctx, "Failed to set replicator as retrying", err)
				continue
			}
			go p.retryReplicator(ctx, key.PeerID)
		}
	}
}

func (p *Peer) setReplicatorAsRetrying(ctx context.Context, key keys.ReplicatorRetryIDKey, rInfo retryInfo) error {
	rInfo.Retrying = true
	rInfo.NumRetries++
	b, err := cbor.Marshal(rInfo)
	if err != nil {
		return err
	}
	return datastore.PeerstoreFrom(p.db.Rootstore()).Set(ctx, key.Bytes(), b)
}

func setReplicatorNextRetry(
	ctx context.Context,
	txn datastore.Txn,
	peerID string,
	retryIntervals []time.Duration,
) error {
	key := keys.NewReplicatorRetryIDKey(peerID)
	b, err := datastore.PeerstoreFrom(txn.Store()).Get(ctx, key.Bytes())
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
	return datastore.PeerstoreFrom(txn.Store()).Set(ctx, key.Bytes(), b)
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
func (p *Peer) retryReplicator(ctx context.Context, peerID string) {
	log.InfoContext(ctx, "Retrying replicator", corelog.String("PeerID", peerID))

	iter, err := datastore.PeerstoreFrom(p.db.Rootstore()).Iterator(ctx, corekv.IterOptions{
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
		err = p.retryDoc(ctx, peerID, key.DocID)
		if err != nil {
			log.ErrorContextE(ctx, "Failed to retry doc", err)
			err = p.handleCompletedReplicatorRetry(ctx, peerID, false)
			if err != nil {
				log.ErrorContextE(ctx, "Failed to handle completed replicator retry", err)
			}
			// if one doc fails, stop retrying the rest and just wait for the next retry
			return
		}
		err = datastore.PeerstoreFrom(p.db.Rootstore()).Delete(ctx, key.Bytes())
		if err != nil {
			log.ErrorContextE(ctx, "Failed to delete retry docID", err)
		}
	}

	err = p.handleCompletedReplicatorRetry(ctx, peerID, true)
	if err != nil {
		log.ErrorContextE(ctx, "Failed to handle completed replicator retry", err)
	}
}

type head struct {
	cid   cid.Cid
	block *coreblock.Block
}

func (p *Peer) getHeads(ctx context.Context, docID string) ([]head, error) {
	ctx, txn := datastore.EnsureContextTxn(ctx, p.db.Rootstore(), false)
	defer txn.Discard(ctx)

	headstore := datastore.HeadstoreFrom(txn.Store())
	blockstore := datastore.BlockstoreFrom(txn.Store()).AsIPLDStorage()

	prefix := keys.HeadstoreDocKey{
		DocID:   docID,
		FieldID: core.COMPOSITE_NAMESPACE,
	}

	iter, err := headstore.Iterator(ctx, corekv.IterOptions{
		Prefix: prefix.Bytes(),
	})
	if err != nil {
		return nil, err
	}
	heads := []head{}
	for {
		select {
		case <-ctx.Done():
			return nil, ErrContextDone
		default:
		}
		hasNext, err := iter.Next()
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}
		headstorekey, err := keys.NewHeadstoreDocKey(string(iter.Key()))
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}
		linkSys := cidlink.DefaultLinkSystem()
		linkSys.SetWriteStorage(blockstore)
		linkSys.SetReadStorage(blockstore)
		linkSys.TrustedStorage = true
		nd, err := linkSys.Load(
			linking.LinkContext{Ctx: ctx},
			cidlink.Link{Cid: headstorekey.Cid},
			coreblock.BlockSchemaPrototype,
		)
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}
		block, err := coreblock.GetFromNode(nd)
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}
		heads = append(heads, head{cid: headstorekey.Cid, block: block})
	}
	return heads, iter.Close()
}

func (p *Peer) retryDoc(ctx context.Context, peerIDString string, docID string) error {
	ctx, txn := datastore.EnsureContextTxn(ctx, p.db.Rootstore(), false)
	defer txn.Discard(ctx)

	heads, err := p.getHeads(ctx, docID)
	if err != nil {
		return err
	}

	for _, head := range heads {
		select {
		case <-ctx.Done():
			return ErrContextDone
		default:
		}

		rawblock, err := head.block.Marshal()
		if err != nil {
			return err
		}
		updateEvent := event.Update{
			DocID:        docID,
			Cid:          head.cid,
			CollectionID: head.block.Delta.GetSchemaVersionID(),
			Block:        rawblock,
			IsRetry:      true,
		}
		peerID, err := peer.IDFromBytes([]byte(peerIDString))
		if err != nil {
			return err
		}
		if err := p.server.pushLog(updateEvent, peerID); err != nil {
			return err
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
		datastore.PeerstoreFrom(txn.Store()),
	)
	if err != nil {
		return false, err
	}

	if len(entries) == 0 {
		key := keys.NewReplicatorRetryIDKey(peerID)
		return true, datastore.PeerstoreFrom(txn.Store()).Delete(ctx, key.Bytes())
	}
	return false, nil
}

// deleteReplicatorRetryAndDocs deletes the replicator retry and all retry docs.
func (p *Peer) deleteReplicatorRetryAndDocs(ctx context.Context, peerID string) error {
	ctx, txn := datastore.EnsureContextTxn(ctx, p.db.Rootstore(), false)
	defer txn.Discard(ctx)

	peerstore := datastore.PeerstoreFrom(txn.Store())

	key := keys.NewReplicatorRetryIDKey(peerID)
	err := peerstore.Delete(ctx, key.Bytes())
	if err != nil {
		return err
	}

	iter, err := peerstore.Iterator(ctx, corekv.IterOptions{
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

		err = peerstore.Delete(ctx, keys.NewReplicatorRetryDocIDKey(peerID, string(iter.Key())).Bytes())
		if err != nil {
			return errors.Join(err, iter.Close())
		}
	}

	return iter.Close()
}

func closeQueryResults(iter corekv.Iterator) {
	if iter == nil {
		return
	}
	err := iter.Close()
	if err != nil {
		log.ErrorE("Failed to close query results", err)
	}
}
