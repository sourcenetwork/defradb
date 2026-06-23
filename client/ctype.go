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

import (
	"encoding/json"
)

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
	P_COUNTER
)

// String representations of the CRDT types, used by [CType.String] and [CTypeFromString].
const (
	noneCRDTString      = "none"
	lwwRegisterString   = "lww"
	objectCRDTString    = "object"
	compositeCRDTString = "composite"
	pnCounterString     = "pncounter"
	pCounterString      = "pcounter"
)

// IsSupportedFieldCType returns true if the type is supported as a document field type.
func (t CType) IsSupportedFieldCType() bool {
	switch t {
	case NONE_CRDT, LWW_REGISTER, PN_COUNTER, P_COUNTER:
		return true
	default:
		return false
	}
}

// IsCompatibleWith returns true if the CRDT is compatible with the field kind
func (t CType) IsCompatibleWith(kind FieldKind) bool {
	switch t {
	case PN_COUNTER, P_COUNTER:
		if kind == FieldKind_NILLABLE_INT || kind == FieldKind_NILLABLE_FLOAT64 || kind == FieldKind_NILLABLE_FLOAT32 {
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
		return noneCRDTString
	case LWW_REGISTER:
		return lwwRegisterString
	case OBJECT:
		return objectCRDTString
	case COMPOSITE:
		return compositeCRDTString
	case PN_COUNTER:
		return pnCounterString
	case P_COUNTER:
		return pCounterString
	default:
		return "unknown"
	}
}

// CTypeFromString returns the [CType] matching the given string representation,
// the inverse of [CType.String].
//
// The second return value is false if the string does not match a known CRDT type.
func CTypeFromString(s string) (CType, bool) {
	switch s {
	case noneCRDTString:
		return NONE_CRDT, true
	case lwwRegisterString:
		return LWW_REGISTER, true
	case objectCRDTString:
		return OBJECT, true
	case compositeCRDTString:
		return COMPOSITE, true
	case pnCounterString:
		return PN_COUNTER, true
	case pCounterString:
		return P_COUNTER, true
	default:
		return 0, false
	}
}

// UnmarshalJSON accepts both the numeric form (used when persisting and patching
// collections) and the string form (used by the collection-describe display output),
// so that a stringified CType round-trips back losslessly while the existing numeric
// reads keep working unchanged.
//
// Note: there is intentionally no MarshalJSON, persistence and JSON-patch evaluation
// must keep emitting the numeric form. Only the describe display path emits the string.
func (t *CType) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		ctype, ok := CTypeFromString(s)
		if !ok {
			return NewErrUnknownCRDTString(s)
		}
		*t = ctype
		return nil
	}

	var n byte
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	*t = CType(n)
	return nil
}
