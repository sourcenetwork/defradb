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
	"github.com/fxamacker/cbor/v2"
	"github.com/sourcenetwork/immutable"
)

type FieldValue struct {
	t       CType
	value   any
	kind    FieldKind
	isDirty bool
}

func NewFieldValue(t CType, val any, kind FieldKind) *FieldValue {
	return &FieldValue{
		t:       t,
		value:   val,
		kind:    kind,
		isDirty: true,
	}
}

func (val FieldValue) Value() any {
	return val.value
}

func (val FieldValue) Type() CType {
	return val.t
}

func (val FieldValue) IsDocument() bool {
	_, ok := val.value.(*Document)
	return ok
}

// IsDirty returns if the value is marked as dirty (unsaved/changed)
func (val FieldValue) IsDirty() bool {
	return val.isDirty
}

func (val *FieldValue) Clean() {
	val.isDirty = false
}

func (val *FieldValue) SetType(t CType) {
	val.t = t
}

func (val FieldValue) Bytes() ([]byte, error) {
	em, err := cbor.EncOptions{Time: cbor.TimeRFC3339}.EncMode()
	if err != nil {
		return nil, err
	}

	var value any
	switch tempVal := val.value.(type) {
	case []immutable.Option[string]:
		value = convertImmutable(tempVal)
	case []immutable.Option[int64]:
		value = convertImmutable(tempVal)
	case []immutable.Option[float64]:
		value = convertImmutable(tempVal)
	case []immutable.Option[bool]:
		value = convertImmutable(tempVal)
	default:
		value = val.value
	}

	return em.Marshal(value)
}

func convertImmutable[T any](vals []immutable.Option[T]) []any {
	var out []any
	for _, val := range vals {
		if !val.HasValue() {
			out = append(out, nil)
			continue
		}
		out = append(out, val.Value())
	}
	return out
}

func (val *FieldValue) Kind() FieldKind {
	return val.kind
}

func valOrNil[T any](value any) (*T, bool) {
	if v, ok := value.(T); ok {
		return &v, true
	}
	if v, ok := value.(*T); ok {
		return v, true
	}

	return nil, false
}

func (val *FieldValue) IsNil() bool {
	return val.value == nil
}

func (val *FieldValue) boolOrNil() (*bool, bool) {
	return valOrNil[bool](val.value)
}

func (val *FieldValue) intOrNil() (*int64, bool) {
	valInt64, ok := valOrNil[int64](val.value)
	if ok {
		return valInt64, true
	}
	valInt32, ok := valOrNil[int32](val.value)
	if ok {
		v := int64(*valInt32)
		return &v, true
	}
	valInt, ok := valOrNil[int](val.value)
	if ok {
		v := int64(*valInt)
		return &v, true
	}
	return nil, false
}

func (val *FieldValue) floatOrNil() (*float64, bool) {
	return valOrNil[float64](val.value)
}

func (val *FieldValue) stringOrNil() (*string, bool) {
	return valOrNil[string](val.value)
}

func (val *FieldValue) Bool() (bool, error) {
	v, ok := val.boolOrNil()
	if !ok || v == nil {
		return false, NewErrUnexpectedType[bool]("", val.value)
	}
	return *v, nil
}

func (val *FieldValue) Int() (int64, error) {
	v, ok := val.intOrNil()
	if !ok || v == nil {
		return 0, NewErrUnexpectedType[int64]("", val.value)
	}
	return *v, nil
}

func (val *FieldValue) Float() (float64, error) {
	v, ok := val.floatOrNil()
	if !ok || v == nil {
		return 0, NewErrUnexpectedType[float64]("", val.value)
	}
	return *v, nil
}

func (val *FieldValue) String() (string, error) {
	v, ok := val.stringOrNil()
	if !ok || v == nil {
		return "", NewErrUnexpectedType[string]("", val.value)
	}
	return *v, nil
}
