// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package se

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fxamacker/cbor/v2"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
	secore "github.com/sourcenetwork/defradb/internal/se/core"
)

const (
	// retryLoopInterval is the interval at which the retry handler checks for
	// SE artifacts that are due for a retry. Same as document replicator.
	retryLoopInterval = 2 * time.Second
	// retryTimeout is the timeout for a single SE artifact retry.
	retryTimeout = 10 * time.Second
)

var log = corelog.NewLogger("defra.se.replication")

// ReplicationCoordinator manages SE artifact replication and storage
type ReplicationCoordinator struct {
	db             DB
	eventBus       event.Bus
	failureSub     event.Subscription
	updateSub      event.Subscription
	retryIntervals []time.Duration
	encKey         []byte // Encryption key for SE artifacts
}

// DB interface required by ReplicationCoordinator
type DB interface {
	Rootstore() corekv.TxnStore
	Events() event.Bus
	MaxTxnRetries() int
	GetCollections(context.Context, client.CollectionFetchOptions) ([]client.Collection, error)
}

// SERetryInfo stores retry information for failed SE replications
type SERetryInfo struct {
	DocID        string
	CollectionID string
	FieldNames   []string
	NextRetry    time.Time
	NumRetries   int
	Retrying     bool
}

// NewReplicationCoordinator creates a new coordinator
func NewReplicationCoordinator(db DB, encKey []byte) (*ReplicationCoordinator, error) {
	rc := &ReplicationCoordinator{
		db:             db,
		eventBus:       db.Events(),
		retryIntervals: defaultRetryIntervals(db.MaxTxnRetries()),
		encKey:         encKey,
	}

	failureSub, err := db.Events().Subscribe(ReplicationFailureEventName)
	if err != nil {
		return nil, err
	}
	rc.failureSub = failureSub

	updateSub, err := db.Events().Subscribe(event.UpdateName)
	if err != nil {
		return nil, err
	}
	rc.updateSub = updateSub

	go rc.processFailureEvents()
	go rc.processUpdateEvents()

	go rc.retrySEReplicators(context.Background())

	return rc, nil
}

func (rc *ReplicationCoordinator) Close() {
	rc.eventBus.Unsubscribe(rc.failureSub)
	rc.eventBus.Unsubscribe(rc.updateSub)
}

// defaultRetryIntervals generates retry intervals based on max retries
func defaultRetryIntervals(maxRetries int) []time.Duration {
	intervals := make([]time.Duration, maxRetries)
	for i := range maxRetries {
		// Exponential backoff: 2s, 4s, 8s, 16s...
		intervals[i] = time.Second * time.Duration(2<<i)
	}
	return intervals
}

// processFailureEvents handles replication failures
func (rc *ReplicationCoordinator) processFailureEvents() {
	for {
		msg, isOpen := <-rc.failureSub.Message()
		if !isOpen {
			return
		}

		if evt, ok := msg.Data.(ReplicationFailureEvent); ok {
			if err := rc.handleReplicationFailure(context.Background(), evt); err != nil {
				log.ErrorE("Failed to handle SE replication failure", err)
			}
		}
	}
}

// processUpdateEvents handles updates to SE artifacts
func (rc *ReplicationCoordinator) processUpdateEvents() {
	for {
		msg, isOpen := <-rc.updateSub.Message()
		if !isOpen {
			return
		}

		if evt, ok := msg.Data.(event.Update); ok {
			if err := rc.handleUpdateEvent(context.Background(), evt); err != nil {
				log.ErrorE("Failed to handle SE update event", err)
			}
		}
	}
}

// handleReplicationFailure stores failed SE replication attempt for retry
func (rc *ReplicationCoordinator) handleReplicationFailure(ctx context.Context, evt ReplicationFailureEvent) error {
	retryKey := keys.NewPeerstoreSERetry(evt.PeerID.String(), evt.CollectionID, evt.DocID)

	// TODO: think if such scenario is possible: "age" field is updated but failed to replicate and while being retried
	// another "name" field is updated. In this case we should not overwrite the retry info.
	retryInfo := SERetryInfo{
		DocID:        evt.DocID,
		CollectionID: evt.CollectionID,
		FieldNames:   evt.FieldNames,
		NextRetry:    time.Now().Add(rc.retryIntervals[0]),
		NumRetries:   0,
	}

	b, err := cbor.Marshal(retryInfo)
	if err != nil {
		return err
	}

	ps := datastore.PeerstoreFrom(rc.db.Rootstore())
	return ps.Set(ctx, retryKey.Bytes(), b)
}

// handleUpdateEvent processes SE update events and stores artifacts
func (rc *ReplicationCoordinator) handleUpdateEvent(ctx context.Context, evt event.Update) error {
	// If this is a retry, we don't need to generate artifacts
	if evt.IsRetry {
		return nil
	}

	block, err := coreblock.GetFromBytes(evt.Block)
	if err != nil {
		return fmt.Errorf("failed to deserialize block: %w", err)
	}

	if !block.Delta.IsComposite() {
		return nil
	}

	updatedFields := []string{}
	for _, link := range block.Links {
		if link.Name != "" && link.Name != "_head" {
			updatedFields = append(updatedFields, link.Name)
		}
	}

	artifacts, err := rc.generateSEArtifacts(ctx, evt.DocID, evt.CollectionID, updatedFields)
	if err != nil {
		return fmt.Errorf("failed to generate SE artifacts: %w", err)
	}

	if len(artifacts) > 0 {
		rc.eventBus.Publish(event.NewMessage(ReplicateEventName, ReplicateEvent{
			DocID:        evt.DocID,
			CollectionID: evt.CollectionID,
			Artifacts:    artifacts,
		}))
	}

	return nil
}

// retrySEReplicators periodically processes failed SE replications
func (rc *ReplicationCoordinator) retrySEReplicators(ctx context.Context) {
	ticker := time.NewTicker(retryLoopInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rc.processSERetries(ctx)
		}
	}
}

// processSERetries checks for due retries and processes them
func (rc *ReplicationCoordinator) processSERetries(ctx context.Context) {
	ps := datastore.PeerstoreFrom(rc.db.Rootstore())
	iter, err := ps.Iterator(ctx, corekv.IterOptions{
		Prefix: keys.NewPeerstoreSERetry("", "", "").Bytes(),
	})
	if err != nil {
		log.ErrorContextE(ctx, "Failed to iterate SE retry keys", err)
		return
	}
	defer iter.Close()

	now := time.Now()
	for {
		hasNext, err := iter.Next()
		if err != nil {
			log.ErrorContextE(ctx, "Failed to get next SE retry key", err)
			break
		}
		if !hasNext {
			break
		}

		value, err := iter.Value()
		if err != nil {
			log.ErrorContextE(ctx, "Failed to get SE retry value", err)
			continue
		}

		retryInfo := SERetryInfo{}
		err = cbor.Unmarshal(value, &retryInfo)
		if err != nil {
			log.ErrorContextE(ctx, "Failed to unmarshal SE retry info", err)
			continue
		}

		// Check if retry is due and not already in progress
		if now.After(retryInfo.NextRetry) && !retryInfo.Retrying {
			key, err := keys.NewPeerstoreSERetryFromString(string(iter.Key()))
			if err != nil {
				log.ErrorContextE(ctx, "Failed to parse SE retry key", err)
				continue
			}

			retryInfo.Retrying = true
			retryInfo.NumRetries++
			b, err := cbor.Marshal(retryInfo)
			if err != nil {
				log.ErrorContextE(ctx, "Failed to marshal SE retry info", err)
				continue
			}
			ps := datastore.PeerstoreFrom(rc.db.Rootstore())
			if err := ps.Set(ctx, iter.Key(), b); err != nil {
				log.ErrorContextE(ctx, "Failed to update SE retry info", err)
				continue
			}

			go rc.retrySEArtifacts(ctx, key.PeerID, retryInfo)
		}
	}
}

// retrySEArtifacts attempts to retry SE artifact replication for a document
//
// Note: This function relies on the SE artifact generation phase to re-generate
// artifacts from the document's field values. We don't store SE artifacts locally
// on the producer node - they are only stored on replicator nodes.
func (rc *ReplicationCoordinator) retrySEArtifacts(ctx context.Context, peerID string, retryInfo SERetryInfo) {
	log.InfoContext(ctx, "Retrying SE artifact replication",
		corelog.String("PeerID", peerID),
		corelog.String("DocID", retryInfo.DocID),
		corelog.String("CollectionID", retryInfo.CollectionID))

	successChan := make(chan bool)
	defer close(successChan)

	artifacts, err := rc.generateSEArtifacts(ctx, retryInfo.DocID, retryInfo.CollectionID, retryInfo.FieldNames)
	if err != nil {
		log.ErrorContextE(ctx, "Failed to regenerate SE artifacts for retry", err)
		rc.updateRetryStatus(ctx, peerID, retryInfo, false)
		return
	}

	rc.eventBus.Publish(event.NewMessage(ReplicateEventName, ReplicateEvent{
		DocID:        retryInfo.DocID,
		CollectionID: retryInfo.CollectionID,
		Artifacts:    artifacts,
		IsRetry:      true,
		Success:      successChan,
	}))

	select {
	case success := <-successChan:
		rc.updateRetryStatus(ctx, peerID, retryInfo, success)
	case <-time.After(retryTimeout):
		log.ErrorContext(ctx, "SE artifact retry timeout",
			corelog.String("PeerID", peerID),
			corelog.String("DocID", retryInfo.DocID))
		rc.updateRetryStatus(ctx, peerID, retryInfo, false)
	}
}

// updateRetryStatus updates the retry status after an attempt
func (rc *ReplicationCoordinator) updateRetryStatus(ctx context.Context, peerID string, retryInfo SERetryInfo, success bool) {
	retryKey := keys.NewPeerstoreSERetry(peerID, retryInfo.CollectionID, retryInfo.DocID)

	if success {
		ps := datastore.PeerstoreFrom(rc.db.Rootstore())
		if err := ps.Delete(ctx, retryKey.Bytes()); err != nil {
			log.ErrorContextE(ctx, "Failed to delete SE retry entry", err)
		}
	} else {
		if retryInfo.NumRetries >= len(rc.retryIntervals) {
			retryInfo.NextRetry = time.Now().Add(rc.retryIntervals[len(rc.retryIntervals)-1])
		} else {
			retryInfo.NextRetry = time.Now().Add(rc.retryIntervals[retryInfo.NumRetries])
		}
		retryInfo.Retrying = false

		b, err := cbor.Marshal(retryInfo)
		if err != nil {
			log.ErrorContextE(ctx, "Failed to marshal SE retry info", err)
			return
		}
		ps := datastore.PeerstoreFrom(rc.db.Rootstore())
		if err := ps.Set(ctx, retryKey.Bytes(), b); err != nil {
			log.ErrorContextE(ctx, "Failed to update SE retry info", err)
		}
	}
}

// DeleteSEArtifacts removes SE artifacts from the datastore.
//
// Parameters:
//   - searchTags: If provided, only delete artifacts with these specific search tags.
//     If empty/nil, delete all artifacts for the given document/index combination.
//
// This is typically called when:
//   - A document is deleted (searchTags is empty)
//   - A field value changes (searchTags contains the old search tags to remove)
func (rc *ReplicationCoordinator) DeleteSEArtifacts(ctx context.Context, collectionID string, indexID string, docID string, searchTags [][]byte) error {
	ds := datastore.DatastoreFrom(rc.db.Rootstore())

	if len(searchTags) > 0 {
		for _, tag := range searchTags {
			key := keys.DatastoreSE{
				CollectionID: collectionID,
				IndexID:      indexID,
				SearchTag:    tag,
				DocID:        docID,
			}
			if err := ds.Delete(ctx, key.Bytes()); err != nil {
				return err
			}
		}
		return nil
	}

	prefix := keys.DatastoreSE{
		CollectionID: collectionID,
		IndexID:      indexID,
	}.Bytes()

	keysToDelete, err := datastore.FetchKeysForPrefix(ctx, prefix, ds)
	if err != nil {
		return err
	}

	for _, key := range keysToDelete {
		keyStr := string(key)
		if strings.HasSuffix(keyStr, "/"+docID) {
			if err := ds.Delete(ctx, key); err != nil {
				return err
			}
		}
	}

	return nil
}

// generateSEArtifacts regenerates SE artifacts for specified fields
//
// This method uses the extracted GenerateArtifacts function to recreate artifacts
// needed for retry.
func (rc *ReplicationCoordinator) generateSEArtifacts(
	ctx context.Context,
	docID, collectionID string,
	fieldNames []string,
) ([]secore.Artifact, error) {
	cols, err := rc.db.GetCollections(ctx, client.CollectionFetchOptions{
		CollectionID: immutable.Some(collectionID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}
	if len(cols) == 0 {
		return nil, fmt.Errorf("collection not found: %s", collectionID)
	}

	col := cols[0]
	docIDType, err := client.NewDocIDFromString(docID)
	if err != nil {
		return nil, fmt.Errorf("invalid document ID: %w", err)
	}

	doc, err := col.Get(ctx, docIDType, false)
	if err != nil {
		if errors.Is(err, client.ErrDocumentNotFoundOrNotAuthorized) {
			// TODO: Handle document deletion - generate delete artifacts
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return GenerateDocArtifacts(ctx, col, doc, fieldNames, rc.encKey)
}
