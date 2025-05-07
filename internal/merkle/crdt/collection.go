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

	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/keys"
)

type MerkleCollection struct {
	headstorePrefix keys.HeadstoreKey
	schemaVersionID string
}

var _ MerkleCRDT = (*MerkleCollection)(nil)
var _ core.ReplicatedData = (*MerkleCollection)(nil)

func NewMerkleCollection(
	schemaVersionID string,
	key keys.HeadstoreColKey,
) *MerkleCollection {
	return &MerkleCollection{
		schemaVersionID: schemaVersionID,
		headstorePrefix: key,
	}
}

func (m *MerkleCollection) HeadstorePrefix() keys.HeadstoreKey {
	return m.headstorePrefix
}

func (m *MerkleCollection) Delta() *crdt.CollectionDelta {
	return &crdt.CollectionDelta{
		SchemaVersionID: m.schemaVersionID,
	}
}

func (c *MerkleCollection) Merge(ctx context.Context, other core.Delta) error {
	// Collection merges don't actually need to do anything, as the delta is empty,
	// and doc-level merges are handled by the document commits.
	return nil
}
