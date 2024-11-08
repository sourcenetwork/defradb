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

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/merkle/clock"
)

type MerkleCollection struct {
	clock *clock.MerkleClock
	reg   *crdt.Collection
}

var _ MerkleCRDT = (*MerkleCollection)(nil)

func NewMerkleCollection(
	store Stores,
	schemaVersionKey keys.CollectionSchemaVersionKey,
	key keys.HeadstoreColKey,
) *MerkleCollection {
	register := crdt.NewCollection(schemaVersionKey)

	clk := clock.NewMerkleClock(store.Headstore(), store.Blockstore(), store.Encstore(), key, register)

	return &MerkleCollection{
		clock: clk,
		reg:   register,
	}
}

func (m *MerkleCollection) Clock() *clock.MerkleClock {
	return m.clock
}

func (m *MerkleCollection) Save(ctx context.Context, links []coreblock.DAGLink) (cidlink.Link, []byte, error) {
	delta := m.reg.Append()
	return m.clock.AddDelta(ctx, delta, links...)
}
