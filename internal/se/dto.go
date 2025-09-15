// Copyright (20\d\d) Democratized Data Foundation
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

type QuerySEArtifactsRequest struct {
	message.MetaData
	CollectionID string
	Queries      []SEFieldQuery
}

// seFieldQuery - Query for a specific field
type SEFieldQuery struct {
	FieldName string
	IndexID   string
	SearchTag []byte
}

// QuerySEArtifactsReply - Reply with matching document IDs
type QuerySEArtifactsReply struct {
	message.MetaData
	DocIDs []string
}

// PushSEArtifactsRequest - Request to push SE artifacts
type PushSEArtifactsRequest struct {
	message.MetaData
	CollectionID string
	Artifacts    []SEArtifact
}

// SEArtifact - Network representation
type SEArtifact struct {
	DocID     string
	IndexID   string
	SearchTag []byte
}

// Reply type
type PushSEArtifactsReply struct {
	message.MetaData
}
