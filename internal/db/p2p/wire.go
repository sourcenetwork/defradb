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

// retryInfo is deliberately not registered: it is CBOR-encoded only into the
// local peerstore, never sent to a peer.
func init() {
	wire.Register[syncBranchableCollectionRequest]()
	wire.Register[syncBranchableCollectionReply]()
	wire.Register[docSyncRequest]()
	wire.Register[docSyncReply]()
}
