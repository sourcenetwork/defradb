// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package p2p

import "github.com/sourcenetwork/defradb/internal/wire"

func init() {
	wire.Register[syncBranchableCollectionRequest]()
	wire.Register[syncBranchableCollectionReply]()
	wire.Register[docSyncRequest]()
	wire.Register[docSyncReply]()

	// Retry bookkeeping stored in the local peerstore, not sent to a peer.
	wire.MarkLocal[retryInfo]()
}
