// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package core

import (
	"context"

	cid "github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
)

// NodeDeltaPair is a Node with its underlying delta already extracted.
// Used in a channel response for streaming.
type NodeDeltaPair interface {
	GetNode() ipld.Node
	GetDelta() Delta
	Error() error
}

// A NodeGetter extended from ipld.NodeGetter with delta-related functions.
type NodeGetter interface {
	ipld.NodeGetter
	GetDelta(context.Context, cid.Cid) (ipld.Node, Delta, error)
	GetDeltas(context.Context, []cid.Cid) <-chan NodeDeltaPair
	GetPriority(context.Context, cid.Cid) (uint64, error)
}
