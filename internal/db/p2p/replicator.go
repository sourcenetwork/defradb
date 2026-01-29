// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package p2p

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/multiformats/go-multiaddr"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/blockstore"
	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/db/p2p/protocol"
	"github.com/sourcenetwork/defradb/internal/keys"
)

const (
	// retryLoopInterval is the interval at which the retry handler checks for
	// replicators that are due for a retry.
	retryLoopInterval = 2 * time.Second
)

func (p *P2P) SetReplicator(ctx context.Context, addresses []string, collectionNames ...string) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	clientTxn := datastore.CtxMustGetClientTxn(ctx)
	txn := datastore.MustGetFromClientTxn(clientTxn)

	// Build a map of replicator ID to list of addresses to handle multiple addresses
	replicatorMap := make(map[string][]string)
	for _, addr := range addresses {
		maddrWithID, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return err
		}
		_, p2ppart := multiaddr.SplitLast(maddrWithID)
		if p2ppart == nil || p2ppart.Protocol().Code != multiaddr.P_P2P {
			return errors.New("multiaddr does not contain peer ID")
		}
		id := p2ppart.Value()
		if id == p.host.ID() {
			return ErrSelfTargetForReplicator
		}
		if replicatorMap[id] != nil {
			replicatorMap[id] = append(replicatorMap[id], addr)
		} else {
			replicatorMap[id] = []string{addr}
		}
	}

	var fetchedCollections []client.Collection
	var err error
	switch {
	case len(collectionNames) > 0:
		// if specific collections are chosen get them by name
		for _, name := range collectionNames {
			cols, err := clientTxn.GetCollections(ctx, client.CollectionFetchOptions{Name: immutable.Some(name)})
			if err != nil {
				return NewErrReplicatorCollections(err)
			}
			if len(cols) == 0 {
				return ErrReplicatorCollections
			}

			fetchedCollections = append(fetchedCollections, cols[0])
		}

	default:
		fetchedCollections, err = clientTxn.GetCollections(ctx, client.CollectionFetchOptions{})
		if err != nil {
			return NewErrReplicatorCollections(err)
		}
	}

	// Update the list of collections for each replicator prior to persisting.
	storedRepCollectionIDs := make(map[string]map[string]struct{}) // replicatorID => collectionID
	addedCols := make(map[string][]client.Collection)              // peerID => list of collections added

	for id, addresses := range replicatorMap {
		if storedRepCollectionIDs[id] == nil {
			storedRepCollectionIDs[id] = make(map[string]struct{})
		}
		repKey := keys.NewReplicatorKey(id)
		hasOldRep, err := txn.Peerstore().Has(ctx, repKey.Bytes())
		if err != nil {
			return err
		}

		storedRep := client.Replicator{}
		if hasOldRep {
			repBytes, err := txn.Peerstore().Get(ctx, repKey.Bytes())
			if err != nil {
				return err
			}
			err = json.Unmarshal(repBytes, &storedRep)
			if err != nil {
				return err
			}
			for _, colID := range storedRep.CollectionIDs {
				storedRepCollectionIDs[id][colID] = struct{}{}
			}
		} else {
			storedRep.ID = id
			storedRep.LastStatusChange = time.Now()
		}
		// Update the list of addresses for this replicator whether it is new or existing.
		storedRep.Addresses = addresses

		for _, col := range fetchedCollections {
			if _, ok := storedRepCollectionIDs[id][col.CollectionID()]; !ok {
				storedRepCollectionIDs[id][col.CollectionID()] = struct{}{}
				addedCols[id] = append(addedCols[id], col)
				storedRep.CollectionIDs = append(storedRep.CollectionIDs, col.CollectionID())
			}
		}

		newRepBytes, err := json.Marshal(storedRep)
		if err != nil {
			return err
		}

		err = txn.Peerstore().Set(ctx, repKey.Bytes(), newRepBytes)
		if err != nil {
			return err
		}
	}

	txn.OnSuccessAsync(func() {
		for id, addresses := range replicatorMap {
			p.updateReplicators(ctx, id, addresses, storedRepCollectionIDs[id])
			for _, col := range addedCols[id] {
				err := p.pushHeadsForAllDocs(context.Background(), col, id)
				if err != nil {
					log.ErrorE(
						"Failed push heads for all docs",
						err,
						corelog.Any("Collection", col.Name()),
					)
				}
			}
		}
		p.db.Events().Publish(event.NewMessage(event.ReplicatorCompletedName, nil))
	})

	return nil
}

// pushHeadsForAllDocs gets all the docID for the given collection and sends them to get
// pushed to the given peer.
func (p *P2P) pushHeadsForAllDocs(ctx context.Context, col client.Collection, peerID string) error {
	// this method cannot be run inside of a transaction
	// so we have to create an unsafe iterator manually
	// instead of calling db.GetAllDocIDs
	type unsafeDatastore interface {
		Unsafe() corekv.ReaderWriter
	}
	shortID, err := id.GetUncachedShortCollectionID(ctx, col.Version().CollectionID, p.db.Multistore().Systemstore())
	if err != nil {
		return err
	}
	prefix := keys.PrimaryDataStoreKey{CollectionShortID: shortID}
	ds := p.db.Multistore().Datastore().(unsafeDatastore).Unsafe() //nolint:forcetypeassert
	iter, err := ds.Iterator(ctx, corekv.IterOptions{Prefix: prefix.Bytes(), KeysOnly: true})
	if err != nil {
		return err
	}
	defer func() {
		if iterErr := iter.Close(); iterErr != nil {
			log.ErrorE("Failed to close docID iter", iterErr)
		}
	}()

	for {
		hasNext, err := iter.Next()
		if err != nil {
			return err
		}
		if !hasNext {
			return nil
		}
		splitString := strings.Split(string(iter.Key()), "/")
		docID := splitString[len(splitString)-1]
		err = p.pushHeadsForDoc(ctx, docID, col.CollectionID(), peerID)
		if err != nil {
			return err
		}
	}
}

// pushHeadsForDoc gets the all the head blocks for a given docID and pushes them
// to the given peer.
func (p *P2P) pushHeadsForDoc(ctx context.Context, docID, collectionID string, peerID string) error {
	heads, err := p.getHeads(ctx, docID)
	if err != nil {
		return err
	}
	for _, head := range heads {
		rawblock, err := head.block.Marshal()
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(ctx, networkRequestTimeout)
		defer cancel()
		pushLogReq := protocol.PushLogRequest{
			DocID:        docID,
			CID:          head.cid.Bytes(),
			CollectionID: collectionID,
			Creator:      p.host.ID(),
			Block:        rawblock,
		}

		if _, err := p.replicatorProtocol.SendRequest(ctx, pushLogReq, peerID); err != nil {
			log.ErrorE(
				"Failed to push doc heads. Handling replicator failure",
				err,
				corelog.Any("DocID", docID),
			)
			err := p.handleReplicatorFailure(ctx, peerID, docID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *P2P) DeleteReplicator(ctx context.Context, id string, collectionNames ...string) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	clientTxn := datastore.CtxMustGetClientTxn(ctx)
	txn := datastore.MustGetFromClientTxn(clientTxn)

	storedRep := client.Replicator{}
	storedCollectionIDs := make(map[string]struct{})
	repKey := keys.NewReplicatorKey(id)
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
	for _, id := range storedRep.CollectionIDs {
		storedCollectionIDs[id] = struct{}{}
	}
	if len(collectionNames) > 0 {
		// if specific collections are chosen get them by name
		for _, name := range collectionNames {
			cols, err := clientTxn.GetCollections(ctx, client.CollectionFetchOptions{Name: immutable.Some(name)})
			if err != nil {
				return NewErrReplicatorCollections(err)
			}
			if len(cols) == 0 {
				return ErrReplicatorCollections
			}
			delete(storedCollectionIDs, cols[0].CollectionID())
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
	key := keys.NewReplicatorKey(id)
	if len(storedRep.CollectionIDs) == 0 {
		err := txn.Peerstore().Delete(ctx, key.Bytes())
		if err != nil {
			return err
		}
	} else {
		repBytes, err := json.Marshal(storedRep)
		if err != nil {
			return err
		}
		err = txn.Peerstore().Set(ctx, key.Bytes(), repBytes)
		if err != nil {
			return err
		}
	}

	txn.OnSuccess(func() {
		p.updateReplicators(ctx, storedRep.ID, storedRep.Addresses, storedCollectionIDs)
		p.db.Events().Publish(event.NewMessage(event.ReplicatorCompletedName, nil))
	})

	return nil
}

func (p *P2P) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	_, reps, err := datastore.DeserializePrefix[client.Replicator](
		ctx,
		keys.NewReplicatorKey("").Bytes(),
		p.db.Multistore().Peerstore(),
	)

	return reps, err
}

func (p *P2P) pushLogToReplicators(lg event.Update) {
	p.repMu.Lock()
	reps, exists := p.replicators[lg.CollectionID]
	p.repMu.Unlock()

	for _, handler := range p.pushHandlers {
		if err := handler.HandlePushToReplicators(context.Background(), lg); err != nil {
			log.ErrorE("Push handler failed", err,
				corelog.String("DocID", lg.DocID),
				corelog.String("CollectionID", lg.CollectionID))
		}
	}

	if exists {
		for peerID := range reps {
			go func() {
				ctx, cancel := context.WithTimeout(p.ctx, networkRequestTimeout)
				defer cancel()
				pushLogReq := protocol.PushLogRequest{
					DocID:        lg.DocID,
					CID:          lg.Cid.Bytes(),
					CollectionID: lg.CollectionID,
					Creator:      p.host.ID(),
					Block:        lg.Block,
				}
				if _, err := p.replicatorProtocol.SendRequest(ctx, pushLogReq, peerID); err != nil {
					log.ErrorE(
						"Failed pushing log",
						err,
						corelog.String("DocID", lg.DocID),
						corelog.Any("CID", lg.Cid),
						corelog.Any("PeerID", peerID))
					if !lg.IsRetry {
						err = p.handleReplicatorFailure(ctx, peerID, lg.DocID)
						if err != nil {
							log.ErrorE("Failed to handle replicator failure.", err)
						}
					}
				}
			}()
		}
	}
}

func (p *P2P) loadAndPublishReplicators(ctx context.Context) error {
	replicators, err := p.GetAllReplicators(ctx)
	if err != nil {
		return err
	}

	for _, rep := range replicators {
		storedCollectionIDs := make(map[string]struct{})
		for _, id := range rep.CollectionIDs {
			storedCollectionIDs[id] = struct{}{}
		}
		p.updateReplicators(ctx, rep.ID, rep.Addresses, storedCollectionIDs)
	}
	return nil
}

// handleReplicatorRetries manages retries for failed replication attempts.
func (p *P2P) handleReplicatorRetries(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case <-time.After(retryLoopInterval):
			p.retryReplicators(ctx)
		}
	}
}

func (p *P2P) handleReplicatorFailure(ctx context.Context, peerID, docID string) error {
	// This method can be called concurrently for the same peerID which can cause some
	// transaction conflicts. Since this is not a performance critical operation, it's
	// safe to use a mutex to prevent unnecessary conflicts.
	p.handleRetryMutex.Lock()
	defer p.handleRetryMutex.Unlock()

	err := updateReplicatorStatus(ctx, peerID, false, p.db.Multistore().Peerstore())
	if err != nil {
		return err
	}
	err = createIfNotExistsReplicatorRetry(ctx, peerID, p.retryIntervals, p.db.Multistore().Peerstore())
	if err != nil {
		return err
	}
	docIDKey := keys.NewReplicatorRetryDocIDKey(peerID, docID)
	return p.db.Multistore().Peerstore().Set(ctx, docIDKey.Bytes(), []byte{})
}

func (p *P2P) handleCompletedReplicatorRetry(ctx context.Context, peerID string, success bool) error {
	if success {
		done, err := deleteReplicatorRetryIfNoMoreDocs(ctx, peerID, p.db.Multistore().Peerstore())
		if err != nil {
			return err
		}
		if done {
			err := updateReplicatorStatus(ctx, peerID, true, p.db.Multistore().Peerstore())
			if err != nil {
				return err
			}
		} else {
			// If there are more docs to retry, set the next retry time to be immediate.
			err := setReplicatorNextRetry(ctx, peerID, []time.Duration{0}, p.db.Multistore().Peerstore())
			if err != nil {
				return err
			}
		}
	} else {
		err := setReplicatorNextRetry(ctx, peerID, p.retryIntervals, p.db.Multistore().Peerstore())
		if err != nil {
			return err
		}
	}
	return nil
}

// updateReplicatorStatus updates the status of a replicator in the peerstore.
func updateReplicatorStatus(
	ctx context.Context,
	peerID string,
	active bool,
	peerstore corekv.ReaderWriter,
) error {
	key := keys.NewReplicatorKey(peerID)
	repBytes, err := peerstore.Get(ctx, key.Bytes())
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
	return peerstore.Set(ctx, key.Bytes(), b)
}

type retryInfo struct {
	NextRetry  time.Time
	NumRetries int
	Retrying   bool
}

func createIfNotExistsReplicatorRetry(
	ctx context.Context,
	peerID string,
	retryIntervals []time.Duration,
	peerstore corekv.ReaderWriter,
) error {
	key := keys.NewReplicatorRetryIDKey(peerID)
	exists, err := peerstore.Has(ctx, key.Bytes())
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
	return peerstore.Set(ctx, key.Bytes(), b)
}

func (p *P2P) retryReplicators(ctx context.Context) {
	iter, err := p.db.Multistore().Peerstore().Iterator(ctx, corekv.IterOptions{
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
			exists, err := p.db.Multistore().Peerstore().Has(ctx, keys.NewReplicatorKey(key.PeerID).Bytes())
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

func (p *P2P) setReplicatorAsRetrying(ctx context.Context, key keys.ReplicatorRetryIDKey, rInfo retryInfo) error {
	rInfo.Retrying = true
	rInfo.NumRetries++
	b, err := cbor.Marshal(rInfo)
	if err != nil {
		return err
	}

	return p.db.Multistore().Peerstore().Set(ctx, key.Bytes(), b)
}

func setReplicatorNextRetry(
	ctx context.Context,
	peerID string,
	retryIntervals []time.Duration,
	peerstore corekv.ReaderWriter,
) error {
	key := keys.NewReplicatorRetryIDKey(peerID)
	b, err := peerstore.Get(ctx, key.Bytes())
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
	return peerstore.Set(ctx, key.Bytes(), b)
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
func (p *P2P) retryReplicator(ctx context.Context, peerID string) {
	log.InfoContext(ctx, "Retrying replicator", corelog.String("PeerID", peerID))

	iter, err := p.db.Multistore().Peerstore().Iterator(ctx, corekv.IterOptions{
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
		err = p.db.Multistore().Peerstore().Delete(ctx, key.Bytes())
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

func (p *P2P) getHeads(ctx context.Context, docID string) ([]head, error) {
	headstore := p.db.Multistore().Headstore()
	blockstore := blockstore.NewIPLDStore(p.db.Multistore().Blockstore())

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

func (p *P2P) retryDoc(ctx context.Context, peerID string, docID string) error {
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

		ctx, cancel := context.WithTimeout(ctx, networkRequestTimeout)
		defer cancel()
		pushLogReq := protocol.PushLogRequest{
			DocID:        docID,
			CID:          head.cid.Bytes(),
			CollectionID: head.block.Delta.GetCollectionVersionID(),
			Creator:      p.host.ID(),
			Block:        rawblock,
		}
		if _, err := p.replicatorProtocol.SendRequest(ctx, pushLogReq, peerID); err != nil {
			return err
		}
	}
	return nil
}

// deleteReplicatorRetryIfNoMoreDocs deletes the replicator retry key if there are no more docs to retry.
// It returns true if there are no more docs to retry, false otherwise.
func deleteReplicatorRetryIfNoMoreDocs(
	ctx context.Context,
	peerID string,
	peerstore corekv.ReaderWriter,
) (bool, error) {
	entries, err := datastore.FetchKeysForPrefix(
		ctx,
		keys.NewReplicatorRetryDocIDKey(peerID, "").Bytes(),
		peerstore,
	)
	if err != nil {
		return false, err
	}

	if len(entries) == 0 {
		key := keys.NewReplicatorRetryIDKey(peerID)
		return true, peerstore.Delete(ctx, key.Bytes())
	}
	return false, nil
}

// deleteReplicatorRetryAndDocs deletes the replicator retry and all retry docs.
func (p *P2P) deleteReplicatorRetryAndDocs(ctx context.Context, peerID string) error {
	key := keys.NewReplicatorRetryIDKey(peerID)
	err := p.db.Multistore().Peerstore().Delete(ctx, key.Bytes())
	if err != nil {
		return err
	}

	iter, err := p.db.Multistore().Peerstore().Iterator(ctx, corekv.IterOptions{
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

		err = p.db.Multistore().Peerstore().Delete(ctx, keys.NewReplicatorRetryDocIDKey(peerID, string(iter.Key())).Bytes())
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

// GetReplicatorsIDs returns a slice of replicator IDs associated with the specified collection.
func (p *P2P) GetReplicatorsIDs(collectionID string) []string {
	p.repMu.Lock()
	defer p.repMu.Unlock()
	colReplicators := p.replicators[collectionID]
	ids := make([]string, 0, len(colReplicators))
	for id := range colReplicators {
		ids = append(ids, id)
	}
	return ids
}
