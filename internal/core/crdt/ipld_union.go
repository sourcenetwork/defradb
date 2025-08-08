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
	LWWDelta                  *LWWDelta
	DocCompositeDelta         *DocCompositeDelta
	CounterDelta              *CounterDelta
	CollectionDelta           *CollectionDelta
	CollectionSetDelta        *CollectionSetDelta
	CollectionDefinitionDelta *CollectionDefinitionDelta
	FieldDefinitionDelta      *FieldDefinitionDelta
}

// NewCRDT returns a new CRDT.
func NewCRDT(delta core.Delta) CRDT {
	switch d := delta.(type) {
	case *LWWDelta:
		return CRDT{LWWDelta: d}
	case *DocCompositeDelta:
		return CRDT{DocCompositeDelta: d}
	case *CounterDelta:
		return CRDT{CounterDelta: d}
	case *CollectionSetDelta:
		return CRDT{CollectionSetDelta: d}
	case *CollectionDelta:
		return CRDT{CollectionDelta: d}
	case *CollectionDefinitionDelta:
		return CRDT{CollectionDefinitionDelta: d}
	case *FieldDefinitionDelta:
		return CRDT{FieldDefinitionDelta: d}
	}
	return CRDT{}
}

// IPLDSchemaBytes returns the IPLD schema representation for the CRDT.
//
// This needs to match the [CRDT] struct or [mustSetSchema] will panic on init.
func (c CRDT) IPLDSchemaBytes() []byte {
	return []byte(`
	type CRDT union {
		| LWWDelta "lww"
		| DocCompositeDelta "composite"
		| CounterDelta "counter"
		| CollectionDelta "collection"
		| CollectionSetDelta "collectionSet"
		| CollectionDefinitionDelta "collectionDefinition"
		| FieldDefinitionDelta "fieldDefinition"
	} representation keyed`)
}

// GetDelta returns the delta that is stored in the CRDT.
func (c CRDT) GetDelta() core.Delta {
	switch {
	case c.LWWDelta != nil:
		return c.LWWDelta
	case c.DocCompositeDelta != nil:
		return c.DocCompositeDelta
	case c.CounterDelta != nil:
		return c.CounterDelta
	case c.CollectionDelta != nil:
		return c.CollectionDelta
	case c.CollectionDelta != nil:
		return c.CollectionSetDelta
	case c.CollectionDefinitionDelta != nil:
		return c.CollectionDefinitionDelta
	case c.FieldDefinitionDelta != nil:
		return c.FieldDefinitionDelta
	}
	return nil
}

// GetPriority returns the priority of the delta.
func (c CRDT) GetPriority() uint64 {
	switch {
	case c.LWWDelta != nil:
		return c.LWWDelta.GetPriority()
	case c.DocCompositeDelta != nil:
		return c.DocCompositeDelta.GetPriority()
	case c.CounterDelta != nil:
		return c.CounterDelta.GetPriority()
	case c.CollectionDelta != nil:
		return c.CollectionDelta.GetPriority()
	case c.CollectionSetDelta != nil:
		return c.CollectionSetDelta.GetPriority()
	case c.CollectionDefinitionDelta != nil:
		return c.CollectionDefinitionDelta.GetPriority()
	case c.FieldDefinitionDelta != nil:
		return c.FieldDefinitionDelta.GetPriority()
	}
	return 0
}

// GetFieldName returns the field name of the delta.
func (c CRDT) GetFieldName() string {
	switch {
	case c.LWWDelta != nil:
		return c.LWWDelta.FieldName
	case c.CounterDelta != nil:
		return c.CounterDelta.FieldName
	}
	return ""
}

// GetDocID returns the docID of the delta.
func (c CRDT) GetDocID() []byte {
	switch {
	case c.LWWDelta != nil:
		return c.LWWDelta.DocID
	case c.DocCompositeDelta != nil:
		return c.DocCompositeDelta.DocID
	case c.CounterDelta != nil:
		return c.CounterDelta.DocID
	case c.CollectionDelta != nil:
		return nil
	}
	return nil
}

// GetSchemaVersionID returns the schema version ID of the delta.
func (c CRDT) GetSchemaVersionID() string {
	switch {
	case c.LWWDelta != nil:
		return c.LWWDelta.SchemaVersionID
	case c.DocCompositeDelta != nil:
		return c.DocCompositeDelta.SchemaVersionID
	case c.CounterDelta != nil:
		return c.CounterDelta.SchemaVersionID
	case c.CollectionDelta != nil:
		return c.CollectionDelta.SchemaVersionID
	}
	return ""
}

// Clone returns a clone of the CRDT.
func (c CRDT) Clone() CRDT {
	var cloned CRDT
	switch {
	case c.LWWDelta != nil:
		cloned.LWWDelta = &LWWDelta{
			DocID:           c.LWWDelta.DocID,
			FieldName:       c.LWWDelta.FieldName,
			Priority:        c.LWWDelta.Priority,
			SchemaVersionID: c.LWWDelta.SchemaVersionID,
			Data:            c.LWWDelta.Data,
		}
	case c.DocCompositeDelta != nil:
		cloned.DocCompositeDelta = &DocCompositeDelta{
			DocID:           c.DocCompositeDelta.DocID,
			Priority:        c.DocCompositeDelta.Priority,
			SchemaVersionID: c.DocCompositeDelta.SchemaVersionID,
			Status:          c.DocCompositeDelta.Status,
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
	case c.CollectionDelta != nil:
		cloned.CollectionDelta = &CollectionDelta{
			Priority:        c.CollectionDelta.Priority,
			SchemaVersionID: c.CollectionDelta.SchemaVersionID,
		}
	}
	return cloned
}

// GetStatus returns the status of the delta.
//
// Currently only implemented for CompositeDAGDelta.
func (c CRDT) GetStatus() uint8 {
	if c.DocCompositeDelta != nil {
		return uint8(c.DocCompositeDelta.Status)
	}
	return 0
}

// GetData returns the data of the delta.
func (c CRDT) GetData() []byte {
	if c.LWWDelta != nil {
		return c.LWWDelta.Data
	} else if c.CounterDelta != nil {
		return c.CounterDelta.Data
	}
	return nil
}

// SetData sets the data of the delta.
func (c CRDT) SetData(data []byte) {
	if c.LWWDelta != nil {
		c.LWWDelta.Data = data
	} else if c.CounterDelta != nil {
		c.CounterDelta.Data = data
	}
}

// IsComposite returns true if the CRDT is a composite CRDT.
func (c CRDT) IsComposite() bool {
	return c.DocCompositeDelta != nil
}

// IsCollection returns true if the CRDT is a collection CRDT.
func (c CRDT) IsCollection() bool {
	return c.CollectionDelta != nil
}

// IsField returns true if the CRDT is a field CRDT.
func (c CRDT) IsField() bool {
	return !c.IsComposite() && !c.IsCollection()
}
