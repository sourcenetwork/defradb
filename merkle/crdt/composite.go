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

	ipld "github.com/ipfs/go-ipld-format"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	corecrdt "github.com/sourcenetwork/defradb/core/crdt"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/events"
	"github.com/sourcenetwork/defradb/merkle/clock"
)

var (
	compFactoryFn = MerkleCRDTFactory(
		func(
			mstore datastore.MultiStore,
			schemaID core.CollectionSchemaVersionKey,
			uCh events.UpdateChannel,
			fieldName string,
		) MerkleCRDTInitFn {
			return func(key core.DataStoreKey) MerkleCRDT {
				return NewMerkleCompositeDAG(
					mstore.Datastore(),
					mstore.Headstore(),
					mstore.DAGstore(),
					schemaID,
					uCh,
					core.DataStoreKey{},
					key,
					fieldName,
				)
			}
		},
	)
)

func init() {
	err := DefaultFactory.Register(client.COMPOSITE, &compFactoryFn)
	if err != nil {
		panic(err)
	}
}

// MerkleCompositeDAG is a MerkleCRDT implementation of the CompositeDAG using MerkleClocks.
type MerkleCompositeDAG struct {
	*baseMerkleCRDT
	// core.ReplicatedData
	reg corecrdt.CompositeDAG
}

// NewMerkleCompositeDAG creates a new instance (or loaded from DB) of a MerkleCRDT
// backed by a CompositeDAG CRDT.
func NewMerkleCompositeDAG(
	datastore datastore.DSReaderWriter,
	headstore datastore.DSReaderWriter,
	dagstore datastore.DAGStore,
	schemaVersionKey core.CollectionSchemaVersionKey,
	uCh events.UpdateChannel,
	ns,
	key core.DataStoreKey,
	fieldName string,
) *MerkleCompositeDAG {
	compositeDag := corecrdt.NewCompositeDAG(
		datastore,
		schemaVersionKey,
		ns,
		key, /* stuff like namespace and ID */
		fieldName,
	)

	clock := clock.NewMerkleClock(headstore, dagstore, key.ToHeadStoreKey(), compositeDag)
	base := &baseMerkleCRDT{clock: clock, crdt: compositeDag, updateChannel: uCh}

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
	log.Debug("Applying delta-mutator 'Delete' on CompositeDAG")
	delta := m.reg.Set([]byte{}, links)
	delta.Status = client.Deleted
	nd, err := m.Publish(ctx, delta)
	if err != nil {
		return nil, 0, err
	}

	return nd, delta.GetPriority(), nil
}

// Set sets the values of CompositeDAG. The value is always the object from the mutation operations.
func (m *MerkleCompositeDAG) Set(
	ctx context.Context,
	patch []byte,
	links []core.DAGLink,
) (ipld.Node, uint64, error) {
	// Set() call on underlying CompositeDAG CRDT
	// persist/publish delta
	log.Debug("Applying delta-mutator 'Set' on CompositeDAG")
	delta := m.reg.Set(patch, links)
	nd, err := m.Publish(ctx, delta)
	if err != nil {
		return nil, 0, err
	}

	return nd, delta.GetPriority(), nil
}

// Value is a no-op for a CompositeDAG.
func (m *MerkleCompositeDAG) Value(ctx context.Context) ([]byte, error) {
	return m.reg.Value(ctx)
}

// Merge writes the provided delta to state using a supplied merge semantic.
// @todo
func (m *MerkleCompositeDAG) Merge(ctx context.Context, other core.Delta) error {
	return m.reg.Merge(ctx, other)
}
