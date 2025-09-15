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
	"github.com/sourcenetwork/defradb/event"
)

// Event names for the event bus
const (
	QuerySEArtifactsEventName = "se-query-artifacts"
)

// RequestSEArtifactsEvent - Request to query SE artifacts from replicators
type RequestSEArtifactsEvent struct {
	CollectionID string
	Queries      []FieldQuery
	Response     chan SEArtifactsResult
}

// SEArtifactsResult - Response containing matching document IDs
type SEArtifactsResult struct {
	DocIDs []string
	Error  error
}

// NewQuerySEArtifactsMessage creates a new SE query message with response channel
func NewQuerySEArtifactsMessage(collectionID string, queries []FieldQuery) (event.Message, chan SEArtifactsResult) {
	response := make(chan SEArtifactsResult, 1)
	request := RequestSEArtifactsEvent{
		CollectionID: collectionID,
		Queries:      queries,
		Response:     response,
	}
	return event.NewMessage(QuerySEArtifactsEventName, request), response
}
