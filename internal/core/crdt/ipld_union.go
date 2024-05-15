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

import "github.com/sourcenetwork/defradb/internal/core"

// CRDT is a union type used for IPLD schemas that can hold any of the CRDT deltas.
type CRDT struct {
	LWWRegDelta       *LWWRegDelta
	CompositeDAGDelta *CompositeDAGDelta
	CounterDeltaInt   *CounterDelta[int64]
	CounterDeltaFloat *CounterDelta[float64]
}

// IPLDSchemaBytes returns the IPLD schema representation for the CRDT.
//
// This needs to match the [CRDT] struct or [mustSetSchema] will panic on init.
func (c CRDT) IPLDSchemaBytes() []byte {
	return []byte(`
	type CRDT union {
		| LWWRegDelta "lww"
		| CompositeDAGDelta "composite"
		| CounterDeltaInt "counterInt"
		| CounterDeltaFloat "counterFloat"
	} representation keyed`)
}

// GetDelta returns the delta that is stored in the CRDT.
func (c CRDT) GetDelta() core.Delta {
	switch {
	case c.LWWRegDelta != nil:
		return c.LWWRegDelta
	case c.CompositeDAGDelta != nil:
		return c.CompositeDAGDelta
	case c.CounterDeltaFloat != nil:
		return c.CounterDeltaFloat
	case c.CounterDeltaInt != nil:
		return c.CounterDeltaInt
	}
	return nil
}

// GetPriority returns the priority of the delta.
func (c CRDT) GetPriority() uint64 {
	switch {
	case c.LWWRegDelta != nil:
		return c.LWWRegDelta.GetPriority()
	case c.CompositeDAGDelta != nil:
		return c.CompositeDAGDelta.GetPriority()
	case c.CounterDeltaFloat != nil:
		return c.CounterDeltaFloat.GetPriority()
	case c.CounterDeltaInt != nil:
		return c.CounterDeltaInt.GetPriority()
	}
	return 0
}

// GetFieldName returns the field name of the delta.
func (c CRDT) GetFieldName() string {
	switch {
	case c.LWWRegDelta != nil:
		return c.LWWRegDelta.FieldName
	case c.CompositeDAGDelta != nil:
		return c.CompositeDAGDelta.FieldName
	case c.CounterDeltaFloat != nil:
		return c.CounterDeltaFloat.FieldName
	case c.CounterDeltaInt != nil:
		return c.CounterDeltaInt.FieldName
	}
	return ""
}

// GetDocID returns the docID of the delta.
func (c CRDT) GetDocID() []byte {
	switch {
	case c.LWWRegDelta != nil:
		return c.LWWRegDelta.DocID
	case c.CompositeDAGDelta != nil:
		return c.CompositeDAGDelta.DocID
	case c.CounterDeltaFloat != nil:
		return c.CounterDeltaFloat.DocID
	case c.CounterDeltaInt != nil:
		return c.CounterDeltaInt.DocID
	}
	return nil
}

// GetSchemaVersionID returns the schema version ID of the delta.
func (c CRDT) GetSchemaVersionID() string {
	switch {
	case c.LWWRegDelta != nil:
		return c.LWWRegDelta.SchemaVersionID
	case c.CompositeDAGDelta != nil:
		return c.CompositeDAGDelta.SchemaVersionID
	case c.CounterDeltaFloat != nil:
		return c.CounterDeltaFloat.SchemaVersionID
	case c.CounterDeltaInt != nil:
		return c.CounterDeltaInt.SchemaVersionID
	}
	return ""
}

// GetStatus returns the status of the delta.
//
// Currently only implemented for CompositeDAGDelta.
func (c CRDT) GetStatus() uint8 {
	if c.CompositeDAGDelta != nil {
		return uint8(c.CompositeDAGDelta.Status)
	}
	return 0
}

// GetData returns the data of the delta.
//
// Currently only implemented for LWWRegDelta.
func (c CRDT) GetData() []byte {
	if c.LWWRegDelta != nil {
		return c.LWWRegDelta.Data
	}
	return nil
}

// IsComposite returns true if the CRDT is a composite CRDT.
func (c CRDT) IsComposite() bool {
	return c.CompositeDAGDelta != nil
}
