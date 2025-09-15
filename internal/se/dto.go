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

import "github.com/sourcenetwork/defradb/internal/db/p2p/message"

// QuerySEArtifactsRequest is the request object to query SE artifacts
type QuerySEArtifactsRequest struct {
	message.MetaData
	CollectionID string
	Queries      []SEFieldQuery
}

// SEFieldQuery is the SE query object for a specific field
type SEFieldQuery struct {
	FieldName string
	IndexID   string
	SearchTag []byte
}

// QuerySEArtifactsReply is the reply object  with matching document IDs for [QuerySEArtifactsRequest] query
type QuerySEArtifactsReply struct {
	message.MetaData
	DocIDs []string
}

// PushSEArtifactsRequest is the request object to push SE artifacts
type PushSEArtifactsRequest struct {
	message.MetaData
	CollectionID string
	Artifacts    []SEArtifact
}

// SEArtifact is the SE artifact object to be pushed
type SEArtifact struct {
	DocID     string
	IndexID   string
	SearchTag []byte
}

// PushSEArtifactsReply is the reply object for [PushSEArtifactsRequest] query
type PushSEArtifactsReply struct {
	message.MetaData
}
