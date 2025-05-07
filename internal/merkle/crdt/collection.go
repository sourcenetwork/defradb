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

	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/merkle/clock"
)

type MerkleCollection struct {
	clock           *clock.MerkleClock
	schemaVersionID string
}

var _ MerkleCRDT = (*MerkleCollection)(nil)
var _ core.ReplicatedData = (*MerkleCollection)(nil)

func NewMerkleCollection(
	store Stores,
	schemaVersionID string,
	key keys.HeadstoreColKey,
) *MerkleCollection {
	dag := &MerkleCollection{
		schemaVersionID: schemaVersionID,
	}

	dag.clock = clock.NewMerkleClock(store.Headstore(), store.Blockstore(), store.Encstore(), key, dag)

	return dag
}

func (m *MerkleCollection) Clock() *clock.MerkleClock {
	return m.clock
}

func (m *MerkleCollection) Save(ctx context.Context, links []coreblock.DAGLink) (cidlink.Link, []byte, error) {
	return m.clock.AddDelta(
		ctx,
		&crdt.CollectionDelta{
			SchemaVersionID: m.schemaVersionID,
		},
		links...,
	)
}

func (c *MerkleCollection) Merge(ctx context.Context, other core.Delta) error {
	// Collection merges don't actually need to do anything, as the delta is empty,
	// and doc-level merges are handled by the document commits.
	return nil
}
