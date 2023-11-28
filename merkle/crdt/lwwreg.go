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

	"github.com/sourcenetwork/defradb/core"
	corecrdt "github.com/sourcenetwork/defradb/core/crdt"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/merkle/clock"
)

// MerkleLWWRegister is a MerkleCRDT implementation of the LWWRegister using MerkleClocks.
type MerkleLWWRegister struct {
	*baseMerkleCRDT

	reg corecrdt.LWWRegister
}

// NewMerkleLWWRegister creates a new instance (or loaded from DB) of a MerkleCRDT
// backed by a LWWRegister CRDT.
func NewMerkleLWWRegister(
	txn datastore.Txn,
	schemaVersionKey core.CollectionSchemaVersionKey,
	key core.DataStoreKey,
	fieldName string,
) *MerkleLWWRegister {
	register := corecrdt.NewLWWRegister(txn.Datastore(), schemaVersionKey, key, fieldName)
	clk := clock.NewMerkleClock(txn.Headstore(), txn.DAGstore(), key.ToHeadStoreKey(), register)
	base := &baseMerkleCRDT{clock: clk, crdt: register}
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
	nd, err := mlwwreg.clock.AddDAGNode(ctx, delta)
	return nd, delta.GetPriority(), err
}
