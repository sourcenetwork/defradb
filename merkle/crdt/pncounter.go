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

// MerklePNCounterRegister is a MerkleCRDT implementation of the PNCounterRegister using MerkleClocks.
type MerklePNCounterRegister struct {
	*baseMerkleCRDT

	reg corecrdt.PNCounterRegister
}

// NewMerklePNCounterRegister creates a new instance (or loaded from DB) of a MerkleCRDT
// backed by a PNCounterRegister CRDT.
func NewMerklePNCounterRegister(
	store Stores,
	schemaVersionKey core.CollectionSchemaVersionKey,
	key core.DataStoreKey,
	fieldName string,
) *MerklePNCounterRegister {
	register := corecrdt.NewPNCounterRegister(store.Datastore(), schemaVersionKey, key, fieldName)
	clk := clock.NewMerkleClock(store.Headstore(), store.DAGstore(), key.ToHeadStoreKey(), register)
	base := &baseMerkleCRDT{clock: clk, crdt: register}
	return &MerklePNCounterRegister{
		baseMerkleCRDT: base,
		reg:            register,
	}
}

// Add the value of the register.
func (mPNC *MerklePNCounterRegister) Add(ctx context.Context, value client.FieldValue) (ipld.Node, uint64, error) {
	delta := mPNC.reg.Set(value)
	nd, err := mPNC.clock.AddDAGNode(ctx, delta)
	return nd, delta.GetPriority(), err
}
