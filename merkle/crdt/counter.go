// Copyright 2024 Democratized Data Foundation
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
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/merkle/clock"
)

// MerkleCounter is a MerkleCRDT implementation of the Counter using MerkleClocks.
type MerkleCounter[T crdt.Incrementable] struct {
	*baseMerkleCRDT

	reg crdt.Counter[T]
}

// NewMerkleCounter creates a new instance (or loaded from DB) of a MerkleCRDT
// backed by a Counter CRDT.
func NewMerkleCounter[T crdt.Incrementable](
	store Stores,
	schemaVersionKey core.CollectionSchemaVersionKey,
	key core.DataStoreKey,
	fieldName string,
	allowDecrement bool,
) *MerkleCounter[T] {
	register := crdt.NewCounter[T](store.Datastore(), schemaVersionKey, key, fieldName, allowDecrement)
	clk := clock.NewMerkleClock(store.Headstore(), store.DAGstore(), key.ToHeadStoreKey(), register)
	base := &baseMerkleCRDT{clock: clk, crdt: register}
	return &MerkleCounter[T]{
		baseMerkleCRDT: base,
		reg:            register,
	}
}

// Save the value of the  Counter to the DAG.
func (mc *MerkleCounter[T]) Save(ctx context.Context, data any) (ipld.Node, uint64, error) {
	value, ok := data.(*client.FieldValue)
	if !ok {
		return nil, 0, NewErrUnexpectedValueType(mc.reg.CType(), &client.FieldValue{}, data)
	}
	delta, err := mc.reg.Increment(ctx, value.Value().(T))
	if err != nil {
		return nil, 0, err
	}
	nd, err := mc.clock.AddDAGNode(ctx, delta)
	return nd, delta.GetPriority(), err
}
