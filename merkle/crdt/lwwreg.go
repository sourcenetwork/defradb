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
	"fmt"

	"github.com/sourcenetwork/defradb/core"
	corecrdt "github.com/sourcenetwork/defradb/core/crdt"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/merkle/clock"

	// "github.com/sourcenetwork/defradb/store"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
)

var (
	lwwFactoryFn = MerkleCRDTFactory(func(mstore core.MultiStore) MerkleCRDTInitFn {
		return func(key ds.Key) MerkleCRDT {
			return NewMerkleLWWRegister(mstore.Datastore(), mstore.Headstore(), mstore.DAGstore(), ds.NewKey(""), key)
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
func NewMerkleLWWRegister(datastore core.DSReaderWriter, headstore core.DSReaderWriter, dagstore core.DAGStore, ns, dockey ds.Key) *MerkleLWWRegister {
	// New Register
	reg := corecrdt.NewLWWRegister(datastore, ns, dockey.String() /* stuff like namespace and ID */)

	// New Clock
	// two possible cases here
	// 1) Primary index
	// 2) Versioned Index

	var headsetKey ds.Key
	list := dockey.List()[1:] // remove collection identifier
	if list[0] == fmt.Sprint(base.PrimaryIndex) {
		// strip collection/index identifier from docKey, and any trailing
		// data AFTER the docKey.
		headsetKey = ds.KeyWithNamespaces(list[1:])
	} else if list[0] == fmt.Sprint(base.VersionIndex) {
		// splice out the Version CID component of the
		// VersionIndex compound index key.
		// Currently, the key is in the following format
		// /VersionIndexID/DocKey/VersionCID/.../FieldIdentifer
		//
		// We want to remove the VersionIndexID and the VersionCID, but keep the rest.
		headsetKey = ds.KeyWithNamespaces(append(list[1:2], list[3:]...))
	} else {
		// error, lets panic for now. TODO: FIX THIS
		panic("invalid index identifier for Merkle CRDT")
	}

	clk := clock.NewMerkleClock(headstore, dagstore, headsetKey.String(), reg)
	// newBaseMerkleCRDT(clock, register)
	base := &baseMerkleCRDT{clk, reg}
	// instatiate MerkleLWWRegister
	// return
	return &MerkleLWWRegister{
		baseMerkleCRDT: base,
		// clock:          clk,
		reg: reg,
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
