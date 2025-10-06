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
	"time"

	"github.com/fxamacker/cbor/v2"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corelog"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
)

const (
	// retryLoopInterval is the interval at which the retry handler checks for
	// SE artifacts that are due for a retry. Same as document replicator.
	retryLoopInterval = 2 * time.Second
)

// SERetryInfo stores retry information for failed SE replications
type SERetryInfo struct {
	DocID        string
	CollectionID string
	FieldNames   []string
	NextRetry    time.Time
	NumRetries   int
	Retrying     bool
	PublicKey    string // Hex-encoded public key for identity reconstruction
	KeyType      string // Key type (secp256k1, ed25519, etc.)
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

// retrySEReplicators periodically processes failed SE replications
func (rc *Coordinator) retrySEReplicators(ctx context.Context) {
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
func (rc *Coordinator) processSERetries(ctx context.Context) {
	ps := datastore.PeerstoreFrom(rc.p2p.DB().Rootstore())
	iter, err := ps.Iterator(ctx, corekv.IterOptions{
		Prefix: keys.NewPeerstoreSERetry("", "", "").Bytes(),
	})
	if err != nil {
		log.ErrorContextE(ctx, "Failed to iterate SE retry keys", err)
		return
	}

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
			ps := datastore.PeerstoreFrom(rc.p2p.DB().Rootstore())
			if err := ps.Set(ctx, iter.Key(), b); err != nil {
				log.ErrorContextE(ctx, "Failed to update SE retry info", err)
				continue
			}

			go rc.retrySEArtifacts(ctx, key.PeerID, retryInfo)
		}
	}

	err = iter.Close()
	if err != nil {
		log.ErrorContextE(ctx, "Failed to close SE retry iterator", err)
	}
}

// retrySEArtifacts attempts to retry SE artifact replication for a document
//
// Note: This function relies on the SE artifact generation phase to re-generate
// artifacts from the document's field values. We don't store SE artifacts locally
// on the producer node - they are only stored on replicator nodes.
func (rc *Coordinator) retrySEArtifacts(ctx context.Context, peerID string, retryInfo SERetryInfo) {
	log.InfoContext(ctx, "Retrying SE replicator", corelog.String("PeerID", peerID))

	identity, err := rc.reconstructIdentity(retryInfo.PublicKey, retryInfo.KeyType)
	if err != nil {
		log.ErrorContextE(ctx, "Failed to reconstruct identity from stored data", err,
			corelog.String("DocID", retryInfo.DocID))
	} else if identity.HasValue() {
		ctx = acpIdentity.WithContext(ctx, identity)
	}

	err = rc.GenerateArtifactsAndPushToReplicators(ctx, retryInfo.DocID,
		retryInfo.CollectionID, retryInfo.FieldNames, identity, true)
	if err != nil {
		log.ErrorContextE(ctx, "Failed to generate and push SE artifacts for retry", err,
			corelog.String("DocID", retryInfo.DocID))
	}

	rc.updateRetryStatus(ctx, peerID, retryInfo, err == nil)
}

// updateRetryStatus updates the retry status after an attempt
func (rc *Coordinator) updateRetryStatus(
	ctx context.Context,
	peerID string,
	retryInfo SERetryInfo,
	success bool,
) {
	retryKey := keys.NewPeerstoreSERetry(peerID, retryInfo.CollectionID, retryInfo.DocID)

	if success {
		ps := datastore.PeerstoreFrom(rc.p2p.DB().Rootstore())
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
		ps := datastore.PeerstoreFrom(rc.p2p.DB().Rootstore())
		if err := ps.Set(ctx, retryKey.Bytes(), b); err != nil {
			log.ErrorContextE(ctx, "Failed to update SE retry info", err)
		}
	}
}
