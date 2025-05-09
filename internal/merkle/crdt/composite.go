// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package merklecrdt

import (
	"context"

	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/defradb/client"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	corecrdt "github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/merkle/clock"
)

// MerkleCompositeDAG is a MerkleCRDT implementation of the CompositeDAG using MerkleClocks.
type MerkleCompositeDAG struct {
	clock *clock.MerkleClock
	// core.ReplicatedData
	reg corecrdt.CompositeDAG
}

var _ MerkleCRDT = (*MerkleCompositeDAG)(nil)

// NewMerkleCompositeDAG creates a new instance (or loaded from DB) of a MerkleCRDT
// backed by a CompositeDAG CRDT.
func NewMerkleCompositeDAG(
	store Stores,
	schemaVersionID string,
	key keys.DataStoreKey,
) *MerkleCompositeDAG {
	compositeDag := corecrdt.NewCompositeDAG(
		store.Datastore(),
		schemaVersionID,
		key,
	)

	clock := clock.NewMerkleClock(store.Headstore(), store.Blockstore(), store.Encstore(), key.ToHeadStoreKey(),
		compositeDag)

	return &MerkleCompositeDAG{
		clock: clock,
		reg:   compositeDag,
	}
}

func (m *MerkleCompositeDAG) Clock() *clock.MerkleClock {
	return m.clock
}

// Delete sets the values of CompositeDAG for a delete.
func (m *MerkleCompositeDAG) Delete(
	ctx context.Context,
) (cidlink.Link, []byte, error) {
	delta := m.reg.NewDelta(client.Deleted)
	return m.clock.AddDelta(ctx, delta)
}

// Save the value of the composite CRDT to DAG.
func (m *MerkleCompositeDAG) Save(ctx context.Context, links []coreblock.DAGLink) (cidlink.Link, []byte, error) {
	delta := m.reg.NewDelta(client.Active)
	return m.clock.AddDelta(ctx, delta, links...)
}
