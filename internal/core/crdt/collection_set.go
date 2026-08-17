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

	"github.com/sourcenetwork/defradb/internal/datastore"
)

type CollectionSetDelta struct {
	Priority uint64
}

var _ Delta = (*CollectionSetDelta)(nil)

func (d *CollectionSetDelta) IPLDSchemaBytes() []byte {
	return []byte(`
	type CollectionSetDelta struct {
		priority  		Int
	}`)
}

func (d *CollectionSetDelta) GetPriority() uint64 {
	return d.Priority
}

type CollectionSetDefinition struct{}

var _ ReplicatedData = (*Collection)(nil)

func NewCollectionSet() *CollectionSetDefinition {
	return &CollectionSetDefinition{}
}

func (c *CollectionSetDefinition) Delta(
	priority uint64,
) *CollectionSetDelta {
	return &CollectionSetDelta{
		Priority: priority,
	}
}

func (c *CollectionSetDefinition) Merge(ctx context.Context, store datastore.Keyedstore, other Delta) error {
	return nil
}
