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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/keys"
)

type CollectionDefinitionDelta struct {
	Priority uint64

	Name string
}

var _ core.Delta = (*CollectionDefinitionDelta)(nil)

func (delta *CollectionDefinitionDelta) IPLDSchemaBytes() []byte {
	return []byte(`
	type CollectionDefinitionDelta struct {
		priority  		Int
		name String
	}`)
}

func (d *CollectionDefinitionDelta) GetPriority() uint64 {
	return d.Priority
}

func (d *CollectionDefinitionDelta) SetPriority(priority uint64) {
	d.Priority = priority
}

type CollectionDefinition struct {
	headstorePrefix keys.HeadstoreCollectionDefinition
}

var _ core.ReplicatedData = (*Collection)(nil)

func NewCollectionDefinition(
	name string,
) *CollectionDefinition {
	return &CollectionDefinition{
		// WARNING: This prefix will need to be rebuilt if/when we allow the mutation of collection
		// name.
		headstorePrefix: keys.HeadstoreCollectionDefinition{
			CollectionName: name,
		},
	}
}

func (m *CollectionDefinition) HeadstorePrefix() keys.HeadstoreKey {
	return m.headstorePrefix
}

func (m *CollectionDefinition) Delta(
	new client.CollectionVersion,
	old client.CollectionVersion,
) (*CollectionDefinitionDelta, bool) {
	if new.Name == old.Name {
		return &CollectionDefinitionDelta{}, false
	}

	return &CollectionDefinitionDelta{
		Name: new.Name,
	}, true
}

func (c *CollectionDefinition) Merge(ctx context.Context, other core.Delta) error {
	return nil
}
