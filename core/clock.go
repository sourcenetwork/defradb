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

// MerkleClock is the core logical clock implementation that manages
// writing to and from the MerkleDAG structure, ensuring a casual
// ordering of
type MerkleClock interface {
	AddDAGNode(ctx context.Context, delta Delta) (cid.Cid, ipld.Node, error) // possibly change to AddDeltaNode?
	ProcessNode(context.Context, NodeGetter, cid.Cid, uint64, Delta, ipld.Node) ([]cid.Cid, error)
}
