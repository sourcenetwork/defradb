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
	"github.com/sourcenetwork/defradb/core/crdt"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/merkle/clock"
)

// MerklePNCounter is a MerkleCRDT implementation of the PNCounter using MerkleClocks.
type MerklePNCounter[T crdt.Incrementable] struct {
	*baseMerkleCRDT

	reg crdt.PNCounter[T]
}

// NewMerklePNCounter creates a new instance (or loaded from DB) of a MerkleCRDT
// backed by a PNCounter CRDT.
func NewMerklePNCounter[T crdt.Incrementable](
	store Stores,
	schemaVersionKey core.CollectionSchemaVersionKey,
	key core.DataStoreKey,
	fieldName string,
) *MerklePNCounter[T] {
	register := crdt.NewPNCounter[T](store.Datastore(), schemaVersionKey, key, fieldName)
	clk := clock.NewMerkleClock(store.Headstore(), store.DAGstore(), key.ToHeadStoreKey(), register)
	base := &baseMerkleCRDT{clock: clk, crdt: register}
	return &MerklePNCounter[T]{
		baseMerkleCRDT: base,
		reg:            register,
	}
}

// Increment the value of the register.
func (mPNC *MerklePNCounter[T]) Save(ctx context.Context, data any) (ipld.Node, uint64, error) {
	value, ok := data.(*client.FieldValue)
	if !ok {
		return nil, 0, errors.New("invalid type")
	}
	delta := mPNC.reg.Increment(value.Value().(T))
	nd, err := mPNC.clock.AddDAGNode(ctx, delta)
	return nd, delta.GetPriority(), err
}
