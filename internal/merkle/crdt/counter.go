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

	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/merkle/clock"
)

// MerkleCounter is a MerkleCRDT implementation of the Counter using MerkleClocks.
type MerkleCounter struct {
	*baseMerkleCRDT

	reg crdt.Counter
}

// NewMerkleCounter creates a new instance (or loaded from DB) of a MerkleCRDT
// backed by a Counter CRDT.
func NewMerkleCounter(
	store Stores,
	schemaVersionKey core.CollectionSchemaVersionKey,
	key core.DataStoreKey,
	fieldName string,
	allowDecrement bool,
	kind client.ScalarKind,
) *MerkleCounter {
	register := crdt.NewCounter(store.Datastore(), schemaVersionKey, key, fieldName, allowDecrement, kind)
	clk := clock.NewMerkleClock(store.Headstore(), store.Blockstore(), key.ToHeadStoreKey(), register)
	base := &baseMerkleCRDT{clock: clk, crdt: register}
	return &MerkleCounter{
		baseMerkleCRDT: base,
		reg:            register,
	}
}

// Save the value of the  Counter to the DAG.
func (mc *MerkleCounter) Save(ctx context.Context, data any) (cidlink.Link, []byte, error) {
	value, ok := data.(*client.FieldValue)
	if !ok {
		return cidlink.Link{}, nil, NewErrUnexpectedValueType(mc.reg.CType(), &client.FieldValue{}, data)
	}
	bytes, err := value.Bytes()
	if err != nil {
		return cidlink.Link{}, nil, err
	}
	delta, err := mc.reg.Increment(ctx, bytes)
	if err != nil {
		return cidlink.Link{}, nil, err
	}
	return mc.clock.AddDelta(ctx, delta)
}
