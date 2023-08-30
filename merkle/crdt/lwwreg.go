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
	lwwFactoryFn = MerkleCRDTFactory(
		func(
			mstore datastore.MultiStore,
			schemaID core.CollectionSchemaVersionKey,
			_ events.UpdateChannel,
			fieldName string,
		) MerkleCRDTInitFn {
			return func(key core.DataStoreKey) MerkleCRDT {
				return NewMerkleLWWRegister(
					mstore.Datastore(),
					mstore.Headstore(),
					mstore.DAGstore(),
					schemaID,
					core.DataStoreKey{},
					key,
					fieldName,
				)
			}
		},
	)
)

func init() {
	err := DefaultFactory.Register(client.LWW_REGISTER, &lwwFactoryFn)
	if err != nil {
		panic(err)
	}
}

// MerkleLWWRegister is a MerkleCRDT implementation of the LWWRegister using MerkleClocks.
type MerkleLWWRegister struct {
	*baseMerkleCRDT
	// core.ReplicatedData

	reg corecrdt.LWWRegister
}

// NewMerkleLWWRegister creates a new instance (or loaded from DB) of a MerkleCRDT
// backed by a LWWRegister CRDT.
func NewMerkleLWWRegister(
	datastore datastore.DSReaderWriter,
	headstore datastore.DSReaderWriter,
	dagstore datastore.DAGStore,
	schemaVersionKey core.CollectionSchemaVersionKey,
	ns, key core.DataStoreKey,
	fieldName string,
) *MerkleLWWRegister {
	register := corecrdt.NewLWWRegister(datastore, schemaVersionKey, key, fieldName /* stuff like namespace and ID */)
	clk := clock.NewMerkleClock(headstore, dagstore, key.ToHeadStoreKey(), register)

	// newBaseMerkleCRDT(clock, register)
	base := &baseMerkleCRDT{clock: clk, crdt: register}
	// instantiate MerkleLWWRegister
	// return
	return &MerkleLWWRegister{
		baseMerkleCRDT: base,
		reg:            register,
	}
}

// Set the value of the register.
func (mlwwreg *MerkleLWWRegister) Set(ctx context.Context, value []byte) (ipld.Node, uint64, error) {
	// Set() call on underlying LWWRegister CRDT
	// persist/publish delta
	delta := mlwwreg.reg.Set(value)
	nd, err := mlwwreg.Publish(ctx, delta)
	return nd, delta.GetPriority(), err
}

// Value will retrieve the current value from the db.
func (mlwwreg *MerkleLWWRegister) Value(ctx context.Context) ([]byte, error) {
	return mlwwreg.reg.Value(ctx)
}

// Merge writes the provided delta to state using a supplied
// merge semantic.
func (mlwwreg *MerkleLWWRegister) Merge(ctx context.Context, other core.Delta) error {
	return mlwwreg.reg.Merge(ctx, other)
}
