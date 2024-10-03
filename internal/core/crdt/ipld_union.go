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
	CounterDelta      *CounterDelta
}

// NewCRDT returns a new CRDT.
func NewCRDT(delta core.Delta) CRDT {
	switch d := delta.(type) {
	case *LWWRegDelta:
		return CRDT{LWWRegDelta: d}
	case *CompositeDAGDelta:
		return CRDT{CompositeDAGDelta: d}
	case *CounterDelta:
		return CRDT{CounterDelta: d}
	}
	return CRDT{}
}

// IPLDSchemaBytes returns the IPLD schema representation for the CRDT.
//
// This needs to match the [CRDT] struct or [mustSetSchema] will panic on init.
func (c CRDT) IPLDSchemaBytes() []byte {
	return []byte(`
	type CRDT union {
		| LWWRegDelta "lww"
		| CompositeDAGDelta "composite"
		| CounterDelta "counter"
	} representation keyed`)
}

// GetDelta returns the delta that is stored in the CRDT.
func (c CRDT) GetDelta() core.Delta {
	switch {
	case c.LWWRegDelta != nil:
		return c.LWWRegDelta
	case c.CompositeDAGDelta != nil:
		return c.CompositeDAGDelta
	case c.CounterDelta != nil:
		return c.CounterDelta
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
	case c.CounterDelta != nil:
		return c.CounterDelta.GetPriority()
	}
	return 0
}

// GetFieldName returns the field name of the delta.
func (c CRDT) GetFieldName() string {
	switch {
	case c.LWWRegDelta != nil:
		return c.LWWRegDelta.FieldName
	case c.CounterDelta != nil:
		return c.CounterDelta.FieldName
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
	case c.CounterDelta != nil:
		return c.CounterDelta.DocID
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
	case c.CounterDelta != nil:
		return c.CounterDelta.SchemaVersionID
	}
	return ""
}

// Clone returns a clone of the CRDT.
func (c CRDT) Clone() CRDT {
	var cloned CRDT
	switch {
	case c.LWWRegDelta != nil:
		cloned.LWWRegDelta = &LWWRegDelta{
			DocID:           c.LWWRegDelta.DocID,
			FieldName:       c.LWWRegDelta.FieldName,
			Priority:        c.LWWRegDelta.Priority,
			SchemaVersionID: c.LWWRegDelta.SchemaVersionID,
			Data:            c.LWWRegDelta.Data,
		}
	case c.CompositeDAGDelta != nil:
		cloned.CompositeDAGDelta = &CompositeDAGDelta{
			DocID:           c.CompositeDAGDelta.DocID,
			Priority:        c.CompositeDAGDelta.Priority,
			SchemaVersionID: c.CompositeDAGDelta.SchemaVersionID,
			Status:          c.CompositeDAGDelta.Status,
		}
	case c.CounterDelta != nil:
		cloned.CounterDelta = &CounterDelta{
			DocID:           c.CounterDelta.DocID,
			FieldName:       c.CounterDelta.FieldName,
			Priority:        c.CounterDelta.Priority,
			SchemaVersionID: c.CounterDelta.SchemaVersionID,
			Nonce:           c.CounterDelta.Nonce,
			Data:            c.CounterDelta.Data,
		}
	}
	return cloned
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
func (c CRDT) GetData() []byte {
	if c.LWWRegDelta != nil {
		return c.LWWRegDelta.Data
	} else if c.CounterDelta != nil {
		return c.CounterDelta.Data
	}
	return nil
}

// SetData sets the data of the delta.
func (c CRDT) SetData(data []byte) {
	if c.LWWRegDelta != nil {
		c.LWWRegDelta.Data = data
	} else if c.CounterDelta != nil {
		c.CounterDelta.Data = data
	}
}

// IsComposite returns true if the CRDT is a composite CRDT.
func (c CRDT) IsComposite() bool {
	return c.CompositeDAGDelta != nil
}
