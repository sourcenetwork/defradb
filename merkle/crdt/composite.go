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

	ipld "github.com/ipfs/go-ipld-format"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	corecrdt "github.com/sourcenetwork/defradb/core/crdt"
	"github.com/sourcenetwork/defradb/merkle/clock"
)

// MerkleCompositeDAG is a MerkleCRDT implementation of the CompositeDAG using MerkleClocks.
type MerkleCompositeDAG struct {
	*baseMerkleCRDT
	// core.ReplicatedData
	reg corecrdt.CompositeDAG
}

// NewMerkleCompositeDAG creates a new instance (or loaded from DB) of a MerkleCRDT
// backed by a CompositeDAG CRDT.
func NewMerkleCompositeDAG(
	store Stores,
	schemaVersionKey core.CollectionSchemaVersionKey,
	key core.DataStoreKey,
	fieldName string,
) *MerkleCompositeDAG {
	compositeDag := corecrdt.NewCompositeDAG(
		store.Datastore(),
		schemaVersionKey,
		key,
		fieldName,
	)

	clock := clock.NewMerkleClock(store.Headstore(), store.DAGstore(), key.ToHeadStoreKey(), compositeDag)
	base := &baseMerkleCRDT{clock: clock, crdt: compositeDag}

	return &MerkleCompositeDAG{
		baseMerkleCRDT: base,
		reg:            compositeDag,
	}
}

// Delete sets the values of CompositeDAG for a delete.
func (m *MerkleCompositeDAG) Delete(
	ctx context.Context,
	links []core.DAGLink,
) (ipld.Node, uint64, error) {
	// Set() call on underlying CompositeDAG CRDT
	// persist/publish delta
	log.DebugContext(ctx, "Applying delta-mutator 'Delete' on CompositeDAG")
	delta := m.reg.Set(links)
	delta.Status = client.Deleted
	nd, err := m.clock.AddDAGNode(ctx, delta)
	if err != nil {
		return nil, 0, err
	}

	return nd, delta.GetPriority(), nil
}

// Save the value of the composite CRDT to DAG.
func (m *MerkleCompositeDAG) Save(ctx context.Context, data any) (ipld.Node, uint64, error) {
	value, ok := data.([]core.DAGLink)
	if !ok {
		return nil, 0, NewErrUnexpectedValueType(client.COMPOSITE, []core.DAGLink{}, data)
	}
	// Set() call on underlying CompositeDAG CRDT
	// persist/publish delta
	log.DebugContext(ctx, "Applying delta-mutator 'Set' on CompositeDAG")
	delta := m.reg.Set(value)
	nd, err := m.clock.AddDAGNode(ctx, delta)
	if err != nil {
		return nil, 0, err
	}

	return nd, delta.GetPriority(), nil
}
