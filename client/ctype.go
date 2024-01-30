// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

// CType indicates CRDT type.
type CType byte

// Available CRDT types.
//
// If a CRDT type is not declared here, we do not support it!
// CRDT types here may not be valid in all contexts.
const (
	NONE_CRDT = CType(iota) // reserved none type
	LWW_REGISTER
	OBJECT
	COMPOSITE
	PN_COUNTER
)

// IsSupportedFieldCType returns true if the type is supported as a document field type.
func (t CType) IsSupportedFieldCType() bool {
	switch t {
	case NONE_CRDT, LWW_REGISTER, PN_COUNTER:
		return true
	default:
		return false
	}
}

// IsCompatibleWith returns true if the CRDT is compatible with the field kind
func (t CType) IsCompatibleWith(kind FieldKind) bool {
	switch t {
	case PN_COUNTER:
		if kind == FieldKind_NILLABLE_INT || kind == FieldKind_NILLABLE_FLOAT {
			return true
		}
		return false
	default:
		return true
	}
}

// String returns the string representation of the CRDT.
func (t CType) String() string {
	switch t {
	case NONE_CRDT:
		return "none"
	case LWW_REGISTER:
		return "lww"
	case OBJECT:
		return "object"
	case COMPOSITE:
		return "composite"
	case PN_COUNTER:
		return "pncounter"
	default:
		return "unknown"
	}
}
