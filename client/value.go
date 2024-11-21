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
	"github.com/sourcenetwork/immutable"
)

type FieldValue struct {
	t       CType
	value   NormalValue
	isDirty bool
}

func NewFieldValue(t CType, val NormalValue) *FieldValue {
	return &FieldValue{
		t:       t,
		value:   val,
		isDirty: true,
	}
}

func (val FieldValue) Value() any {
	jsonVal, ok := val.value.JSON()
	if ok {
		return jsonVal.Unwrap()
	}
	return val.value.Unwrap()
}

func (val FieldValue) NormalValue() NormalValue {
	return val.value
}

func (val FieldValue) Type() CType {
	return val.t
}

func (val FieldValue) IsDocument() bool {
	_, ok := val.value.Document()
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
	em, err := CborEncodingOptions().EncMode()
	if err != nil {
		return nil, err
	}

	var value any
	if v, ok := val.value.NillableStringArray(); ok {
		value = convertImmutable(v)
	} else if v, ok := val.value.NillableIntArray(); ok {
		value = convertImmutable(v)
	} else if v, ok := val.value.NillableFloatArray(); ok {
		value = convertImmutable(v)
	} else if v, ok := val.value.NillableBoolArray(); ok {
		value = convertImmutable(v)
	} else {
		value = val.value.Unwrap()
	}

	return em.Marshal(value)
}

func convertImmutable[T any](vals []immutable.Option[T]) []any {
	out := make([]any, len(vals))
	for i := range vals {
		if vals[i].HasValue() {
			out[i] = vals[i].Value()
		}
	}
	return out
}
