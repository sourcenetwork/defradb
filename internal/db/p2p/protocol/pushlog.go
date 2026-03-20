// Copyright 2026 Democratized Data Foundation
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
	// CollectionID is the collection ID.
	CollectionID string
	// Creator is the peer ID of the original creator.
	Creator string
	// Documents contains document infos (always used, even for single doc).
	Documents []DocumentInfo
}

// DocumentInfo contains the metadata for a single document.
type DocumentInfo struct {
	DocID string
	CID   []byte
	Block []byte
	CAR   []byte // Per-doc CAR data containing DAG blocks for this document
}

// PushLogReply is the expected response struct that should be received after
// an pushlog request.
type PushLogReply struct {
	message.MetaData
}
