// Copyright 2020 Source Inc.
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
	"github.com/sourcenetwork/defradb/core"
	corecrdt "github.com/sourcenetwork/defradb/core/crdt"
	"github.com/sourcenetwork/defradb/merkle/clock"

	// "github.com/sourcenetwork/defradb/store"

	"github.com/ipfs/go-cid"
)

var (
	lwwFactoryFn = MerkleCRDTFactory(func(mstore core.MultiStore) MerkleCRDTInitFn {
		return func(key core.DataStoreKey) MerkleCRDT {
			return NewMerkleLWWRegister(mstore.Datastore(), mstore.Headstore(), mstore.DAGstore(), core.DataStoreKey{}, key)
		}
	})
)

func init() {
	DefaultFactory.Register(core.LWW_REGISTER, &lwwFactoryFn)
}

// MerkleLWWRegister is a MerkleCRDT implementation of the LWWRegister
// using MerkleClocks
type MerkleLWWRegister struct {
	*baseMerkleCRDT
	// core.ReplicatedData

	reg   corecrdt.LWWRegister
	clock core.MerkleClock
}

// NewMerkleLWWRegister creates a new instance (or loaded from DB) of a MerkleCRDT
// backed by a LWWRegister CRDT
func NewMerkleLWWRegister(datastore core.DSReaderWriter, headstore core.DSReaderWriter, dagstore core.DAGStore, ns, key core.DataStoreKey) *MerkleLWWRegister {
	register := corecrdt.NewLWWRegister(datastore, key /* stuff like namespace and ID */)

	clock := clock.NewMerkleClock(headstore, dagstore, key.ToHeadStoreKey(), register)
	base := &baseMerkleCRDT{clock, register}

	return &MerkleLWWRegister{
		baseMerkleCRDT: base,
		reg:            register,
	}
}

// Set the value of the register
func (mlwwreg *MerkleLWWRegister) Set(value []byte) (cid.Cid, error) {
	// Set() call on underlying LWWRegister CRDT
	// persist/publish delta
	delta := mlwwreg.reg.Set(value)
	return mlwwreg.Publish(delta)
}

// Value will retrieve the current value from the db
func (mlwwreg *MerkleLWWRegister) Value() ([]byte, error) {
	return mlwwreg.reg.Value()
}

// Merge writes the provided delta to state using a supplied
// merge semantic
func (mlwwreg *MerkleLWWRegister) Merge(other core.Delta, id string) error {
	return mlwwreg.reg.Merge(other, id)
}
