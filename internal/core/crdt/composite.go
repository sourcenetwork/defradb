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
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/core"
)

// CompositeDAGDelta represents a delta-state update made of sub-MerkleCRDTs.
type CompositeDAGDelta struct {
	// This property is duplicated from field-level blocks.
	//
	// We could remove this without much hassle from the composite, however long-term
	// the ideal solution would be to remove it from the field-level commits *excluding*
	// the initial field level commit where it must exist in order to scope it to a particular
	// document.  This would require a local index in order to handle field level commit-queries.
	DocID    []byte
	Priority uint64
	// SchemaVersionID is the schema version datastore key at the time of commit.
	//
	// It can be used to identify the collection datastructure state at the time of commit.
	//
	// This property is deliberately duplicated from field-level blocks as it makes the P2P code
	// quite a lot easier - we can remove this from here at some point if we want to.
	//
	// Conversely we could remove this from the field-level commits and leave it on the composite,
	// however that would complicate commit-queries and would require us to maintain an index elsewhere.
	SchemaVersionID string
	// Status represents the status of the document. By default it is `Active`.
	// Alternatively, if can be set to `Deleted`.
	Status client.DocumentStatus
}

var _ core.Delta = (*CompositeDAGDelta)(nil)

// IPLDSchemaBytes returns the IPLD schema representation for the type.
//
// This needs to match the [CompositeDAGDelta] struct or [coreblock.mustSetSchema] will panic on init.
func (delta *CompositeDAGDelta) IPLDSchemaBytes() []byte {
	return []byte(`
	type CompositeDAGDelta struct {
		docID     		Bytes
		priority  		Int
		schemaVersionID String
		status          Int
	}`)
}

// GetPriority gets the current priority for this delta.
func (delta *CompositeDAGDelta) GetPriority() uint64 {
	return delta.Priority
}

// SetPriority will set the priority for this delta.
func (delta *CompositeDAGDelta) SetPriority(prio uint64) {
	delta.Priority = prio
}
