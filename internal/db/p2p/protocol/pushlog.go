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

// DocumentInfo carries the per-document payload within a batched PushLogRequest.
type DocumentInfo struct {
	DocID string
	CID   []byte
	Block []byte
	CAR   []byte `cbor:"car,omitempty"`
}

// PushLogRequest is the struct used to send a resource update to a peer node.
// It may carry either a single document (DocID/CID/Block/CAR) or a batch of
// documents (Documents).  The receiver handles whichever is present.
type PushLogRequest struct {
	message.MetaData
	DocID        string
	CID          []byte
	CollectionID string
	Creator      string
	Block        []byte
	// CAR contains a Content Addressable aRchive of the full DAG rooted at CID.
	// When present, the receiver imports it directly without a round-trip DAG sync.
	CAR []byte `cbor:"car,omitempty"`
	// Documents, when non-empty, carries a batch of per-document payloads.
	// DocID/CID/Block/CAR fields are unused when this is set.
	Documents []DocumentInfo `cbor:"documents,omitempty"`
}

// PushLogReply is the expected response struct that should be received after
// an pushlog request.
type PushLogReply struct {
	message.MetaData
}
