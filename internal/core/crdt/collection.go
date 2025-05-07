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
