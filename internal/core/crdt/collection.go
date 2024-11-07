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

	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// Collection is a simple CRDT type that tracks changes to the contents of a
// collection in a similar way to a document composite commit, only simpler,
// without the need to track status and a simpler [Merge] function.
type Collection struct {
	store datastore.DSReaderWriter

	// schemaVersionKey is the schema version datastore key at the time of commit.
	//
	// It can be used to identify the collection datastructure state at the time of commit.
	schemaVersionKey keys.CollectionSchemaVersionKey
}

var _ core.ReplicatedData = (*Collection)(nil)

func NewCollection(store datastore.DSReaderWriter, schemaVersionKey keys.CollectionSchemaVersionKey) *Collection {
	return &Collection{
		store:            store,
		schemaVersionKey: schemaVersionKey,
	}
}

func (c *Collection) Merge(ctx context.Context, other core.Delta) error {
	// Collection merges don't actually need to do anything, as the delta is empty,
	// and doc-level merges are handled by the document commits.
	return nil
}

func (c *Collection) Append() *CollectionDelta {
	return &CollectionDelta{
		SchemaVersionID: c.schemaVersionKey.SchemaVersionID,
	}
}

type CollectionDelta struct {
	Priority uint64

	// As we do not yet have a global collection id we temporarily rely on the schema
	// version id for tracking which collection this belongs to.  See:
	// https://github.com/sourcenetwork/defradb/issues/3215
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
