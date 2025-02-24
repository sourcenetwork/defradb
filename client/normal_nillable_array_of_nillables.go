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

type normalNillableBoolNillableArray struct {
	baseNillableArrayNormalValue[[]immutable.Option[bool]]
}

func (v normalNillableBoolNillableArray) NillableBoolNillableArray() (
	immutable.Option[[]immutable.Option[bool]], bool,
) {
	return v.val, true
}

func (v normalNillableBoolNillableArray) Equal(other NormalValue) bool {
	return areNormalNillableArraysOfNillablesEqual(v.val, other.NillableBoolNillableArray)
}

type normalNillableIntNillableArray struct {
	baseNillableArrayNormalValue[[]immutable.Option[int64]]
}

func (v normalNillableIntNillableArray) NillableIntNillableArray() (
	immutable.Option[[]immutable.Option[int64]], bool,
) {
	return v.val, true
}

func (v normalNillableIntNillableArray) Equal(other NormalValue) bool {
	return areNormalNillableArraysOfNillablesEqual(v.val, other.NillableIntNillableArray)
}

type normalNillableFloat64NillableArray struct {
	baseNillableArrayNormalValue[[]immutable.Option[float64]]
}

func (v normalNillableFloat64NillableArray) NillableFloat64NillableArray() (
	immutable.Option[[]immutable.Option[float64]], bool,
) {
	return v.val, true
}

func (v normalNillableFloat64NillableArray) Equal(other NormalValue) bool {
	return areNormalNillableArraysOfNillablesEqual(v.val, other.NillableFloat64NillableArray)
}

type normalNillableFloat32NillableArray struct {
	baseNillableArrayNormalValue[[]immutable.Option[float32]]
}

func (v normalNillableFloat32NillableArray) NillableFloat32NillableArray() (
	immutable.Option[[]immutable.Option[float32]], bool,
) {
	return v.val, true
}

func (v normalNillableFloat32NillableArray) Equal(other NormalValue) bool {
	return areNormalNillableArraysOfNillablesEqual(v.val, other.NillableFloat32NillableArray)
}

type normalNillableStringNillableArray struct {
	baseNillableArrayNormalValue[[]immutable.Option[string]]
}

func (v normalNillableStringNillableArray) NillableStringNillableArray() (
	immutable.Option[[]immutable.Option[string]], bool,
) {
	return v.val, true
}

func (v normalNillableStringNillableArray) Equal(other NormalValue) bool {
	return areNormalNillableArraysOfNillablesEqual(v.val, other.NillableStringNillableArray)
}

type normalNillableBytesNillableArray struct {
	baseNillableArrayNormalValue[[]immutable.Option[[]byte]]
}

func (v normalNillableBytesNillableArray) NillableBytesNillableArray() (
	immutable.Option[[]immutable.Option[[]byte]], bool,
) {
	return v.val, true
}

func (v normalNillableBytesNillableArray) Equal(other NormalValue) bool {
	if otherVal, ok := other.NillableBytesNillableArray(); ok {
		if v.val.HasValue() && otherVal.HasValue() {
			return areArraysOfNillableBytesEqual(v.val.Value(), otherVal.Value())
		}
		return !v.val.HasValue() && !otherVal.HasValue()
	}
	return false
}

type normalNillableTimeNillableArray struct {
	baseNillableArrayNormalValue[[]immutable.Option[time.Time]]
}

func (v normalNillableTimeNillableArray) NillableTimeNillableArray() (
	immutable.Option[[]immutable.Option[time.Time]], bool,
) {
	return v.val, true
}

func (v normalNillableTimeNillableArray) Equal(other NormalValue) bool {
	return areNormalNillableArraysOfNillablesEqual(v.val, other.NillableTimeNillableArray)
}

type normalNillableDocumentNillableArray struct {
	baseNillableArrayNormalValue[[]immutable.Option[*Document]]
}

func (v normalNillableDocumentNillableArray) NillableDocumentNillableArray() (
	immutable.Option[[]immutable.Option[*Document]], bool,
) {
	return v.val, true
}

func (v normalNillableDocumentNillableArray) Equal(other NormalValue) bool {
	return areNormalNillableArraysOfNillablesEqual(v.val, other.NillableDocumentNillableArray)
}

// NewNormalNillableBoolNillableArray creates a new NormalValue that represents a
// `immutable.Option[[]immutable.Option[bool]]` value.
func NewNormalNillableBoolNillableArray(val immutable.Option[[]immutable.Option[bool]]) NormalValue {
	return normalNillableBoolNillableArray{newBaseNillableArrayNormalValue(val)}
}

// NewNormalNillableIntNillableArray creates a new NormalValue that represents a
// `immutable.Option[[]immutable.Option[int64]]` value.
func NewNormalNillableIntNillableArray[T constraints.Integer | constraints.Float](
	val immutable.Option[[]immutable.Option[T]],
) NormalValue {
	return normalNillableIntNillableArray{
		newBaseNillableArrayNormalValue(normalizeNillableNumNillableArr[int64](val)),
	}
}

// NewNormalNillableFloat64NillableArray creates a new NormalValue that represents a
// `immutable.Option[[]immutable.Option[float64]]` value.
func NewNormalNillableFloat64NillableArray[T constraints.Integer | constraints.Float](
	val immutable.Option[[]immutable.Option[T]],
) NormalValue {
	return normalNillableFloat64NillableArray{
		newBaseNillableArrayNormalValue(normalizeNillableNumNillableArr[float64](val)),
	}
}

// NewNormalNillableFloat32NillableArray creates a new NormalValue that represents a
// `immutable.Option[[]immutable.Option[float32]]` value.
func NewNormalNillableFloat32NillableArray[T constraints.Integer | constraints.Float](
	val immutable.Option[[]immutable.Option[T]],
) NormalValue {
	return normalNillableFloat32NillableArray{
		newBaseNillableArrayNormalValue(normalizeNillableNumNillableArr[float32](val)),
	}
}

// NewNormalNillableStringNillableArray creates a new NormalValue that represents a
// `immutable.Option[[]immutable.Option[string]]` value.
func NewNormalNillableStringNillableArray[T string | []byte](val immutable.Option[[]immutable.Option[T]]) NormalValue {
	return normalNillableStringNillableArray{
		newBaseNillableArrayNormalValue(normalizeNillableCharsNillableArr[string](val)),
	}
}

// NewNormalNillableBytesNillableArray creates a new NormalValue that represents a
// `immutable.Option[[]immutable.Option[[]byte]]` value.
func NewNormalNillableBytesNillableArray[T string | []byte](val immutable.Option[[]immutable.Option[T]]) NormalValue {
	return normalNillableBytesNillableArray{
		newBaseNillableArrayNormalValue(normalizeNillableCharsNillableArr[[]byte](val)),
	}
}

// NewNormalNillableTimeNillableArray creates a new NormalValue that represents a
// `immutable.Option[[]immutable.Option[time.Time]]` value.
func NewNormalNillableTimeNillableArray(val immutable.Option[[]immutable.Option[time.Time]]) NormalValue {
	return normalNillableTimeNillableArray{newBaseNillableArrayNormalValue(val)}
}

// NewNormalNillableDocumentNillableArray creates a new NormalValue that represents a
// `immutable.Option[[]immutable.Option[*Document]]` value.
func NewNormalNillableDocumentNillableArray(val immutable.Option[[]immutable.Option[*Document]]) NormalValue {
	return normalNillableDocumentNillableArray{newBaseNillableArrayNormalValue(val)}
}

func normalizeNillableNumNillableArr[R int64 | float64 | float32, T constraints.Integer | constraints.Float](
	val immutable.Option[[]immutable.Option[T]],
) immutable.Option[[]immutable.Option[R]] {
	if val.HasValue() {
		return immutable.Some(normalizeNillableNumArr[R](val.Value()))
	}
	return immutable.None[[]immutable.Option[R]]()
}

func normalizeNillableCharsNillableArr[R string | []byte, T string | []byte](
	val immutable.Option[[]immutable.Option[T]],
) immutable.Option[[]immutable.Option[R]] {
	if val.HasValue() {
		return immutable.Some(normalizeNillableCharsArr[R](val.Value()))
	}
	return immutable.None[[]immutable.Option[R]]()
}

func areNormalNillableArraysOfNillablesEqual[T comparable](
	val immutable.Option[[]immutable.Option[T]],
	f func() (immutable.Option[[]immutable.Option[T]], bool),
) bool {
	if otherVal, ok := f(); ok {
		return areNillableArraysOfNillablesEqual(val, otherVal)
	}
	return false
}

func areNillableArraysOfNillablesEqual[T comparable](a, b immutable.Option[[]immutable.Option[T]]) bool {
	if a.HasValue() && b.HasValue() {
		return areArraysOfNillablesEqual(a.Value(), b.Value())
	}
	return !a.HasValue() && !b.HasValue()
}
