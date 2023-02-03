// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import "github.com/libp2p/go-libp2p/core/peer"

// Replicator is a peer that a set of local collections are replicated to.
type Replicator struct {
	Info    peer.AddrInfo
	Schemas []string
}
