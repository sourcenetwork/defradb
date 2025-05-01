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
)

// Collection is a simple CRDT type that tracks changes to the contents of a
// collection in a similar way to a document composite commit, only simpler,
// without the need to track status and a simpler [Merge] function.
type Collection struct {
	schemaVersionID string
}

var _ core.ReplicatedData = (*Collection)(nil)

func NewCollection(schemaVersionID string) *Collection {
	return &Collection{
		schemaVersionID: schemaVersionID,
	}
}

func (c *Collection) Merge(ctx context.Context, other core.Delta) error {
	// Collection merges don't actually need to do anything, as the delta is empty,
	// and doc-level merges are handled by the document commits.
	return nil
}

func (c *Collection) NewDelta() *CollectionDelta {
	return &CollectionDelta{
		SchemaVersionID: c.schemaVersionID,
	}
}

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
