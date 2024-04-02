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
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/core/crdt"
	"github.com/sourcenetwork/defradb/merkle/clock"
)

// MerklePCounter is a MerkleCRDT implementation of the PCounter using MerkleClocks.
type MerklePCounter[T crdt.Incrementable] struct {
	*baseMerkleCRDT

	reg crdt.PCounter[T]
}

// NewMerklePCounter creates a new instance (or loaded from DB) of a MerkleCRDT
// backed by a PCounter CRDT.
func NewMerklePCounter[T crdt.Incrementable](
	store Stores,
	schemaVersionKey core.CollectionSchemaVersionKey,
	key core.DataStoreKey,
	fieldName string,
) *MerklePCounter[T] {
	register := crdt.NewPCounter[T](store.Datastore(), schemaVersionKey, key, fieldName)
	clk := clock.NewMerkleClock(store.Headstore(), store.DAGstore(), key.ToHeadStoreKey(), register)
	base := &baseMerkleCRDT{clock: clk, crdt: register}
	return &MerklePCounter[T]{
		baseMerkleCRDT: base,
		reg:            register,
	}
}

// Save the value of the PN Counter to the DAG.
func (mPNC *MerklePCounter[T]) Save(ctx context.Context, data any) (ipld.Node, uint64, error) {
	value, ok := data.(*client.FieldValue)
	if !ok {
		return nil, 0, NewErrUnexpectedValueType(client.PN_COUNTER, &client.FieldValue{}, data)
	}
	delta, err := mPNC.reg.Increment(ctx, value.Value().(T))
	if err != nil {
		return nil, 0, err
	}
	nd, err := mPNC.clock.AddDAGNode(ctx, delta)
	return nd, delta.GetPriority(), err
}
