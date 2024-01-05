// Copyright 2023 Democratized Data Foundation
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
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/sourcenetwork/defradb/errors"
)

const (
	errPushLog                 = "failed to push log"
	errFailedToGetDocID        = "failed to get DocID from broadcast message"
	errPublishingToDocIDTopic  = "can't publish log %s for docID %s"
	errPublishingToSchemaTopic = "can't publish log %s for schema %s"
	errReplicatorExists        = "replicator already exists for %s with peerID %s"
	errReplicatorDocID         = "failed to get docID for replicator %s with peerID %s"
	errReplicatorCollections   = "failed to get collections for replicator"
)

var (
	ErrPeerConnectionWaitTimout = errors.New("waiting for peer connection timed out")
	ErrPubSubWaitTimeout        = errors.New("waiting for pubsub timed out")
	ErrPushLogWaitTimeout       = errors.New("waiting for pushlog timed out")
	ErrNilDB                    = errors.New("database object can't be nil")
	ErrNilUpdateChannel         = errors.New("tried to subscribe to update channel, but update channel is nil")
	ErrSelfTargetForReplicator  = errors.New("can't target ourselves as a replicator")
)

func NewErrPushLog(inner error, kv ...errors.KV) error {
	return errors.Wrap(errPushLog, inner, kv...)
}

func NewErrFailedToGetDocID(inner error, kv ...errors.KV) error {
	return errors.Wrap(errFailedToGetDocID, inner, kv...)
}

func NewErrPublishingToDocIDTopic(inner error, cid, docID string, kv ...errors.KV) error {
	return errors.Wrap(fmt.Sprintf(errPublishingToDocIDTopic, cid, docID), inner, kv...)
}

func NewErrPublishingToSchemaTopic(inner error, cid, docID string, kv ...errors.KV) error {
	return errors.Wrap(fmt.Sprintf(errPublishingToSchemaTopic, cid, docID), inner, kv...)
}

func NewErrReplicatorExists(collection string, peerID peer.ID, kv ...errors.KV) error {
	return errors.New(fmt.Sprintf(errReplicatorExists, collection, peerID), kv...)
}

func NewErrReplicatorDocID(inner error, collection string, peerID peer.ID, kv ...errors.KV) error {
	return errors.Wrap(fmt.Sprintf(errReplicatorDocID, collection, peerID), inner, kv...)
}

func NewErrReplicatorCollections(inner error, kv ...errors.KV) error {
	return errors.Wrap(errReplicatorCollections, inner, kv...)
}
