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

// CARRequest asks a peer for the CARs rooted at the given head CIDs. A receiver sends it only
// for heads it has checked it needs, after an update announced them without a CAR.
type CARRequest struct {
	message.MetaData
	CIDs [][]byte
}

// CARReply answers a CARRequest. CARs is parallel to a prefix of the request's CIDs: an empty
// entry is a head the peer could not or would not serve, which the requester walks the DAG for
// instead, and heads past the end of CARs were not attempted and may be asked for again.
type CARReply struct {
	message.MetaData
	CARs [][]byte
}
