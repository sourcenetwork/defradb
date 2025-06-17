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
	"github.com/libp2p/go-libp2p/core/peer"
	secore "github.com/sourcenetwork/defradb/internal/se/core"
)

// Event names for the event bus
const (
	UpdateEventName             = "se-update"
	StoreArtifactsEventName     = "se-store-artifacts"
	ReplicationFailureEventName = "se-replication-failure"
)

// ReplicateEvent - Published when SE artifacts need replication
type ReplicateEvent struct {
	DocID        string
	CollectionID string
	Artifacts    []secore.Artifact
	IsRetry      bool
	Success      chan bool // Used for synchronous retry feedback
}

// StoreArtifactsEvent - Published when receiving SE artifacts from peers
type StoreArtifactsEvent struct {
	Artifacts []secore.Artifact
	FromPeer  peer.ID
}

// ReplicationFailureEvent - Published when artifact replication fails
type ReplicationFailureEvent struct {
	DocID        string
	CollectionID string
	PeerID       peer.ID
	FieldNames   []string
}
