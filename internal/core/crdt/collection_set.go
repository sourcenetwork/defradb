// Copyright 2025 Democratized Data Foundation
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

type CollectionSetDelta struct {
	Priority uint64
}

var _ core.Delta = (*CollectionSetDelta)(nil)

func (delta *CollectionSetDelta) IPLDSchemaBytes() []byte {
	return []byte(`
	type CollectionSetDelta struct {
		priority  		Int
	}`)
}

func (d *CollectionSetDelta) GetPriority() uint64 {
	return d.Priority
}

func (d *CollectionSetDelta) SetPriority(priority uint64) {
	d.Priority = priority
}

type CollectionSetDefinition struct {
	headstorePrefix keys.HeadstoreCollectionSetDefinition
}

var _ core.ReplicatedData = (*Collection)(nil)

func NewCollectionSet(
	firstCollectionID string,
) *CollectionSetDefinition {
	return &CollectionSetDefinition{
		headstorePrefix: keys.HeadstoreCollectionSetDefinition{
			FirstCollectionID: firstCollectionID,
		},
	}
}

func (m *CollectionSetDefinition) HeadstorePrefix() keys.HeadstoreKey {
	return m.headstorePrefix
}

func (m *CollectionSetDefinition) Delta() *CollectionSetDelta {
	return &CollectionSetDelta{}
}

func (c *CollectionSetDefinition) Merge(ctx context.Context, other core.Delta) error {
	return nil
}
