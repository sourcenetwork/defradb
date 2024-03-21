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
)

// NormalValue is the interface for the normal value types.
// It is used to represent the normal (or standard) values across the system and to avoid
// asserting all possible types like int, int32, int64, etc.
//
// All methods returning a specific type returns the value and the second boolean flag indicating
// if the value is of the requested type.
//
// All nillable values are represented as immutable.Option[T].
type NormalValue interface {
	// Any returns the underlying value.
	// If the value is nillable the result will be of type `immutable.Option[T]`.
	Any() any
	// Unwrap returns the underlying value.
	// For not nillable values it will act as `Any()`
	// For nillable values it will return result of `Value()` of `immutable.Option` if it
	// has value, otherwise it will return `nil`.
	Unwrap() any

	// IsNil returns if the value is nil.
	IsNil() bool
	// IsNillable returns if the value can be nil.
	IsNillable() bool

	// Bool returns the value as a bool.
	Bool() (bool, bool)
	// Int returns the value as an int64.
	Int() (int64, bool)
	// Float returns the value as a float64.
	Float() (float64, bool)
	// String returns the value as a string.
	String() (string, bool)
	// Bytes returns the value as a byte slice.
	Bytes() ([]byte, bool)
	// Time returns the value as a time.Time.
	Time() (time.Time, bool)
	// Document returns the value as a *Document.
	Document() (*Document, bool)

	// NillableBool returns the value as a nillable bool.
	NillableBool() (immutable.Option[bool], bool)
	// NillableInt returns the value as a nillable int64.
	NillableInt() (immutable.Option[int64], bool)
	// NillableFloat returns the value as a nillable float64.
	NillableFloat() (immutable.Option[float64], bool)
	// NillableString returns the value as a nillable string.
	NillableString() (immutable.Option[string], bool)
	// NillableBytes returns the value as a nillable byte slice.
	NillableBytes() (immutable.Option[[]byte], bool)
	// NillableTime returns the value as a nillable time.Time.
	NillableTime() (immutable.Option[time.Time], bool)
	// NillableDocument returns the value as a nillable *Document.
	NillableDocument() (immutable.Option[*Document], bool)

	// IsArray returns if the value is an array.
	IsArray() bool

	// BoolArray returns the value as a bool array.
	BoolArray() ([]bool, bool)
	// IntArray returns the value as an int64 array.
	IntArray() ([]int64, bool)
	// FloatArray returns the value as a float64 array.
	FloatArray() ([]float64, bool)
	// StringArray returns the value as a string array.
	StringArray() ([]string, bool)
	// BytesArray returns the value as a byte slice array.
	BytesArray() ([][]byte, bool)
	// TimeArray returns the value as a time.Time array.
	TimeArray() ([]time.Time, bool)
	// DocumentArray returns the value as a *Document array.
	DocumentArray() ([]*Document, bool)

	// NillableBoolArray returns the value as nillable array of bool elements.
	BoolNillableArray() (immutable.Option[[]bool], bool)
	// NillableIntArray returns the value as nillable array of int64 elements.
	IntNillableArray() (immutable.Option[[]int64], bool)
	// NillableFloatArray returns the value as nillable array of float64 elements.
	FloatNillableArray() (immutable.Option[[]float64], bool)
	// NillableStringArray returns the value as nillable array of string elements.
	StringNillableArray() (immutable.Option[[]string], bool)
	// NillableBytesArray returns the value as nillable array of byte slice elements.
	BytesNillableArray() (immutable.Option[[][]byte], bool)
	// NillableTimeArray returns the value as nillable array of time.Time elements.
	TimeNillableArray() (immutable.Option[[]time.Time], bool)
	// NillableDocumentArray returns the value as nillable array of *Document elements.
	DocumentNillableArray() (immutable.Option[[]*Document], bool)

	// NillableBoolArray returns the value as array of nillable bool elements.
	NillableBoolArray() ([]immutable.Option[bool], bool)
	// NillableIntArray returns the value as array of nillable int64 elements.
	NillableIntArray() ([]immutable.Option[int64], bool)
	// NillableFloatArray returns the value as array of nillable float64 elements.
	NillableFloatArray() ([]immutable.Option[float64], bool)
	// NillableStringArray returns the value as array of nillable string elements.
	NillableStringArray() ([]immutable.Option[string], bool)
	// NillableBytesArray returns the value as array of nillable byte slice elements.
	NillableBytesArray() ([]immutable.Option[[]byte], bool)
	// NillableTimeArray returns the value as array of nillable time.Time elements.
	NillableTimeArray() ([]immutable.Option[time.Time], bool)
	// NillableDocumentArray returns the value as array of nillable *Document elements.
	NillableDocumentArray() ([]immutable.Option[*Document], bool)

	// NillableBoolNillableArray returns the value as nillable array of nillable bool elements.
	NillableBoolNillableArray() (immutable.Option[[]immutable.Option[bool]], bool)
	// NillableIntNillableArray returns the value as nillable array of nillable int64 elements.
	NillableIntNillableArray() (immutable.Option[[]immutable.Option[int64]], bool)
	// NillableFloatNillableArray returns the value as nillable array of nillable float64 elements.
	NillableFloatNillableArray() (immutable.Option[[]immutable.Option[float64]], bool)
	// NillableStringNillableArray returns the value as nillable array of nillable string elements.
	NillableStringNillableArray() (immutable.Option[[]immutable.Option[string]], bool)
	// NillableBytesNillableArray returns the value as nillable array of nillable byte slice elements.
	NillableBytesNillableArray() (immutable.Option[[]immutable.Option[[]byte]], bool)
	// NillableTimeNillableArray returns the value as nillable array of nillable time.Time elements.
	NillableTimeNillableArray() (immutable.Option[[]immutable.Option[time.Time]], bool)
	// NillableDocumentNillableArray returns the value as nillable array of nillable *Document elements.
	NillableDocumentNillableArray() (immutable.Option[[]immutable.Option[*Document]], bool)
}
