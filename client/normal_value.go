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

import (
	"time"

	"github.com/sourcenetwork/immutable"
	"golang.org/x/exp/constraints"
)

type NormalValue interface {
	Any() any

	IsNil() bool
	IsNillable() bool

	Bool() (bool, bool)
	Int() (int64, bool)
	Float() (float64, bool)
	String() (string, bool)
	Bytes() ([]byte, bool)
	Time() (time.Time, bool)
	Document() (*Document, bool)

	NillableBool() (immutable.Option[bool], bool)
	NillableInt() (immutable.Option[int64], bool)
	NillableFloat() (immutable.Option[float64], bool)
	NillableString() (immutable.Option[string], bool)
	NillableBytes() (immutable.Option[[]byte], bool)
	NillableTime() (immutable.Option[time.Time], bool)
	NillableDocument() (immutable.Option[*Document], bool)

	IsArray() bool

	BoolArray() ([]bool, bool)
	IntArray() ([]int64, bool)
	FloatArray() ([]float64, bool)
	StringArray() ([]string, bool)
	BytesArray() ([][]byte, bool)
	TimeArray() ([]time.Time, bool)
	DocumentArray() ([]*Document, bool)

	BoolNillableArray() (immutable.Option[[]bool], bool)
	IntNillableArray() (immutable.Option[[]int64], bool)
	FloatNillableArray() (immutable.Option[[]float64], bool)
	StringNillableArray() (immutable.Option[[]string], bool)
	BytesNillableArray() (immutable.Option[[][]byte], bool)
	TimeNillableArray() (immutable.Option[[]time.Time], bool)
	DocumentNillableArray() (immutable.Option[[]*Document], bool)

	NillableBoolArray() ([]immutable.Option[bool], bool)
	NillableIntArray() ([]immutable.Option[int64], bool)
	NillableFloatArray() ([]immutable.Option[float64], bool)
	NillableStringArray() ([]immutable.Option[string], bool)
	NillableBytesArray() ([]immutable.Option[[]byte], bool)
	NillableTimeArray() ([]immutable.Option[time.Time], bool)
	NillableDocumentArray() ([]immutable.Option[*Document], bool)

	NillableBoolNillableArray() (immutable.Option[[]immutable.Option[bool]], bool)
	NillableIntNillableArray() (immutable.Option[[]immutable.Option[int64]], bool)
	NillableFloatNillableArray() (immutable.Option[[]immutable.Option[float64]], bool)
	NillableStringNillableArray() (immutable.Option[[]immutable.Option[string]], bool)
	NillableBytesNillableArray() (immutable.Option[[]immutable.Option[[]byte]], bool)
	NillableTimeNillableArray() (immutable.Option[[]immutable.Option[time.Time]], bool)
	NillableDocumentNillableArray() (immutable.Option[[]immutable.Option[*Document]], bool)
}

type NormalVoid struct{}

func (NormalVoid) IsNil() bool {
	return false
}

func (NormalVoid) IsNillable() bool {
	return false
}

func (NormalVoid) Bool() (bool, bool) {
	return false, false
}

func (NormalVoid) Int() (int64, bool) {
	return 0, false
}

func (NormalVoid) Float() (float64, bool) {
	return 0, false
}

func (NormalVoid) String() (string, bool) {
	return "", false
}

func (NormalVoid) Bytes() ([]byte, bool) {
	return nil, false
}

func (NormalVoid) Time() (time.Time, bool) {
	return time.Time{}, false
}

func (NormalVoid) Document() (*Document, bool) {
	return nil, false
}

func (NormalVoid) NillableBool() (immutable.Option[bool], bool) {
	return immutable.None[bool](), false
}

func (NormalVoid) NillableInt() (immutable.Option[int64], bool) {
	return immutable.None[int64](), false
}

func (NormalVoid) NillableFloat() (immutable.Option[float64], bool) {
	return immutable.None[float64](), false
}

func (NormalVoid) NillableString() (immutable.Option[string], bool) {
	return immutable.None[string](), false
}

func (NormalVoid) NillableBytes() (immutable.Option[[]byte], bool) {
	return immutable.None[[]byte](), false
}

func (NormalVoid) NillableTime() (immutable.Option[time.Time], bool) {
	return immutable.None[time.Time](), false
}

func (NormalVoid) NillableDocument() (immutable.Option[*Document], bool) {
	return immutable.None[*Document](), false
}

func (NormalVoid) IsArray() bool {
	return false
}

func (NormalVoid) BoolArray() ([]bool, bool) {
	return nil, false
}

func (NormalVoid) IntArray() ([]int64, bool) {
	return nil, false
}

func (NormalVoid) FloatArray() ([]float64, bool) {
	return nil, false
}

func (NormalVoid) StringArray() ([]string, bool) {
	return nil, false
}

func (NormalVoid) BytesArray() ([][]byte, bool) {
	return nil, false
}

func (NormalVoid) TimeArray() ([]time.Time, bool) {
	return nil, false
}

func (NormalVoid) DocumentArray() ([]*Document, bool) {
	return nil, false
}

func (NormalVoid) NillableBoolArray() ([]immutable.Option[bool], bool) {
	return nil, false
}

func (NormalVoid) NillableIntArray() ([]immutable.Option[int64], bool) {
	return nil, false
}

func (NormalVoid) NillableFloatArray() ([]immutable.Option[float64], bool) {
	return nil, false
}

func (NormalVoid) NillableStringArray() ([]immutable.Option[string], bool) {
	return nil, false
}

func (NormalVoid) NillableBytesArray() ([]immutable.Option[[]byte], bool) {
	return nil, false
}

func (NormalVoid) NillableTimeArray() ([]immutable.Option[time.Time], bool) {
	return nil, false
}

func (NormalVoid) NillableDocumentArray() ([]immutable.Option[*Document], bool) {
	return nil, false
}

func (NormalVoid) BoolNillableArray() (immutable.Option[[]bool], bool) {
	return immutable.None[[]bool](), false
}

func (NormalVoid) IntNillableArray() (immutable.Option[[]int64], bool) {
	return immutable.None[[]int64](), false
}

func (NormalVoid) FloatNillableArray() (immutable.Option[[]float64], bool) {
	return immutable.None[[]float64](), false
}

func (NormalVoid) StringNillableArray() (immutable.Option[[]string], bool) {
	return immutable.None[[]string](), false
}

func (NormalVoid) BytesNillableArray() (immutable.Option[[][]byte], bool) {
	return immutable.None[[][]byte](), false
}

func (NormalVoid) TimeNillableArray() (immutable.Option[[]time.Time], bool) {
	return immutable.None[[]time.Time](), false
}

func (NormalVoid) DocumentNillableArray() (immutable.Option[[]*Document], bool) {
	return immutable.None[[]*Document](), false
}

func (NormalVoid) NillableBoolNillableArray() (immutable.Option[[]immutable.Option[bool]], bool) {
	return immutable.None[[]immutable.Option[bool]](), false
}

func (NormalVoid) NillableIntNillableArray() (immutable.Option[[]immutable.Option[int64]], bool) {
	return immutable.None[[]immutable.Option[int64]](), false
}

func (NormalVoid) NillableFloatNillableArray() (immutable.Option[[]immutable.Option[float64]], bool) {
	return immutable.None[[]immutable.Option[float64]](), false
}

func (NormalVoid) NillableStringNillableArray() (immutable.Option[[]immutable.Option[string]], bool) {
	return immutable.None[[]immutable.Option[string]](), false
}

func (NormalVoid) NillableBytesNillableArray() (immutable.Option[[]immutable.Option[[]byte]], bool) {
	return immutable.None[[]immutable.Option[[]byte]](), false
}

func (NormalVoid) NillableTimeNillableArray() (immutable.Option[[]immutable.Option[time.Time]], bool) {
	return immutable.None[[]immutable.Option[time.Time]](), false
}

func (NormalVoid) NillableDocumentNillableArray() (immutable.Option[[]immutable.Option[*Document]], bool) {
	return immutable.None[[]immutable.Option[*Document]](), false
}

type baseNormalValue[T any] struct {
	NormalVoid
	val T
}

func (v baseNormalValue[T]) Any() any {
	return v.val
}

func newBaseNormalValue[T any](val T) baseNormalValue[T] {
	return baseNormalValue[T]{val: val}
}

type baseArrayNormalValue[T any] struct {
	NormalVoid
	val T
}

func (v baseArrayNormalValue[T]) Any() any {
	return v.val
}

func (v baseArrayNormalValue[T]) IsArray() bool {
	return true
}

func newBaseArrayNormalValue[T any](val T) baseArrayNormalValue[T] {
	return baseArrayNormalValue[T]{val: val}
}

type baseNillableNormalValue[T any] struct {
	baseNormalValue[immutable.Option[T]]
}

func (v baseNillableNormalValue[T]) IsNil() bool {
	return !v.val.HasValue()
}

func (v baseNillableNormalValue[T]) IsNillable() bool {
	return true
}

func newBaseNillableNormalValue[T any](val immutable.Option[T]) baseNillableNormalValue[T] {
	return baseNillableNormalValue[T]{newBaseNormalValue(val)}
}

type baseNillableArrayNormalValue[T any] struct {
	baseArrayNormalValue[immutable.Option[T]]
}

func (v baseNillableArrayNormalValue[T]) IsNil() bool {
	return !v.val.HasValue()
}

func (v baseNillableArrayNormalValue[T]) IsNillable() bool {
	return true
}

func (v baseNillableArrayNormalValue[T]) IsArray() bool {
	return true
}

func newBaseNillableArrayNormalValue[T any](val immutable.Option[T]) baseNillableArrayNormalValue[T] {
	return baseNillableArrayNormalValue[T]{newBaseArrayNormalValue(val)}
}

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

type normalNil struct {
	NormalVoid
}

func (normalNil) IsNil() bool {
	return true
}

func (normalNil) IsNillable() bool {
	return true
}

func (normalNil) Any() any {
	return nil
}

type normalBool struct {
	baseNormalValue[bool]
}

func (v normalBool) Bool() (bool, bool) {
	return v.val, true
}

type normalInt struct {
	baseNormalValue[int64]
}

func (v normalInt) Int() (int64, bool) {
	return v.val, true
}

type normalFloat struct {
	baseNormalValue[float64]
}

func (v normalFloat) Float() (float64, bool) {
	return v.val, true
}

type normalString struct {
	baseNormalValue[string]
}

func (v normalString) String() (string, bool) {
	return v.val, true
}

type normalBytes struct {
	baseNormalValue[[]byte]
}

func (v normalBytes) Bytes() ([]byte, bool) {
	return v.val, true
}

type normalTime struct {
	baseNormalValue[time.Time]
}

func (v normalTime) Time() (time.Time, bool) {
	return v.val, true
}

type normalDocument struct {
	baseNormalValue[*Document]
}

func (v normalDocument) Document() (*Document, bool) {
	return v.val, true
}

type normalNillableBool struct {
	baseNillableNormalValue[bool]
}

func (v normalNillableBool) NillableBool() (immutable.Option[bool], bool) {
	return v.val, true
}

type normalNillableInt struct {
	baseNillableNormalValue[int64]
}

func (v normalNillableInt) NillableInt() (immutable.Option[int64], bool) {
	return v.val, true
}

type normalNillableFloat struct {
	baseNillableNormalValue[float64]
}

func (v normalNillableFloat) NillableFloat() (immutable.Option[float64], bool) {
	return v.val, true
}

type normalNillableString struct {
	baseNillableNormalValue[string]
}

func (v normalNillableString) NillableString() (immutable.Option[string], bool) {
	return v.val, true
}

type normalNillableBytes struct {
	baseNillableNormalValue[[]byte]
}

func (v normalNillableBytes) NillableBytes() (immutable.Option[[]byte], bool) {
	return v.val, true
}

type normalNillableTime struct {
	baseNillableNormalValue[time.Time]
}

func (v normalNillableTime) NillableTime() (immutable.Option[time.Time], bool) {
	return v.val, true
}

type normalNillableDocument struct {
	baseNillableNormalValue[*Document]
}

func (v normalNillableDocument) NillableDocument() (immutable.Option[*Document], bool) {
	return v.val, true
}

type normalBoolArray struct {
	baseArrayNormalValue[[]bool]
}

func (v normalBoolArray) BoolArray() ([]bool, bool) {
	return v.val, true
}

type normalIntArray struct {
	baseArrayNormalValue[[]int64]
}

func (v normalIntArray) IntArray() ([]int64, bool) {
	return v.val, true
}

type normalFloatArray struct {
	baseArrayNormalValue[[]float64]
}

func (v normalFloatArray) FloatArray() ([]float64, bool) {
	return v.val, true
}

type normalStringArray struct {
	baseArrayNormalValue[[]string]
}

func (v normalStringArray) StringArray() ([]string, bool) {
	return v.val, true
}

type normalBytesArray struct {
	baseArrayNormalValue[[][]byte]
}

func (v normalBytesArray) BytesArray() ([][]byte, bool) {
	return v.val, true
}

type normalTimeArray struct {
	baseArrayNormalValue[[]time.Time]
}

func (v normalTimeArray) TimeArray() ([]time.Time, bool) {
	return v.val, true
}

type normalDocumentArray struct {
	baseArrayNormalValue[[]*Document]
}

func (v normalDocumentArray) DocumentArray() ([]*Document, bool) {
	return v.val, true
}

type normalBoolNillableArray struct {
	baseNillableArrayNormalValue[[]bool]
}

func (v normalBoolNillableArray) BoolNillableArray() (immutable.Option[[]bool], bool) {
	return v.val, true
}

type normalIntNillableArray struct {
	baseNillableArrayNormalValue[[]int64]
}

func (v normalIntNillableArray) IntNillableArray() (immutable.Option[[]int64], bool) {
	return v.val, true
}

type normalFloatNillableArray struct {
	baseNillableArrayNormalValue[[]float64]
}

func (v normalFloatNillableArray) FloatNillableArray() (immutable.Option[[]float64], bool) {
	return v.val, true
}

type normalStringNillableArray struct {
	baseNillableArrayNormalValue[[]string]
}

func (v normalStringNillableArray) StringNillableArray() (immutable.Option[[]string], bool) {
	return v.val, true
}

type normalBytesNillableArray struct {
	baseNillableArrayNormalValue[[][]byte]
}

func (v normalBytesNillableArray) BytesNillableArray() (immutable.Option[[][]byte], bool) {
	return v.val, true
}

type normalTimeNillableArray struct {
	baseNillableArrayNormalValue[[]time.Time]
}

func (v normalTimeNillableArray) TimeNillableArray() (immutable.Option[[]time.Time], bool) {
	return v.val, true
}

type normalDocumentNillableArray struct {
	baseNillableArrayNormalValue[[]*Document]
}

func (v normalDocumentNillableArray) DocumentNillableArray() (immutable.Option[[]*Document], bool) {
	return v.val, true
}

type normalNillableBoolArray struct {
	baseArrayNormalValue[[]immutable.Option[bool]]
}

func (v normalNillableBoolArray) NillableBoolArray() ([]immutable.Option[bool], bool) {
	return v.val, true
}

type normalNillableIntArray struct {
	baseArrayNormalValue[[]immutable.Option[int64]]
}

func (v normalNillableIntArray) NillableIntArray() ([]immutable.Option[int64], bool) {
	return v.val, true
}

type normalNillableFloatArray struct {
	baseArrayNormalValue[[]immutable.Option[float64]]
}

func (v normalNillableFloatArray) NillableFloatArray() ([]immutable.Option[float64], bool) {
	return v.val, true
}

type normalNillableStringArray struct {
	baseArrayNormalValue[[]immutable.Option[string]]
}

func (v normalNillableStringArray) NillableStringArray() ([]immutable.Option[string], bool) {
	return v.val, true
}

type normalNillableBytesArray struct {
	baseArrayNormalValue[[]immutable.Option[[]byte]]
}

func (v normalNillableBytesArray) NillableBytesArray() ([]immutable.Option[[]byte], bool) {
	return v.val, true
}

type normalNillableTimeArray struct {
	baseArrayNormalValue[[]immutable.Option[time.Time]]
}

func (v normalNillableTimeArray) NillableTimeArray() ([]immutable.Option[time.Time], bool) {
	return v.val, true
}

type normalNillableDocumentArray struct {
	baseArrayNormalValue[[]immutable.Option[*Document]]
}

func (v normalNillableDocumentArray) NillableDocumentArray() ([]immutable.Option[*Document], bool) {
	return v.val, true
}

type normalNillableBoolNillableArray struct {
	baseNillableArrayNormalValue[[]immutable.Option[bool]]
}

func (v normalNillableBoolNillableArray) NillableBoolNillableArray() (
	immutable.Option[[]immutable.Option[bool]], bool,
) {
	return v.val, true
}

type normalNillableIntNillableArray struct {
	baseNillableArrayNormalValue[[]immutable.Option[int64]]
}

func (v normalNillableIntNillableArray) NillableIntNillableArray() (
	immutable.Option[[]immutable.Option[int64]], bool,
) {
	return v.val, true
}

type normalNillableFloatNillableArray struct {
	baseNillableArrayNormalValue[[]immutable.Option[float64]]
}

func (v normalNillableFloatNillableArray) NillableFloatNillableArray() (
	immutable.Option[[]immutable.Option[float64]], bool,
) {
	return v.val, true
}

type normalNillableStringNillableArray struct {
	baseNillableArrayNormalValue[[]immutable.Option[string]]
}

func (v normalNillableStringNillableArray) NillableStringNillableArray() (
	immutable.Option[[]immutable.Option[string]], bool,
) {
	return v.val, true
}

type normalNillableBytesNillableArray struct {
	baseNillableArrayNormalValue[[]immutable.Option[[]byte]]
}

func (v normalNillableBytesNillableArray) NillableBytesNillableArray() (
	immutable.Option[[]immutable.Option[[]byte]], bool,
) {
	return v.val, true
}

type normalNillableTimeNillableArray struct {
	baseNillableArrayNormalValue[[]immutable.Option[time.Time]]
}

func (v normalNillableTimeNillableArray) NillableTimeNillableArray() (
	immutable.Option[[]immutable.Option[time.Time]], bool,
) {
	return v.val, true
}

type normalNillableDocumentNillableArray struct {
	baseNillableArrayNormalValue[[]immutable.Option[*Document]]
}

func (v normalNillableDocumentNillableArray) NillableDocumentNillableArray() (
	immutable.Option[[]immutable.Option[*Document]], bool,
) {
	return v.val, true
}

func newNormalInt(val int64) NormalValue {
	return normalInt{newBaseNormalValue(val)}
}

func newNormalFloat(val float64) NormalValue {
	return normalFloat{newBaseNormalValue(val)}
}

func NewNormalValue(val any) (NormalValue, error) {
	if val == nil {
		return normalNil{}, nil
	}
	switch v := val.(type) {
	case bool:
		return NewNormalBool(v), nil
	case int8:
		return newNormalInt(int64(v)), nil
	case int16:
		return newNormalInt(int64(v)), nil
	case int32:
		return newNormalInt(int64(v)), nil
	case int64:
		return newNormalInt(v), nil
	case int:
		return newNormalInt(int64(v)), nil
	case uint8:
		return newNormalInt(int64(v)), nil
	case uint16:
		return newNormalInt(int64(v)), nil
	case uint32:
		return newNormalInt(int64(v)), nil
	case uint64:
		return newNormalInt(int64(v)), nil
	case uint:
		return newNormalInt(int64(v)), nil
	case float32:
		return newNormalFloat(float64(v)), nil
	case float64:
		return newNormalFloat(v), nil
	case string:
		return NewNormalString(v), nil
	case []byte:
		return NewNormalBytes(v), nil
	case time.Time:
		return NewNormalTime(v), nil
	case *Document:
		return NewNormalDocument(v), nil

	case immutable.Option[bool]:
		return NewNormalNillableBool(v), nil
	case immutable.Option[int8]:
		return NewNormalNillableInt(v), nil
	case immutable.Option[int16]:
		return NewNormalNillableInt(v), nil
	case immutable.Option[int32]:
		return NewNormalNillableInt(v), nil
	case immutable.Option[int64]:
		return NewNormalNillableInt(v), nil
	case immutable.Option[int]:
		return NewNormalNillableInt(v), nil
	case immutable.Option[uint8]:
		return NewNormalNillableInt(v), nil
	case immutable.Option[uint16]:
		return NewNormalNillableInt(v), nil
	case immutable.Option[uint32]:
		return NewNormalNillableInt(v), nil
	case immutable.Option[uint64]:
		return NewNormalNillableInt(v), nil
	case immutable.Option[uint]:
		return NewNormalNillableInt(v), nil
	case immutable.Option[float32]:
		return NewNormalNillableFloat(v), nil
	case immutable.Option[float64]:
		return NewNormalNillableFloat(v), nil
	case immutable.Option[string]:
		return NewNormalNillableString(v), nil
	case immutable.Option[[]byte]:
		return NewNormalNillableBytes(v), nil
	case immutable.Option[time.Time]:
		return NewNormalNillableTime(v), nil
	case immutable.Option[*Document]:
		return NewNormalNillableDocument(v), nil

	case []bool:
		return NewNormalBoolArray(v), nil
	case []int8:
		return NewNormalIntArray(v), nil
	case []int16:
		return NewNormalIntArray(v), nil
	case []int32:
		return NewNormalIntArray(v), nil
	case []int64:
		return NewNormalIntArray(v), nil
	case []int:
		return NewNormalIntArray(v), nil
	case []uint16:
		return NewNormalIntArray(v), nil
	case []uint32:
		return NewNormalIntArray(v), nil
	case []uint64:
		return NewNormalIntArray(v), nil
	case []uint:
		return NewNormalIntArray(v), nil
	case []float32:
		return NewNormalFloatArray(v), nil
	case []float64:
		return NewNormalFloatArray(v), nil
	case []string:
		return NewNormalStringArray(v), nil
	case [][]byte:
		return NewNormalBytesArray(v), nil
	case []time.Time:
		return NewNormalTimeArray(v), nil
	case []*Document:
		return NewNormalDocumentArray(v), nil

	case []immutable.Option[bool]:
		return NewNormalNillableBoolArray(v), nil
	case []immutable.Option[int8]:
		return NewNormalNillableIntArray(v), nil
	case []immutable.Option[int16]:
		return NewNormalNillableIntArray(v), nil
	case []immutable.Option[int32]:
		return NewNormalNillableIntArray(v), nil
	case []immutable.Option[int64]:
		return NewNormalNillableIntArray(v), nil
	case []immutable.Option[int]:
		return NewNormalNillableIntArray(v), nil
	case []immutable.Option[uint8]:
		return NewNormalNillableIntArray(v), nil
	case []immutable.Option[uint16]:
		return NewNormalNillableIntArray(v), nil
	case []immutable.Option[uint32]:
		return NewNormalNillableIntArray(v), nil
	case []immutable.Option[uint64]:
		return NewNormalNillableIntArray(v), nil
	case []immutable.Option[uint]:
		return NewNormalNillableIntArray(v), nil
	case []immutable.Option[float32]:
		return NewNormalNillableFloatArray(v), nil
	case []immutable.Option[float64]:
		return NewNormalNillableFloatArray(v), nil
	case []immutable.Option[string]:
		return NewNormalNillableStringArray(v), nil
	case []immutable.Option[[]byte]:
		return NewNormalNillableBytesArray(v), nil
	case []immutable.Option[time.Time]:
		return NewNormalNillableTimeArray(v), nil
	case []immutable.Option[*Document]:
		return NewNormalNillableDocumentArray(v), nil

	case immutable.Option[[]bool]:
		return NewNormalBoolNillableArray(v), nil
	case immutable.Option[[]int8]:
		return NewNormalIntNillableArray(v), nil
	case immutable.Option[[]int16]:
		return NewNormalIntNillableArray(v), nil
	case immutable.Option[[]int32]:
		return NewNormalIntNillableArray(v), nil
	case immutable.Option[[]int64]:
		return NewNormalIntNillableArray(v), nil
	case immutable.Option[[]int]:
		return NewNormalIntNillableArray(v), nil
	case immutable.Option[[]uint16]:
		return NewNormalIntNillableArray(v), nil
	case immutable.Option[[]uint32]:
		return NewNormalIntNillableArray(v), nil
	case immutable.Option[[]uint64]:
		return NewNormalIntNillableArray(v), nil
	case immutable.Option[[]uint]:
		return NewNormalIntNillableArray(v), nil
	case immutable.Option[[]float32]:
		return NewNormalFloatNillableArray(v), nil
	case immutable.Option[[]float64]:
		return NewNormalFloatNillableArray(v), nil
	case immutable.Option[[]string]:
		return NewNormalStringNillableArray(v), nil
	case immutable.Option[[][]byte]:
		return NewNormalBytesNillableArray(v), nil
	case immutable.Option[[]time.Time]:
		return NewNormalTimeNillableArray(v), nil
	case immutable.Option[[]*Document]:
		return NewNormalDocumentNillableArray(v), nil

	case immutable.Option[[]immutable.Option[bool]]:
		return NewNormalNillableBoolNillableArray(v), nil
	case immutable.Option[[]immutable.Option[int8]]:
		return NewNormalNillableIntNillableArray(v), nil
	case immutable.Option[[]immutable.Option[int16]]:
		return NewNormalNillableIntNillableArray(v), nil
	case immutable.Option[[]immutable.Option[int32]]:
		return NewNormalNillableIntNillableArray(v), nil
	case immutable.Option[[]immutable.Option[int64]]:
		return NewNormalNillableIntNillableArray(v), nil
	case immutable.Option[[]immutable.Option[int]]:
		return NewNormalNillableIntNillableArray(v), nil
	case immutable.Option[[]immutable.Option[uint8]]:
		return NewNormalNillableIntNillableArray(v), nil
	case immutable.Option[[]immutable.Option[uint16]]:
		return NewNormalNillableIntNillableArray(v), nil
	case immutable.Option[[]immutable.Option[uint32]]:
		return NewNormalNillableIntNillableArray(v), nil
	case immutable.Option[[]immutable.Option[uint64]]:
		return NewNormalNillableIntNillableArray(v), nil
	case immutable.Option[[]immutable.Option[uint]]:
		return NewNormalNillableIntNillableArray(v), nil
	case immutable.Option[[]immutable.Option[float32]]:
		return NewNormalNillableFloatNillableArray(v), nil
	case immutable.Option[[]immutable.Option[float64]]:
		return NewNormalNillableFloatNillableArray(v), nil
	case immutable.Option[[]immutable.Option[string]]:
		return NewNormalNillableStringNillableArray(v), nil
	case immutable.Option[[]immutable.Option[[]byte]]:
		return NewNormalNillableBytesNillableArray(v), nil
	case immutable.Option[[]immutable.Option[time.Time]]:
		return NewNormalNillableTimeNillableArray(v), nil
	case immutable.Option[[]immutable.Option[*Document]]:
		return NewNormalNillableDocumentNillableArray(v), nil

	case []any:
		if len(v) == 0 {
			return nil, NewCanNotNormalizeValue(val)
		}
		first, err := NewNormalValue(v[0])
		if err != nil {
			return nil, err
		}
		if _, ok := first.Bool(); ok {
			return convertAnyArrToTypedArr[bool](v, NewNormalBoolArray, NewNormalNillableBoolArray)
		}
		if _, ok := first.Int(); ok {
			return convertAnyArrToIntOrFloatArr(v)
		}
		if _, ok := first.Float(); ok {
			return convertAnyArrToFloatArr(v)
		}
		if _, ok := first.String(); ok {
			return convertAnyArrToTypedArr[string](v, NewNormalStringArray, NewNormalNillableStringArray)
		}
		if _, ok := first.Bytes(); ok {
			return convertAnyArrToTypedArr[[]byte](v, NewNormalBytesArray, NewNormalNillableBytesArray)
		}
		if _, ok := first.Time(); ok {
			return convertAnyArrToTypedArr[time.Time](v, NewNormalTimeArray, NewNormalNillableTimeArray)
		}
		if _, ok := first.Document(); ok {
			return convertAnyArrToTypedArr[*Document](v, NewNormalDocumentArray, NewNormalNillableDocumentArray)
		}
	}
	return nil, NewCanNotNormalizeValue(val)
}

func convertAnyArrToIntOrFloatArr(arr []any) (NormalValue, error) {
	result := make([]int64, len(arr))
	for i := range arr {
		if arr[i] == nil {
			return convertAnyArrToNillableIntOrFloatArr(arr)
		}
		switch v := arr[i].(type) {
		case int64:
			result[i] = v
		case float64, float32:
			return convertAnyArrToFloatArr(arr)
		case int8:
			result[i] = int64(v)
		case int16:
			result[i] = int64(v)
		case int32:
			result[i] = int64(v)
		case int:
			result[i] = int64(v)
		case uint8:
			result[i] = int64(v)
		case uint16:
			result[i] = int64(v)
		case uint32:
			result[i] = int64(v)
		case uint64:
			result[i] = int64(v)
		case uint:
			result[i] = int64(v)
		default:
			return nil, NewCanNotNormalizeValue(arr)
		}
	}
	return NewNormalIntArray(result), nil
}

func convertAnyArrToNillableIntOrFloatArr(arr []any) (NormalValue, error) {
	result := make([]immutable.Option[int64], len(arr))
	for i := range arr {
		if arr[i] == nil {
			result[i] = immutable.None[int64]()
			continue
		}
		var intVal int64
		switch v := arr[i].(type) {
		case int64:
			intVal = v
		case float64, float32:
			return convertAnyArrToFloatArr(arr)
		case int8:
			intVal = int64(v)
		case int16:
			intVal = int64(v)
		case int32:
			intVal = int64(v)
		case int:
			intVal = int64(v)
		case uint8:
			intVal = int64(v)
		case uint16:
			intVal = int64(v)
		case uint32:
			intVal = int64(v)
		case uint64:
			intVal = int64(v)
		case uint:
			intVal = int64(v)
		default:
			return nil, NewCanNotNormalizeValue(arr)
		}
		result[i] = immutable.Some(intVal)
	}
	return NewNormalNillableIntArray(result), nil
}

func convertAnyArrToFloatArr(arr []any) (NormalValue, error) {
	result := make([]float64, len(arr))
	for i := range arr {
		if arr[i] == nil {
			return convertAnyArrToNillableFloatArr(arr)
		}

		var floatVal float64
		switch v := arr[i].(type) {
		case float64:
			floatVal = v
		case float32:
			floatVal = float64(v)
		case int8:
			floatVal = float64(v)
		case int16:
			floatVal = float64(v)
		case int32:
			floatVal = float64(v)
		case int64:
			floatVal = float64(v)
		case int:
			floatVal = float64(v)
		case uint8:
			floatVal = float64(v)
		case uint16:
			floatVal = float64(v)
		case uint32:
			floatVal = float64(v)
		case uint64:
			floatVal = float64(v)
		case uint:
			floatVal = float64(v)
		default:
			return nil, NewCanNotNormalizeValue(arr)
		}
		result[i] = floatVal
	}
	return NewNormalFloatArray(result), nil
}

func convertAnyArrToNillableFloatArr(arr []any) (NormalValue, error) {
	result := make([]immutable.Option[float64], len(arr))
	for i := range arr {
		if arr[i] == nil {
			result[i] = immutable.None[float64]()
			continue
		}
		var floatVal float64
		switch v := arr[i].(type) {
		case float64:
			floatVal = v
		case float32:
			floatVal = float64(v)
		case int8:
			floatVal = float64(v)
		case int16:
			floatVal = float64(v)
		case int32:
			floatVal = float64(v)
		case int64:
			floatVal = float64(v)
		case int:
			floatVal = float64(v)
		case uint8:
			floatVal = float64(v)
		case uint16:
			floatVal = float64(v)
		case uint32:
			floatVal = float64(v)
		case uint64:
			floatVal = float64(v)
		case uint:
			floatVal = float64(v)
		default:
			return nil, NewCanNotNormalizeValue(arr)
		}
		result[i] = immutable.Some(floatVal)
	}
	return NewNormalNillableFloatArray(result), nil
}

func convertAnyArrToTypedArr[T any](
	arr []any,
	newNormalArr func([]T) NormalValue,
	newNormalNillableArr func([]immutable.Option[T]) NormalValue,
) (NormalValue, error) {
	result := make([]T, len(arr))
	for i := range arr {
		if arr[i] == nil {
			return convertAnyArrToNillableTypedArr[T](arr, newNormalNillableArr)
		}
		if v, ok := arr[i].(T); ok {
			result[i] = v
		} else {
			return nil, NewCanNotNormalizeValue(arr)
		}
	}
	return newNormalArr(result), nil
}

func convertAnyArrToNillableTypedArr[T any](
	arr []any,
	newNormalNillableArr func([]immutable.Option[T]) NormalValue,
) (NormalValue, error) {
	result := make([]immutable.Option[T], len(arr))
	for i := range arr {
		if arr[i] == nil {
			result[i] = immutable.None[T]()
			continue
		}
		if v, ok := arr[i].(T); ok {
			result[i] = immutable.Some(v)
		} else {
			return nil, NewCanNotNormalizeValue(arr)
		}
	}
	return newNormalNillableArr(result), nil
}

func NewNormalNil() NormalValue {
	return normalNil{}
}

func NewNormalBool(val bool) NormalValue {
	return normalBool{baseNormalValue[bool]{val: val}}
}

func NewNormalInt[T constraints.Integer | constraints.Float](val T) NormalValue {
	return normalInt{baseNormalValue[int64]{val: int64(val)}}
}

func NewNormalFloat[T constraints.Integer | constraints.Float](val T) NormalValue {
	return normalFloat{baseNormalValue[float64]{val: float64(val)}}
}

func NewNormalString[T string | []byte](val T) NormalValue {
	return normalString{baseNormalValue[string]{val: string(val)}}
}

func NewNormalBytes[T string | []byte](val T) NormalValue {
	return normalBytes{baseNormalValue[[]byte]{val: []byte(val)}}
}

func NewNormalTime(val time.Time) NormalValue {
	return normalTime{baseNormalValue[time.Time]{val: val}}
}

func NewNormalDocument(val *Document) NormalValue {
	return normalDocument{baseNormalValue[*Document]{val: val}}
}

func NewNormalNillableBool(val immutable.Option[bool]) NormalValue {
	return normalNillableBool{newBaseNillableNormalValue(val)}
}

func NewNormalNillableInt[T constraints.Integer | constraints.Float](val immutable.Option[T]) NormalValue {
	return normalNillableInt{newBaseNillableNormalValue(normalizeNillableNum[int64](val))}
}

func NewNormalNillableFloat[T constraints.Integer | constraints.Float](val immutable.Option[T]) NormalValue {
	return normalNillableFloat{newBaseNillableNormalValue(normalizeNillableNum[float64](val))}
}

func NewNormalNillableString[T string | []byte](val immutable.Option[T]) NormalValue {
	return normalNillableString{newBaseNillableNormalValue(normalizeNillableChars[string](val))}
}

func NewNormalNillableBytes[T string | []byte](val immutable.Option[T]) NormalValue {
	return normalNillableBytes{newBaseNillableNormalValue(normalizeNillableChars[[]byte](val))}
}

func NewNormalNillableTime(val immutable.Option[time.Time]) NormalValue {
	return normalNillableTime{newBaseNillableNormalValue(val)}
}

func NewNormalNillableDocument(val immutable.Option[*Document]) NormalValue {
	return normalNillableDocument{newBaseNillableNormalValue(val)}
}

func NewNormalBoolArray(val []bool) NormalValue {
	return normalBoolArray{newBaseArrayNormalValue(val)}
}

func NewNormalIntArray[T constraints.Integer | constraints.Float](val []T) NormalValue {
	return normalIntArray{newBaseArrayNormalValue(normalizeNumArr[int64](val))}
}

func NewNormalFloatArray[T constraints.Integer | constraints.Float](val []T) NormalValue {
	return normalFloatArray{newBaseArrayNormalValue(normalizeNumArr[float64](val))}
}

func NewNormalStringArray[T string | []byte](val []T) NormalValue {
	return normalStringArray{newBaseArrayNormalValue(normalizeCharsArr[string](val))}
}

func NewNormalBytesArray[T string | []byte](val []T) NormalValue {
	return normalBytesArray{newBaseArrayNormalValue(normalizeCharsArr[[]byte](val))}
}

func NewNormalTimeArray(val []time.Time) NormalValue {
	return normalTimeArray{newBaseArrayNormalValue(val)}
}

func NewNormalDocumentArray(val []*Document) NormalValue {
	return normalDocumentArray{newBaseArrayNormalValue(val)}
}

func NewNormalBoolNillableArray(val immutable.Option[[]bool]) NormalValue {
	return normalBoolNillableArray{newBaseNillableArrayNormalValue(val)}
}

func NewNormalIntNillableArray[T constraints.Integer | constraints.Float](val immutable.Option[[]T]) NormalValue {
	return normalIntNillableArray{newBaseNillableArrayNormalValue(normalizeNumNillableArr[int64](val))}
}

func NewNormalFloatNillableArray[T constraints.Integer | constraints.Float](val immutable.Option[[]T]) NormalValue {
	return normalFloatNillableArray{newBaseNillableArrayNormalValue(normalizeNumNillableArr[float64](val))}
}

func NewNormalStringNillableArray[T string | []byte](val immutable.Option[[]T]) NormalValue {
	return normalStringNillableArray{newBaseNillableArrayNormalValue(normalizeCharsNillableArr[string](val))}
}

func NewNormalBytesNillableArray[T string | []byte](val immutable.Option[[]T]) NormalValue {
	return normalBytesNillableArray{newBaseNillableArrayNormalValue(normalizeCharsNillableArr[[]byte](val))}
}

func NewNormalTimeNillableArray(val immutable.Option[[]time.Time]) NormalValue {
	return normalTimeNillableArray{newBaseNillableArrayNormalValue(val)}
}

func NewNormalDocumentNillableArray(val immutable.Option[[]*Document]) NormalValue {
	return normalDocumentNillableArray{newBaseNillableArrayNormalValue(val)}
}

func NewNormalNillableBoolArray(val []immutable.Option[bool]) NormalValue {
	return normalNillableBoolArray{newBaseArrayNormalValue(val)}
}

func NewNormalNillableIntArray[T constraints.Integer | constraints.Float](val []immutable.Option[T]) NormalValue {
	return normalNillableIntArray{newBaseArrayNormalValue(normalizeNillableNumArr[int64](val))}
}

func NewNormalNillableFloatArray[T constraints.Integer | constraints.Float](
	val []immutable.Option[T],
) NormalValue {
	return normalNillableFloatArray{newBaseArrayNormalValue(normalizeNillableNumArr[float64](val))}
}

func NewNormalNillableStringArray[T string | []byte](val []immutable.Option[T]) NormalValue {
	return normalNillableStringArray{newBaseArrayNormalValue(normalizeNillableCharsArr[string](val))}
}

func NewNormalNillableBytesArray[T string | []byte](val []immutable.Option[T]) NormalValue {
	return normalNillableBytesArray{newBaseArrayNormalValue(normalizeNillableCharsArr[[]byte](val))}
}

func NewNormalNillableTimeArray(val []immutable.Option[time.Time]) NormalValue {
	return normalNillableTimeArray{newBaseArrayNormalValue(val)}
}

func NewNormalNillableDocumentArray(val []immutable.Option[*Document]) NormalValue {
	return normalNillableDocumentArray{newBaseArrayNormalValue(val)}
}

func NewNormalNillableBoolNillableArray(val immutable.Option[[]immutable.Option[bool]]) NormalValue {
	return normalNillableBoolNillableArray{newBaseNillableArrayNormalValue(val)}
}

func NewNormalNillableIntNillableArray[T constraints.Integer | constraints.Float](
	val immutable.Option[[]immutable.Option[T]],
) NormalValue {
	return normalNillableIntNillableArray{
		newBaseNillableArrayNormalValue(normalizeNillableNumNillableArr[int64](val)),
	}
}

func NewNormalNillableFloatNillableArray[T constraints.Integer | constraints.Float](
	val immutable.Option[[]immutable.Option[T]],
) NormalValue {
	return normalNillableFloatNillableArray{
		newBaseNillableArrayNormalValue(normalizeNillableNumNillableArr[float64](val)),
	}
}

func NewNormalNillableStringNillableArray[T string | []byte](val immutable.Option[[]immutable.Option[T]]) NormalValue {
	return normalNillableStringNillableArray{
		newBaseNillableArrayNormalValue(normalizeNillableCharsNillableArr[string](val)),
	}
}

func NewNormalNillableBytesNillableArray[T string | []byte](val immutable.Option[[]immutable.Option[T]]) NormalValue {
	return normalNillableBytesNillableArray{
		newBaseNillableArrayNormalValue(normalizeNillableCharsNillableArr[[]byte](val)),
	}
}

func NewNormalNillableTimeNillableArray(val immutable.Option[[]immutable.Option[time.Time]]) NormalValue {
	return normalNillableTimeNillableArray{newBaseNillableArrayNormalValue(val)}
}

func NewNormalNillableDocumentNillableArray(val immutable.Option[[]immutable.Option[*Document]]) NormalValue {
	return normalNillableDocumentNillableArray{newBaseNillableArrayNormalValue(val)}
}

func normalizeNillableNum[R int64 | float64, T constraints.Integer | constraints.Float](
	val immutable.Option[T],
) immutable.Option[R] {
	if val.HasValue() {
		return immutable.Some(R(val.Value()))
	}
	return immutable.None[R]()
}

func normalizeNumArr[R int64 | float64, T constraints.Integer | constraints.Float](val []T) []R {
	var v any = val
	if arr, ok := v.([]R); ok {
		return arr
	}
	arr := make([]R, len(val))
	for i, v := range val {
		arr[i] = R(v)
	}
	return arr
}

func normalizeNumNillableArr[R int64 | float64, T constraints.Integer | constraints.Float](
	val immutable.Option[[]T],
) immutable.Option[[]R] {
	if val.HasValue() {
		return immutable.Some(normalizeNumArr[R](val.Value()))
	}
	return immutable.None[[]R]()
}

func normalizeNillableNumArr[R int64 | float64, T constraints.Integer | constraints.Float](
	val []immutable.Option[T],
) []immutable.Option[R] {
	var v any = val
	if arr, ok := v.([]immutable.Option[R]); ok {
		return arr
	}
	arr := make([]immutable.Option[R], len(val))
	for i, v := range val {
		arr[i] = normalizeNillableNum[R](v)
	}
	return arr
}

func normalizeNillableNumNillableArr[R int64 | float64, T constraints.Integer | constraints.Float](
	val immutable.Option[[]immutable.Option[T]],
) immutable.Option[[]immutable.Option[R]] {
	if val.HasValue() {
		return immutable.Some(normalizeNillableNumArr[R](val.Value()))
	}
	return immutable.None[[]immutable.Option[R]]()
}

func normalizeNillableChars[R string | []byte, T string | []byte](val immutable.Option[T]) immutable.Option[R] {
	if val.HasValue() {
		return immutable.Some(R(val.Value()))
	}
	return immutable.None[R]()
}

func normalizeCharsArr[R string | []byte, T string | []byte](val []T) []R {
	var v any = val
	if arr, ok := v.([]R); ok {
		return arr
	}
	arr := make([]R, len(val))
	for i, v := range val {
		arr[i] = R(v)
	}
	return arr
}

func normalizeCharsNillableArr[R string | []byte, T string | []byte](val immutable.Option[[]T]) immutable.Option[[]R] {
	if val.HasValue() {
		return immutable.Some(normalizeCharsArr[R](val.Value()))
	}
	return immutable.None[[]R]()
}

func normalizeNillableCharsArr[R string | []byte, T string | []byte](val []immutable.Option[T]) []immutable.Option[R] {
	var v any = val
	if arr, ok := v.([]immutable.Option[R]); ok {
		return arr
	}
	arr := make([]immutable.Option[R], len(val))
	for i, v := range val {
		if v.HasValue() {
			arr[i] = immutable.Some(R(v.Value()))
		} else {
			arr[i] = immutable.None[R]()
		}
	}
	return arr
}

func normalizeNillableCharsNillableArr[R string | []byte, T string | []byte](
	val immutable.Option[[]immutable.Option[T]],
) immutable.Option[[]immutable.Option[R]] {
	if val.HasValue() {
		return immutable.Some(normalizeNillableCharsArr[R](val.Value()))
	}
	return immutable.None[[]immutable.Option[R]]()
}
