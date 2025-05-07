// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package crdt

import (
	"context"

	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/keys"
)

type CollectionDelta struct {
	Priority        uint64
	SchemaVersionID string
}

var _ core.Delta = (*CollectionDelta)(nil)

func (delta *CollectionDelta) IPLDSchemaBytes() []byte {
	return []byte(`
	type CollectionDelta struct {
		priority  		Int
		schemaVersionID String
	}`)
}

func (d *CollectionDelta) GetPriority() uint64 {
	return d.Priority
}

func (d *CollectionDelta) SetPriority(priority uint64) {
	d.Priority = priority
}

type MerkleCollection struct {
	headstorePrefix keys.HeadstoreKey
	schemaVersionID string
}

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

func (m *MerkleCollection) Delta() *CollectionDelta {
	return &CollectionDelta{
		SchemaVersionID: m.schemaVersionID,
	}
}

func (c *MerkleCollection) Merge(ctx context.Context, other core.Delta) error {
	// Collection merges don't actually need to do anything, as the delta is empty,
	// and doc-level merges are handled by the document commits.
	return nil
}
