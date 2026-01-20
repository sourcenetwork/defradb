// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package protocol

import (
	"github.com/sourcenetwork/defradb/internal/db/p2p/message"
)

// PushLogRequest is the struct used to send a resource update to a peer node.
type PushLogRequest struct {
	message.MetaData
	// DocID is the document ID for single-document requests.
	// Empty when Documents is populated (batched mode).
	DocID string
	// CID is the root CID for single-document requests.
	// Empty when Documents is populated (batched mode).
	CID []byte
	// CollectionID is the collection ID. Required for both modes.
	CollectionID string
	// Creator is the peer ID of the original creator.
	Creator string
	// Block is the root block data for single-document requests.
	// Empty when Documents is populated (batched mode).
	Block []byte
	// CAR contains the DAG blocks in CAR format.
	// For batched mode, contains blocks for ALL documents in Documents slice.
	CAR []byte
	// Documents contains multiple document infos for batched requests.
	// When populated, DocID/CID/Block should be empty.
	Documents []DocumentInfo
}

// DocumentInfo contains the metadata for a single document in a batch.
type DocumentInfo struct {
	DocID string
	CID   []byte
	Block []byte
}

// IsBatched returns true if this is a batched request with multiple documents.
func (r *PushLogRequest) IsBatched() bool {
	return len(r.Documents) > 0
}

// PushLogReply is the expected response struct that should be received after
// an pushlog request.
type PushLogReply struct {
	message.MetaData
}
