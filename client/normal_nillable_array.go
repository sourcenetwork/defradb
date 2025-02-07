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

type baseNillableArrayNormalValue[T any] struct {
	baseArrayNormalValue[immutable.Option[T]]
}

func (v baseNillableArrayNormalValue[T]) Unwrap() any {
	if v.val.HasValue() {
		return v.val.Value()
	}
	return nil
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

type normalBoolNillableArray struct {
	baseNillableArrayNormalValue[[]bool]
}

func (v normalBoolNillableArray) BoolNillableArray() (immutable.Option[[]bool], bool) {
	return v.val, true
}

func (v normalBoolNillableArray) Equal(other NormalValue) bool {
	return areOptionsArrEqual(v.val, other.BoolNillableArray)
}

type normalIntNillableArray struct {
	baseNillableArrayNormalValue[[]int64]
}

func (v normalIntNillableArray) IntNillableArray() (immutable.Option[[]int64], bool) {
	return v.val, true
}

func (v normalIntNillableArray) Equal(other NormalValue) bool {
	return areOptionsArrEqual(v.val, other.IntNillableArray)
}

type normalFloat64NillableArray struct {
	baseNillableArrayNormalValue[[]float64]
}

func (v normalFloat64NillableArray) Float64NillableArray() (immutable.Option[[]float64], bool) {
	return v.val, true
}

func (v normalFloat64NillableArray) Equal(other NormalValue) bool {
	return areOptionsArrEqual(v.val, other.Float64NillableArray)
}

type normalFloat32NillableArray struct {
	baseNillableArrayNormalValue[[]float32]
}

func (v normalFloat32NillableArray) Float32NillableArray() (immutable.Option[[]float32], bool) {
	return v.val, true
}

func (v normalFloat32NillableArray) Equal(other NormalValue) bool {
	return areOptionsArrEqual(v.val, other.Float32NillableArray)
}

type normalStringNillableArray struct {
	baseNillableArrayNormalValue[[]string]
}

func (v normalStringNillableArray) StringNillableArray() (immutable.Option[[]string], bool) {
	return v.val, true
}

func (v normalStringNillableArray) Equal(other NormalValue) bool {
	return areOptionsArrEqual(v.val, other.StringNillableArray)
}

type normalBytesNillableArray struct {
	baseNillableArrayNormalValue[[][]byte]
}

func (v normalBytesNillableArray) BytesNillableArray() (immutable.Option[[][]byte], bool) {
	return v.val, true
}

func (v normalBytesNillableArray) Equal(other NormalValue) bool {
	if otherVal, ok := other.BytesNillableArray(); ok {
		if v.val.HasValue() && otherVal.HasValue() {
			return are2DArraysEqual(v.val.Value(), otherVal.Value())
		}
		return !v.val.HasValue() && !otherVal.HasValue()
	}
	return false
}

type normalTimeNillableArray struct {
	baseNillableArrayNormalValue[[]time.Time]
}

func (v normalTimeNillableArray) TimeNillableArray() (immutable.Option[[]time.Time], bool) {
	return v.val, true
}

func (v normalTimeNillableArray) Equal(other NormalValue) bool {
	return areOptionsArrEqual(v.val, other.TimeNillableArray)
}

type normalDocumentNillableArray struct {
	baseNillableArrayNormalValue[[]*Document]
}

func (v normalDocumentNillableArray) DocumentNillableArray() (immutable.Option[[]*Document], bool) {
	return v.val, true
}

func (v normalDocumentNillableArray) Equal(other NormalValue) bool {
	return areOptionsArrEqual(v.val, other.DocumentNillableArray)
}

// NewNormalNillableBoolArray creates a new NormalValue that represents a `immutable.Option[[]bool]` value.
func NewNormalBoolNillableArray(val immutable.Option[[]bool]) NormalValue {
	return normalBoolNillableArray{newBaseNillableArrayNormalValue(val)}
}

// NewNormalNillableIntArray creates a new NormalValue that represents a `immutable.Option[[]int64]` value.
func NewNormalIntNillableArray[T constraints.Integer | constraints.Float](val immutable.Option[[]T]) NormalValue {
	return normalIntNillableArray{newBaseNillableArrayNormalValue(normalizeNumNillableArr[int64](val))}
}

// NewNormalNillableFloat64Array creates a new NormalValue that represents a `immutable.Option[[]float64]` value.
func NewNormalFloat64NillableArray[T constraints.Integer | constraints.Float](val immutable.Option[[]T]) NormalValue {
	return normalFloat64NillableArray{newBaseNillableArrayNormalValue(normalizeNumNillableArr[float64](val))}
}

// NewNormalNillableFloat32Array creates a new NormalValue that represents a `immutable.Option[[]float32]` value.
func NewNormalFloat32NillableArray[T constraints.Integer | constraints.Float](val immutable.Option[[]T]) NormalValue {
	return normalFloat32NillableArray{newBaseNillableArrayNormalValue(normalizeNumNillableArr[float32](val))}
}

// NewNormalNillableStringArray creates a new NormalValue that represents a `immutable.Option[[]string]` value.
func NewNormalStringNillableArray[T string | []byte](val immutable.Option[[]T]) NormalValue {
	return normalStringNillableArray{newBaseNillableArrayNormalValue(normalizeCharsNillableArr[string](val))}
}

// NewNormalNillableBytesArray creates a new NormalValue that represents a `immutable.Option[[][]byte]` value.
func NewNormalBytesNillableArray[T string | []byte](val immutable.Option[[]T]) NormalValue {
	return normalBytesNillableArray{newBaseNillableArrayNormalValue(normalizeCharsNillableArr[[]byte](val))}
}

// NewNormalNillableTimeArray creates a new NormalValue that represents a `immutable.Option[[]time.Time]` value.
func NewNormalTimeNillableArray(val immutable.Option[[]time.Time]) NormalValue {
	return normalTimeNillableArray{newBaseNillableArrayNormalValue(val)}
}

// NewNormalNillableDocumentArray creates a new NormalValue that represents a `immutable.Option[[]*Document]` value.
func NewNormalDocumentNillableArray(val immutable.Option[[]*Document]) NormalValue {
	return normalDocumentNillableArray{newBaseNillableArrayNormalValue(val)}
}

func normalizeNumNillableArr[R int64 | float64 | float32, T constraints.Integer | constraints.Float](
	val immutable.Option[[]T],
) immutable.Option[[]R] {
	if val.HasValue() {
		return immutable.Some(normalizeNumArr[R](val.Value()))
	}
	return immutable.None[[]R]()
}

func normalizeCharsNillableArr[R string | []byte, T string | []byte](val immutable.Option[[]T]) immutable.Option[[]R] {
	if val.HasValue() {
		return immutable.Some(normalizeCharsArr[R](val.Value()))
	}
	return immutable.None[[]R]()
}

func areOptionsArrEqual[T comparable](val immutable.Option[[]T], f func() (immutable.Option[[]T], bool)) bool {
	if otherVal, ok := f(); ok {
		if val.HasValue() && otherVal.HasValue() {
			return areArraysEqual(val.Value(), otherVal.Value())
		}
		return !val.HasValue() && !otherVal.HasValue()
	}
	return false
}
