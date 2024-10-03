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
	"bytes"
	"time"

	"github.com/sourcenetwork/immutable"
	"golang.org/x/exp/constraints"
)

type normalNillableBoolArray struct {
	baseArrayNormalValue[[]immutable.Option[bool]]
}

func (v normalNillableBoolArray) NillableBoolArray() ([]immutable.Option[bool], bool) {
	return v.val, true
}

func (v normalNillableBoolArray) Equal(other NormalValue) bool {
	return areNormalArraysOfNillablesEqual(v.val, other.NillableBoolArray)
}

type normalNillableIntArray struct {
	baseArrayNormalValue[[]immutable.Option[int64]]
}

func (v normalNillableIntArray) NillableIntArray() ([]immutable.Option[int64], bool) {
	return v.val, true
}

func (v normalNillableIntArray) Equal(other NormalValue) bool {
	return areNormalArraysOfNillablesEqual(v.val, other.NillableIntArray)
}

type normalNillableFloatArray struct {
	baseArrayNormalValue[[]immutable.Option[float64]]
}

func (v normalNillableFloatArray) NillableFloatArray() ([]immutable.Option[float64], bool) {
	return v.val, true
}

func (v normalNillableFloatArray) Equal(other NormalValue) bool {
	return areNormalArraysOfNillablesEqual(v.val, other.NillableFloatArray)
}

type normalNillableStringArray struct {
	baseArrayNormalValue[[]immutable.Option[string]]
}

func (v normalNillableStringArray) NillableStringArray() ([]immutable.Option[string], bool) {
	return v.val, true
}

func (v normalNillableStringArray) Equal(other NormalValue) bool {
	return areNormalArraysOfNillablesEqual(v.val, other.NillableStringArray)
}

type normalNillableBytesArray struct {
	baseArrayNormalValue[[]immutable.Option[[]byte]]
}

func (v normalNillableBytesArray) NillableBytesArray() ([]immutable.Option[[]byte], bool) {
	return v.val, true
}

func (v normalNillableBytesArray) Equal(other NormalValue) bool {
	if otherVal, ok := other.NillableBytesArray(); ok {
		return areArraysOfNillableBytesEqual(v.val, otherVal)
	}
	return false
}

type normalNillableTimeArray struct {
	baseArrayNormalValue[[]immutable.Option[time.Time]]
}

func (v normalNillableTimeArray) NillableTimeArray() ([]immutable.Option[time.Time], bool) {
	return v.val, true
}

func (v normalNillableTimeArray) Equal(other NormalValue) bool {
	return areNormalArraysOfNillablesEqual(v.val, other.NillableTimeArray)
}

type normalNillableDocumentArray struct {
	baseArrayNormalValue[[]immutable.Option[*Document]]
}

func (v normalNillableDocumentArray) NillableDocumentArray() ([]immutable.Option[*Document], bool) {
	return v.val, true
}

func (v normalNillableDocumentArray) Equal(other NormalValue) bool {
	return areNormalArraysOfNillablesEqual(v.val, other.NillableDocumentArray)
}

// NewNormalNillableBoolNillableArray creates a new NormalValue that represents a
// `immutable.Option[[]immutable.Option[bool]]` value.
func NewNormalNillableBoolArray(val []immutable.Option[bool]) NormalValue {
	return normalNillableBoolArray{newBaseArrayNormalValue(val)}
}

// NewNormalNillableIntArray creates a new NormalValue that represents a `[]immutable.Option[int64]` value.
func NewNormalNillableIntArray[T constraints.Integer | constraints.Float](val []immutable.Option[T]) NormalValue {
	return normalNillableIntArray{newBaseArrayNormalValue(normalizeNillableNumArr[int64](val))}
}

// NewNormalNillableFloatArray creates a new NormalValue that represents a `[]immutable.Option[float64]` value.
func NewNormalNillableFloatArray[T constraints.Integer | constraints.Float](
	val []immutable.Option[T],
) NormalValue {
	return normalNillableFloatArray{newBaseArrayNormalValue(normalizeNillableNumArr[float64](val))}
}

// NewNormalNillableStringArray creates a new NormalValue that represents a `[]immutable.Option[string]` value.
func NewNormalNillableStringArray[T string | []byte](val []immutable.Option[T]) NormalValue {
	return normalNillableStringArray{newBaseArrayNormalValue(normalizeNillableCharsArr[string](val))}
}

// NewNormalNillableBytesArray creates a new NormalValue that represents a `[]immutable.Option[[]byte]` value.
func NewNormalNillableBytesArray[T string | []byte](val []immutable.Option[T]) NormalValue {
	return normalNillableBytesArray{newBaseArrayNormalValue(normalizeNillableCharsArr[[]byte](val))}
}

// NewNormalNillableTimeArray creates a new NormalValue that represents a `[]immutable.Option[time.Time]` value.
func NewNormalNillableTimeArray(val []immutable.Option[time.Time]) NormalValue {
	return normalNillableTimeArray{newBaseArrayNormalValue(val)}
}

// NewNormalNillableDocumentArray creates a new NormalValue that represents a `[]immutable.Option[*Document]` value.
func NewNormalNillableDocumentArray(val []immutable.Option[*Document]) NormalValue {
	return normalNillableDocumentArray{newBaseArrayNormalValue(val)}
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

func areNormalArraysOfNillablesEqual[T comparable](
	val []immutable.Option[T],
	f func() ([]immutable.Option[T], bool),
) bool {
	if otherVal, ok := f(); ok {
		return areArraysOfNillablesEqual(val, otherVal)
	}
	return false
}

func areArraysOfNillablesEqual[T comparable](a, b []immutable.Option[T]) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func areArraysOfNillableBytesEqual(a, b []immutable.Option[[]byte]) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v.HasValue() && b[i].HasValue() {
			if !bytes.Equal(v.Value(), b[i].Value()) {
				return false
			}
		} else if v.HasValue() || b[i].HasValue() {
			return false
		}
	}
	return true
}
