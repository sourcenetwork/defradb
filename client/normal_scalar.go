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

	"golang.org/x/exp/constraints"
)

// NormalValue is dummy implementation of NormalValue to be embedded in other types.
type baseNormalValue[T any] struct {
	NormalVoid
	val T
}

func (v baseNormalValue[T]) Unwrap() any {
	return v.val
}

func newBaseNormalValue[T any](val T) baseNormalValue[T] {
	return baseNormalValue[T]{val: val}
}

type normalBool struct {
	baseNormalValue[bool]
}

func (v normalBool) Bool() (bool, bool) {
	return v.val, true
}

func (v normalBool) Equal(other NormalValue) bool {
	return areNormalScalarsEqual(v.val, other.Bool)
}

type normalInt struct {
	baseNormalValue[int64]
}

func (v normalInt) Int() (int64, bool) {
	return v.val, true
}

func (v normalInt) Equal(other NormalValue) bool {
	return areNormalScalarsEqual(v.val, other.Int)
}

type normalFloat struct {
	baseNormalValue[float64]
}

func (v normalFloat) Float() (float64, bool) {
	return v.val, true
}

func (v normalFloat) Equal(other NormalValue) bool {
	return areNormalScalarsEqual(v.val, other.Float)
}

type normalString struct {
	baseNormalValue[string]
}

func (v normalString) String() (string, bool) {
	return v.val, true
}

func (v normalString) Equal(other NormalValue) bool {
	return areNormalScalarsEqual(v.val, other.String)
}

type normalBytes struct {
	baseNormalValue[[]byte]
}

func (v normalBytes) Bytes() ([]byte, bool) {
	return v.val, true
}

func (v normalBytes) Equal(other NormalValue) bool {
	if otherVal, ok := other.Bytes(); ok {
		return bytes.Equal(v.val, otherVal)
	}
	return false
}

type normalTime struct {
	baseNormalValue[time.Time]
}

func (v normalTime) Time() (time.Time, bool) {
	return v.val, true
}

func (v normalTime) Equal(other NormalValue) bool {
	return areNormalScalarsEqual(v.val, other.Time)
}

type normalDocument struct {
	baseNormalValue[*Document]
}

func (v normalDocument) Equal(other NormalValue) bool {
	return areNormalScalarsEqual(v.val, other.Document)
}

func (v normalDocument) Document() (*Document, bool) {
	return v.val, true
}

func newNormalInt(val int64) NormalValue {
	return normalInt{newBaseNormalValue(val)}
}

func newNormalFloat(val float64) NormalValue {
	return normalFloat{newBaseNormalValue(val)}
}

// NewNormalBool creates a new NormalValue that represents a `bool` value.
func NewNormalBool(val bool) NormalValue {
	return normalBool{baseNormalValue[bool]{val: val}}
}

// NewNormalInt creates a new NormalValue that represents an `int64` value.
func NewNormalInt[T constraints.Integer | constraints.Float](val T) NormalValue {
	return normalInt{baseNormalValue[int64]{val: int64(val)}}
}

// NewNormalFloat creates a new NormalValue that represents a `float64` value.
func NewNormalFloat[T constraints.Integer | constraints.Float](val T) NormalValue {
	return normalFloat{baseNormalValue[float64]{val: float64(val)}}
}

// NewNormalString creates a new NormalValue that represents a `string` value.
func NewNormalString[T string | []byte](val T) NormalValue {
	return normalString{baseNormalValue[string]{val: string(val)}}
}

// NewNormalBytes creates a new NormalValue that represents a `[]byte` value.
func NewNormalBytes[T string | []byte](val T) NormalValue {
	return normalBytes{baseNormalValue[[]byte]{val: []byte(val)}}
}

// NewNormalTime creates a new NormalValue that represents a `time.Time` value.
func NewNormalTime(val time.Time) NormalValue {
	return normalTime{baseNormalValue[time.Time]{val: val}}
}

// NewNormalDocument creates a new NormalValue that represents a `*Document` value.
func NewNormalDocument(val *Document) NormalValue {
	return normalDocument{baseNormalValue[*Document]{val: val}}
}

func areNormalScalarsEqual[T comparable](val T, f func() (T, bool)) bool {
	if otherVal, ok := f(); ok {
		return val == otherVal
	}
	return false
}
