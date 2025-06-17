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
	"strings"
	"time"

	"github.com/fxamacker/cbor/v2"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/event"
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
	db             DB // Interface to access datastore
	eventBus       *event.Bus
	mergeSub       *event.Subscription
	failureSub     *event.Subscription
	retryIntervals []time.Duration
}

// DB interface required by ReplicationCoordinator
type DB interface {
	Datastore() datastore.DSReaderWriter
	Peerstore() datastore.DSReaderWriter
	Events() *event.Bus
	MaxTxnRetries() int
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
func NewReplicationCoordinator(db DB) (*ReplicationCoordinator, error) {
	rc := &ReplicationCoordinator{
		db:             db,
		eventBus:       db.Events(),
		retryIntervals: defaultRetryIntervals(db.MaxTxnRetries()),
	}

	sub, err := db.Events().Subscribe(MergeEventName)
	if err != nil {
		return nil, err
	}
	rc.mergeSub = sub

	failureSub, err := db.Events().Subscribe(ReplicationFailureEventName)
	if err != nil {
		return nil, err
	}
	rc.failureSub = failureSub

	go rc.processMergeEvents()
	go rc.processFailureEvents()

	go rc.retrySEReplicators(context.Background())

	return rc, nil
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

// processMergeEvents handles incoming SE artifacts from peers
func (rc *ReplicationCoordinator) processMergeEvents() {
	for {
		msg, isOpen := <-rc.mergeSub.Message()
		if !isOpen {
			return
		}

		if evt, ok := msg.Data.(MergeEvent); ok {
			if err := rc.storeSEArtifacts(context.Background(), evt.Artifacts); err != nil {
				log.ErrorE("Failed to store SE artifacts", err)
			}
		}
	}
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

// handleReplicationFailure stores failed SE replication attempt for retry
func (rc *ReplicationCoordinator) handleReplicationFailure(ctx context.Context, evt ReplicationFailureEvent) error {
	// Store retry information similar to document replicator
	retryKey := keys.NewPeerstoreSERetry(evt.PeerID.String(), evt.CollectionID, evt.DocID)

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

	return rc.db.Peerstore().Set(ctx, retryKey.Bytes(), b)
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
	iter, err := rc.db.Peerstore().Iterator(ctx, corekv.IterOptions{
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
			if err := rc.db.Peerstore().Set(ctx, iter.Key(), b); err != nil {
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

	// Regenerate artifacts from document fields
	artifacts, err := rc.regenerateArtifactsForRetry(ctx, retryInfo)
	if err != nil {
		log.ErrorContextE(ctx, "Failed to regenerate SE artifacts for retry", err)
		rc.updateRetryStatus(ctx, peerID, retryInfo, false)
		return
	}

	// Publish the retry update event
	rc.eventBus.Publish(event.NewMessage(UpdateEventName, UpdateEvent{
		DocID:        retryInfo.DocID,
		CollectionID: retryInfo.CollectionID,
		Artifacts:    artifacts,
		IsRetry:      true,
		Success:      successChan,
	}))

	// Wait for retry result
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
		if err := rc.db.Peerstore().Delete(ctx, retryKey.Bytes()); err != nil {
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
		if err := rc.db.Peerstore().Set(ctx, retryKey.Bytes(), b); err != nil {
			log.ErrorContextE(ctx, "Failed to update SE retry info", err)
		}
	}
}

// storeSEArtifacts stores artifacts in datastore
//
// This stores SE artifacts on the replicator node to enable encrypted search queries.
// The artifacts contain only the encrypted search tags and document IDs - no actual
// field values or encryption keys are stored.
func (rc *ReplicationCoordinator) storeSEArtifacts(ctx context.Context, artifacts []secore.Artifact) error {
	ds := rc.db.Datastore()

	for _, artifact := range artifacts {
		key := keys.DatastoreSE{
			CollectionID: artifact.CollectionID,
			IndexID:      artifact.IndexID,
			SearchTag:    artifact.SearchTag,
			DocID:        artifact.DocID,
		}

		// Store empty value - we only need the key for search lookups
		if err := ds.Set(ctx, key.Bytes(), []byte{}); err != nil {
			return err
		}
	}

	return nil
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
	ds := rc.db.Datastore()

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

// regenerateArtifactsForRetry regenerates SE artifacts for specified fields
//
// This method fetches the current document fields and regenerates the SE artifacts
// needed for retry. It's similar to ProcessBlock but works with document fields
// instead of blocks.
func (rc *ReplicationCoordinator) regenerateArtifactsForRetry(ctx context.Context, retryInfo SERetryInfo) ([]secore.Artifact, error) {
	// TODO: This needs to be implemented with access to:
	// 1. Collection configuration (encrypted fields)
	// 2. Document fetcher to get current field values
	// 3. SE key for tag generation
	//
	// For now, return empty to avoid compilation errors
	log.InfoContext(ctx, "SE artifact regeneration not yet implemented",
		corelog.String("DocID", retryInfo.DocID),
		corelog.Any("FieldNames", retryInfo.FieldNames))

	return []secore.Artifact{}, nil
}

// Close stops the coordinator
func (rc *ReplicationCoordinator) Close() {
	if rc.mergeSub != nil {
		rc.eventBus.Unsubscribe(rc.mergeSub)
	}
	if rc.failureSub != nil {
		rc.eventBus.Unsubscribe(rc.failureSub)
	}
}
