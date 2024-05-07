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
	"github.com/sourcenetwork/defradb/internal/core"
	corecrdt "github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/merkle/clock"
)

// MerkleLWWRegister is a MerkleCRDT implementation of the LWWRegister using MerkleClocks.
type MerkleLWWRegister struct {
	*baseMerkleCRDT

	reg corecrdt.LWWRegister
}

// NewMerkleLWWRegister creates a new instance (or loaded from DB) of a MerkleCRDT
// backed by a LWWRegister CRDT.
func NewMerkleLWWRegister(
	store Stores,
	schemaVersionKey core.CollectionSchemaVersionKey,
	key core.DataStoreKey,
	fieldName string,
) *MerkleLWWRegister {
	register := corecrdt.NewLWWRegister(store.Datastore(), schemaVersionKey, key, fieldName)
	clk := clock.NewMerkleClock(store.Headstore(), store.DAGstore(), key.ToHeadStoreKey(), register)
	base := &baseMerkleCRDT{clock: clk, crdt: register}
	return &MerkleLWWRegister{
		baseMerkleCRDT: base,
		reg:            register,
	}
}

// Save the value of the register to the DAG.
func (mlwwreg *MerkleLWWRegister) Save(ctx context.Context, data any) (ipld.Node, uint64, error) {
	value, ok := data.(*client.FieldValue)
	if !ok {
		return nil, 0, NewErrUnexpectedValueType(client.LWW_REGISTER, &client.FieldValue{}, data)
	}
	bytes, err := value.Bytes()
	if err != nil {
		return nil, 0, err
	}
	// Set() call on underlying LWWRegister CRDT
	// persist/publish delta
	delta := mlwwreg.reg.Set(bytes)
	nd, err := mlwwreg.clock.AddDAGNode(ctx, delta)
	return nd, delta.GetPriority(), err
}
