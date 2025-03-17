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

	"golang.org/x/exp/constraints"
)

type baseArrayNormalValue[T any] struct {
	NormalVoid
	val T
}

func (v baseArrayNormalValue[T]) Unwrap() any {
	return v.val
}

func (v baseArrayNormalValue[T]) IsArray() bool {
	return true
}

func newBaseArrayNormalValue[T any](val T) baseArrayNormalValue[T] {
	return baseArrayNormalValue[T]{val: val}
}

type normalBoolArray struct {
	baseArrayNormalValue[[]bool]
}

func (v normalBoolArray) BoolArray() ([]bool, bool) {
	return v.val, true
}

func (v normalBoolArray) Equal(other NormalValue) bool {
	return areNormalArraysEqual(v.val, other.BoolArray)
}

type normalIntArray struct {
	baseArrayNormalValue[[]int64]
}

func (v normalIntArray) IntArray() ([]int64, bool) {
	return v.val, true
}

func (v normalIntArray) Equal(other NormalValue) bool {
	return areNormalArraysEqual(v.val, other.IntArray)
}

type normalFloat64Array struct {
	baseArrayNormalValue[[]float64]
}

func (v normalFloat64Array) Float64Array() ([]float64, bool) {
	return v.val, true
}

func (v normalFloat64Array) Equal(other NormalValue) bool {
	return areNormalArraysEqual(v.val, other.Float64Array)
}

type normalFloat32Array struct {
	baseArrayNormalValue[[]float32]
}

func (v normalFloat32Array) Float32Array() ([]float32, bool) {
	return v.val, true
}

func (v normalFloat32Array) Equal(other NormalValue) bool {
	return areNormalArraysEqual(v.val, other.Float32Array)
}

type normalStringArray struct {
	baseArrayNormalValue[[]string]
}

func (v normalStringArray) StringArray() ([]string, bool) {
	return v.val, true
}

func (v normalStringArray) Equal(other NormalValue) bool {
	return areNormalArraysEqual(v.val, other.StringArray)
}

type normalBytesArray struct {
	baseArrayNormalValue[[][]byte]
}

func (v normalBytesArray) BytesArray() ([][]byte, bool) {
	return v.val, true
}

func (v normalBytesArray) Equal(other NormalValue) bool {
	if otherVal, ok := other.BytesArray(); ok {
		return are2DArraysEqual(v.val, otherVal)
	}
	return false
}

type normalTimeArray struct {
	baseArrayNormalValue[[]time.Time]
}

func (v normalTimeArray) TimeArray() ([]time.Time, bool) {
	return v.val, true
}

func (v normalTimeArray) Equal(other NormalValue) bool {
	return areNormalArraysEqual(v.val, other.TimeArray)
}

type normalDocumentArray struct {
	baseArrayNormalValue[[]*Document]
}

func (v normalDocumentArray) DocumentArray() ([]*Document, bool) {
	return v.val, true
}

func (v normalDocumentArray) Equal(other NormalValue) bool {
	return areNormalArraysEqual(v.val, other.DocumentArray)
}

type normalJSONArray struct {
	baseArrayNormalValue[[]JSON]
}

func (v normalJSONArray) JSONArray() ([]JSON, bool) {
	return v.val, true
}

func (v normalJSONArray) Equal(other NormalValue) bool {
	return areNormalArraysEqual(v.val, other.JSONArray)
}

// NewNormalBoolArray creates a new NormalValue that represents a `[]bool` value.
func NewNormalBoolArray(val []bool) NormalValue {
	return normalBoolArray{newBaseArrayNormalValue(val)}
}

// NewNormalIntArray creates a new NormalValue that represents a `[]int64` value.
func NewNormalIntArray[T constraints.Integer | constraints.Float](val []T) NormalValue {
	return normalIntArray{newBaseArrayNormalValue(normalizeNumArr[int64](val))}
}

// NewNormalFloat64Array creates a new NormalValue that represents a `[]float64` value.
func NewNormalFloat64Array[T constraints.Integer | constraints.Float](val []T) NormalValue {
	return normalFloat64Array{newBaseArrayNormalValue(normalizeNumArr[float64](val))}
}

// NewNormalFloat32Array creates a new NormalValue that represents a `[]float32` value.
func NewNormalFloat32Array[T constraints.Integer | constraints.Float](val []T) NormalValue {
	return normalFloat32Array{newBaseArrayNormalValue(normalizeNumArr[float32](val))}
}

// NewNormalStringArray creates a new NormalValue that represents a `[]string` value.
func NewNormalStringArray[T string | []byte](val []T) NormalValue {
	return normalStringArray{newBaseArrayNormalValue(normalizeCharsArr[string](val))}
}

// NewNormalBytesArray creates a new NormalValue that represents a `[][]byte` value.
func NewNormalBytesArray[T string | []byte](val []T) NormalValue {
	return normalBytesArray{newBaseArrayNormalValue(normalizeCharsArr[[]byte](val))}
}

// NewNormalTimeArray creates a new NormalValue that represents a `[]time.Time` value.
func NewNormalTimeArray(val []time.Time) NormalValue {
	return normalTimeArray{newBaseArrayNormalValue(val)}
}

// NewNormalDocumentArray creates a new NormalValue that represents a `[]*Document` value.
func NewNormalDocumentArray(val []*Document) NormalValue {
	return normalDocumentArray{newBaseArrayNormalValue(val)}
}

// NewNormalJSONArray creates a new NormalValue that represents a `[]JSON` value.
func NewNormalJSONArray(val []JSON) NormalValue {
	return normalJSONArray{newBaseArrayNormalValue(val)}
}

func normalizeNumArr[R int64 | float64 | float32, T constraints.Integer | constraints.Float](val []T) []R {
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

func areArraysEqual[T comparable](arr1, arr2 []T) bool {
	if len(arr1) != len(arr2) {
		return false
	}
	for i, v := range arr1 {
		if v != arr2[i] {
			return false
		}
	}
	return true
}

func areNormalArraysEqual[T comparable](val []T, f func() ([]T, bool)) bool {
	if otherVal, ok := f(); ok {
		return areArraysEqual(val, otherVal)
	}
	return false
}

func are2DArraysEqual[T comparable](arr1, arr2 [][]T) bool {
	if len(arr1) != len(arr2) {
		return false
	}
	for i, v := range arr1 {
		if !areArraysEqual(v, arr2[i]) {
			return false
		}
	}
	return true
}
