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
	"github.com/sourcenetwork/defradb/internal/core"
)

// LWWRegDelta is a single delta operation for an LWWRegister
// @todo: Expand delta metadata (investigate if needed)
type LWWRegDelta struct {
	DocID     []byte
	FieldName string
	Priority  uint64
	// SchemaVersionID is the schema version datastore key at the time of commit.
	//
	// It can be used to identify the collection datastructure state at the time of commit.
	SchemaVersionID string
	Data            []byte
}

var _ core.Delta = (*LWWRegDelta)(nil)

// IPLDSchemaBytes returns the IPLD schema representation for the type.
//
// This needs to match the [LWWRegDelta] struct or [coreblock.mustSetSchema] will panic on init.
func (delta LWWRegDelta) IPLDSchemaBytes() []byte {
	return []byte(`
	type LWWRegDelta struct {
		docID     		Bytes
		fieldName 		String
		priority  		Int
		schemaVersionID String
		data            Bytes
	}`)
}

// GetPriority gets the current priority for this delta.
func (delta *LWWRegDelta) GetPriority() uint64 {
	return delta.Priority
}

// SetPriority will set the priority for this delta.
func (delta *LWWRegDelta) SetPriority(prio uint64) {
	delta.Priority = prio
}
