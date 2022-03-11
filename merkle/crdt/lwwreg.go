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
	"github.com/sourcenetwork/defradb/merkle/clock"

	// "github.com/sourcenetwork/defradb/datastore"

	"github.com/ipfs/go-cid"
)

var (
	lwwFactoryFn = MerkleCRDTFactory(func(mstore client.MultiStore, _ string, _ corenet.Broadcaster) MerkleCRDTInitFn {
		return func(key core.DataStoreKey) MerkleCRDT {
			return NewMerkleLWWRegister(mstore.Datastore(), mstore.Headstore(), mstore.DAGstore(), core.DataStoreKey{}, key)
		}
	})
)

func init() {
	err := DefaultFactory.Register(core.LWW_REGISTER, &lwwFactoryFn)
	if err != nil {
		panic(err)
	}
}

// MerkleLWWRegister is a MerkleCRDT implementation of the LWWRegister
// using MerkleClocks
type MerkleLWWRegister struct {
	*baseMerkleCRDT
	// core.ReplicatedData

	reg corecrdt.LWWRegister
}

// NewMerkleLWWRegister creates a new instance (or loaded from DB) of a MerkleCRDT
// backed by a LWWRegister CRDT
func NewMerkleLWWRegister(datastore client.DSReaderWriter, headstore client.DSReaderWriter, dagstore client.DAGStore, ns, key core.DataStoreKey) *MerkleLWWRegister {
	register := corecrdt.NewLWWRegister(datastore, key /* stuff like namespace and ID */)
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

// Set the value of the register
func (mlwwreg *MerkleLWWRegister) Set(ctx context.Context, value []byte) (cid.Cid, error) {
	// Set() call on underlying LWWRegister CRDT
	// persist/publish delta
	delta := mlwwreg.reg.Set(value)
	c, _, err := mlwwreg.Publish(ctx, delta)
	return c, err
}

// Value will retrieve the current value from the db
func (mlwwreg *MerkleLWWRegister) Value(ctx context.Context) ([]byte, error) {
	return mlwwreg.reg.Value(ctx)
}

// Merge writes the provided delta to state using a supplied
// merge semantic
func (mlwwreg *MerkleLWWRegister) Merge(ctx context.Context, other core.Delta, id string) error {
	return mlwwreg.reg.Merge(ctx, other, id)
}
