// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

// ToArrayOfNormalValues converts a NormalValue into a slice of NormalValue if the given value
// is an array. If the given value is not an array, an error is returned.
func ToArrayOfNormalValues(val NormalValue) ([]NormalValue, error) {
	if !val.IsArray() {
		return nil, NewCanNotTurnNormalValueIntoArray(val)
	}
	if !val.IsNillable() {
		if v, ok := val.BoolArray(); ok {
			return toNormalArray(v, NewNormalBool), nil
		}
		if v, ok := val.IntArray(); ok {
			return toNormalArray(v, NewNormalInt), nil
		}
		if v, ok := val.FloatArray(); ok {
			return toNormalArray(v, NewNormalFloat), nil
		}
		if v, ok := val.StringArray(); ok {
			return toNormalArray(v, NewNormalString), nil
		}
		if v, ok := val.BytesArray(); ok {
			return toNormalArray(v, NewNormalBytes), nil
		}
		if v, ok := val.TimeArray(); ok {
			return toNormalArray(v, NewNormalTime), nil
		}
		if v, ok := val.DocumentArray(); ok {
			return toNormalArray(v, NewNormalDocument), nil
		}
		if v, ok := val.NillableBoolArray(); ok {
			return toNormalArray(v, NewNormalNillableBool), nil
		}
		if v, ok := val.NillableIntArray(); ok {
			return toNormalArray(v, NewNormalNillableInt), nil
		}
		if v, ok := val.NillableFloatArray(); ok {
			return toNormalArray(v, NewNormalNillableFloat), nil
		}
		if v, ok := val.NillableStringArray(); ok {
			return toNormalArray(v, NewNormalNillableString), nil
		}
		if v, ok := val.NillableBytesArray(); ok {
			return toNormalArray(v, NewNormalNillableBytes), nil
		}
		if v, ok := val.NillableTimeArray(); ok {
			return toNormalArray(v, NewNormalNillableTime), nil
		}
		if v, ok := val.NillableDocumentArray(); ok {
			return toNormalArray(v, NewNormalNillableDocument), nil
		}
	} else {
		if val.IsNil() {
			return nil, nil
		}
		if v, ok := val.NillableBoolNillableArray(); ok {
			return toNormalArray(v.Value(), NewNormalNillableBool), nil
		}
		if v, ok := val.NillableIntNillableArray(); ok {
			return toNormalArray(v.Value(), NewNormalNillableInt), nil
		}
		if v, ok := val.NillableFloatNillableArray(); ok {
			return toNormalArray(v.Value(), NewNormalNillableFloat), nil
		}
		if v, ok := val.NillableStringNillableArray(); ok {
			return toNormalArray(v.Value(), NewNormalNillableString), nil
		}
		if v, ok := val.NillableBytesNillableArray(); ok {
			return toNormalArray(v.Value(), NewNormalNillableBytes), nil
		}
		if v, ok := val.NillableTimeNillableArray(); ok {
			return toNormalArray(v.Value(), NewNormalNillableTime), nil
		}
		if v, ok := val.NillableDocumentNillableArray(); ok {
			return toNormalArray(v.Value(), NewNormalNillableDocument), nil
		}
		if v, ok := val.BoolNillableArray(); ok {
			return toNormalArray(v.Value(), NewNormalBool), nil
		}
		if v, ok := val.IntNillableArray(); ok {
			return toNormalArray(v.Value(), NewNormalInt), nil
		}
		if v, ok := val.FloatNillableArray(); ok {
			return toNormalArray(v.Value(), NewNormalFloat), nil
		}
		if v, ok := val.StringNillableArray(); ok {
			return toNormalArray(v.Value(), NewNormalString), nil
		}
		if v, ok := val.BytesNillableArray(); ok {
			return toNormalArray(v.Value(), NewNormalBytes), nil
		}
		if v, ok := val.TimeNillableArray(); ok {
			return toNormalArray(v.Value(), NewNormalTime), nil
		}
		if v, ok := val.DocumentNillableArray(); ok {
			return toNormalArray(v.Value(), NewNormalDocument), nil
		}
	}
	return nil, NewCanNotTurnNormalValueIntoArray(val)
}

func toNormalArray[T any](val []T, f func(T) NormalValue) []NormalValue {
	res := make([]NormalValue, len(val))
	for i := range val {
		res[i] = f(val[i])
	}
	return res
}
