// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package crdt

import (
	"context"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	corecrdt "github.com/sourcenetwork/defradb/core/crdt"
	corenet "github.com/sourcenetwork/defradb/core/net"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/merkle/clock"

	"github.com/ipfs/go-cid"
)

var (
	compFactoryFn = MerkleCRDTFactory(func(mstore datastore.MultiStore, schemaID string, bs corenet.Broadcaster) MerkleCRDTInitFn {
		return func(key core.DataStoreKey) MerkleCRDT {
			return NewMerkleCompositeDAG(mstore.Datastore(), mstore.Headstore(), mstore.DAGstore(), schemaID, bs, core.DataStoreKey{}, key)
		}
	})
)

func init() {
	err := DefaultFactory.Register(client.COMPOSITE, &compFactoryFn)
	if err != nil {
		panic(err)
	}
}

// MerkleCompositeDAG is a MerkleCRDT implementation of the CompositeDAG
// using MerkleClocks
type MerkleCompositeDAG struct {
	*baseMerkleCRDT
	// core.ReplicatedData
	reg corecrdt.CompositeDAG
}

// NewMerkleCompositeDAG creates a new instance (or loaded from DB) of a MerkleCRDT
// backed by a CompositeDAG CRDT
func NewMerkleCompositeDAG(datastore datastore.DSReaderWriter, headstore datastore.DSReaderWriter, dagstore datastore.DAGStore, schemaID string, bs corenet.Broadcaster, ns, key core.DataStoreKey) *MerkleCompositeDAG {
	compositeDag := corecrdt.NewCompositeDAG(datastore, schemaID, ns, key.ToString() /* stuff like namespace and ID */)

	clock := clock.NewMerkleClock(headstore, dagstore, key.ToHeadStoreKey(), compositeDag)
	base := &baseMerkleCRDT{clock: clock, crdt: compositeDag, broadcaster: bs}

	return &MerkleCompositeDAG{
		baseMerkleCRDT: base,
		reg:            compositeDag,
	}
}

// Set sets the values of CompositeDAG.
// The value is always the object from the
// mutation operations.
func (m *MerkleCompositeDAG) Set(ctx context.Context, patch []byte, links []core.DAGLink) (cid.Cid, error) {
	// Set() call on underlying CompositeDAG CRDT
	// persist/publish delta
	log.Debug(ctx, "Applying delta-mutator 'Set' on CompositeDAG")
	delta := m.reg.Set(patch, links)
	c, nd, err := m.Publish(ctx, delta)
	if err != nil {
		return cid.Undef, err
	}

	return c, m.Broadcast(ctx, nd, delta)
}

// Value is a no-op for a CompositeDAG
func (m *MerkleCompositeDAG) Value(ctx context.Context) ([]byte, error) {
	return m.reg.Value(ctx)
}

// Merge writes the provided delta to state using a supplied
// merge semantic
// @todo
func (m *MerkleCompositeDAG) Merge(ctx context.Context, other core.Delta, id string) error {
	return m.reg.Merge(ctx, other, id)
}
