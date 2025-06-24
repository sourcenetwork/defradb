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
	"github.com/sourcenetwork/defradb/event"
	secore "github.com/sourcenetwork/defradb/internal/se/core"
)

// Event names for the event bus
const (
	ReplicateEventName          = "se-replicate"
	ReplicationFailureEventName = "se-replication-failure"
	QuerySEArtifactsEventName   = "se-query-artifacts"
)

// ReplicateEvent - Published when SE artifacts need replication
type ReplicateEvent struct {
	DocID        string
	CollectionID string
	Artifacts    []secore.Artifact
	IsRetry      bool
	Success      chan bool // Used for synchronous retry feedback
}

// ReplicationFailureEvent - Published when artifact replication fails
type ReplicationFailureEvent struct {
	DocID        string
	CollectionID string
	PeerID       peer.ID
	FieldNames   []string
}

// QuerySEArtifactsRequest - Request to query SE artifacts from replicators
type QuerySEArtifactsRequest struct {
	CollectionID string
	Queries      []FieldQuery
	Response     chan QuerySEArtifactsResponse
}

// QuerySEArtifactsResponse - Response containing matching document IDs
type QuerySEArtifactsResponse struct {
	DocIDs []string
	Error  error
}

// NewQuerySEArtifactsMessage creates a new SE query message with response channel
func NewQuerySEArtifactsMessage(collectionID string, queries []FieldQuery) (event.Message, chan QuerySEArtifactsResponse) {
	response := make(chan QuerySEArtifactsResponse, 1)
	request := QuerySEArtifactsRequest{
		CollectionID: collectionID,
		Queries:      queries,
		Response:     response,
	}
	return event.NewMessage(QuerySEArtifactsEventName, request), response
}
