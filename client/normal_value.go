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
// if the value is of the requested type. They act similar to the Go's type assertion.
//
// All nillable values are represented as [immutable.Option[T]].
type NormalValue interface {
	// Unwrap returns the underlying value.
	// For not nillable values it will return the value as is.
	// For nillable values (of type [immutable.Option[T]]) it will return the value itself
	// if the option has value, otherwise it will return nil.
	Unwrap() any

	// IsNil returns if the value is nil. For not nillable values it will always return false.
	IsNil() bool
	// IsNillable returns if the value can be nil.
	IsNillable() bool
	// IsArray returns if the value is an array.
	IsArray() bool

	// Bool returns the value as a bool. The second return flag is true if the value is a bool.
	Bool() (bool, bool)
	// Int returns the value as an int64. The second return flag is true if the value is an int64.
	Int() (int64, bool)
	// Float returns the value as a float64. The second return flag is true if the value is a float64.
	Float() (float64, bool)
	// String returns the value as a string. The second return flag is true if the value is a string.
	String() (string, bool)
	// Bytes returns the value as a []byte. The second return flag is true if the value is a []byte.
	Bytes() ([]byte, bool)
	// Time returns the value as a [time.Time]. The second return flag is true if the value is a [time.Time].
	Time() (time.Time, bool)
	// Document returns the value as a [*Document]. The second return flag is true if the value is a [*Document].
	Document() (*Document, bool)

	// NillableBool returns the value as a nillable bool.
	// The second return flag is true if the value is [immutable.Option[bool]].
	NillableBool() (immutable.Option[bool], bool)
	// NillableInt returns the value as a nillable int64.
	// The second return flag is true if the value is [immutable.Option[int64]].
	NillableInt() (immutable.Option[int64], bool)
	// NillableFloat returns the value as a nillable float64.
	// The second return flag is true if the value is [immutable.Option[float64]].
	NillableFloat() (immutable.Option[float64], bool)
	// NillableString returns the value as a nillable string.
	// The second return flag is true if the value is [immutable.Option[string]].
	NillableString() (immutable.Option[string], bool)
	// NillableBytes returns the value as a nillable byte slice.
	// The second return flag is true if the value is [immutable.Option[[]byte]].
	NillableBytes() (immutable.Option[[]byte], bool)
	// NillableTime returns the value as a nillable time.Time.
	// The second return flag is true if the value is [immutable.Option[time.Time]].
	NillableTime() (immutable.Option[time.Time], bool)
	// NillableDocument returns the value as a nillable *Document.
	// The second return flag is true if the value is [immutable.Option[*Document]].
	NillableDocument() (immutable.Option[*Document], bool)

	// BoolArray returns the value as a bool array.
	// The second return flag is true if the value is a []bool.
	BoolArray() ([]bool, bool)
	// IntArray returns the value as an int64 array.
	// The second return flag is true if the value is a []int64.
	IntArray() ([]int64, bool)
	// FloatArray returns the value as a float64 array.
	// The second return flag is true if the value is a []float64.
	FloatArray() ([]float64, bool)
	// StringArray returns the value as a string array.
	// The second return flag is true if the value is a []string.
	StringArray() ([]string, bool)
	// BytesArray returns the value as a byte slice array.
	// The second return flag is true if the value is a [][]byte.
	BytesArray() ([][]byte, bool)
	// TimeArray returns the value as a time.Time array.
	// The second return flag is true if the value is a [[]time.Time].
	TimeArray() ([]time.Time, bool)
	// DocumentArray returns the value as a [*Document] array.
	// The second return flag is true if the value is a [[]*Document].
	DocumentArray() ([]*Document, bool)

	// NillableBoolArray returns the value as nillable array of bool elements.
	// The second return flag is true if the value is [immutable.Option[[]bool]].
	BoolNillableArray() (immutable.Option[[]bool], bool)
	// NillableIntArray returns the value as nillable array of int64 elements.
	// The second return flag is true if the value is [immutable.Option[[]int64]].
	IntNillableArray() (immutable.Option[[]int64], bool)
	// NillableFloatArray returns the value as nillable array of float64 elements.
	// The second return flag is true if the value is [immutable.Option[[]float64]].
	FloatNillableArray() (immutable.Option[[]float64], bool)
	// NillableStringArray returns the value as nillable array of string elements.
	// The second return flag is true if the value is [immutable.Option[[]string]].
	StringNillableArray() (immutable.Option[[]string], bool)
	// NillableBytesArray returns the value as nillable array of byte slice elements.
	// The second return flag is true if the value is [immutable.Option[[][]byte]].
	BytesNillableArray() (immutable.Option[[][]byte], bool)
	// NillableTimeArray returns the value as nillable array of [time.Time] elements.
	// The second return flag is true if the value is [immutable.Option[[]time.Time]].
	TimeNillableArray() (immutable.Option[[]time.Time], bool)
	// NillableDocumentArray returns the value as nillable array of [*Document] elements.
	// The second return flag is true if the value is [immutable.Option[[]*Document]].
	DocumentNillableArray() (immutable.Option[[]*Document], bool)

	// NillableBoolArray returns the value as array of nillable bool elements.
	// The second return flag is true if the value is []immutable.Option[bool].
	NillableBoolArray() ([]immutable.Option[bool], bool)
	// NillableIntArray returns the value as array of nillable int64 elements.
	// The second return flag is true if the value is []immutable.Option[int64].
	NillableIntArray() ([]immutable.Option[int64], bool)
	// NillableFloatArray returns the value as array of nillable float64 elements.
	// The second return flag is true if the value is []immutable.Option[float64].
	NillableFloatArray() ([]immutable.Option[float64], bool)
	// NillableStringArray returns the value as array of nillable string elements.
	// The second return flag is true if the value is []immutable.Option[string].
	NillableStringArray() ([]immutable.Option[string], bool)
	// NillableBytesArray returns the value as array of nillable byte slice elements.
	// The second return flag is true if the value is []immutable.Option[[]byte].
	NillableBytesArray() ([]immutable.Option[[]byte], bool)
	// NillableTimeArray returns the value as array of nillable time.Time elements.
	// The second return flag is true if the value is []immutable.Option[time.Time].
	NillableTimeArray() ([]immutable.Option[time.Time], bool)
	// NillableDocumentArray returns the value as array of nillable *Document elements.
	// The second return flag is true if the value is []immutable.Option[*Document].
	NillableDocumentArray() ([]immutable.Option[*Document], bool)

	// NillableBoolNillableArray returns the value as nillable array of nillable bool elements.
	// The second return flag is true if the value is [immutable.Option[[]immutable.Option[bool]]].
	NillableBoolNillableArray() (immutable.Option[[]immutable.Option[bool]], bool)
	// NillableIntNillableArray returns the value as nillable array of nillable int64 elements.
	// The second return flag is true if the value is [immutable.Option[[]immutable.Option[int64]]].
	NillableIntNillableArray() (immutable.Option[[]immutable.Option[int64]], bool)
	// NillableFloatNillableArray returns the value as nillable array of nillable float64 elements.
	// The second return flag is true if the value is [immutable.Option[[]immutable.Option[float64]]].
	NillableFloatNillableArray() (immutable.Option[[]immutable.Option[float64]], bool)
	// NillableStringNillableArray returns the value as nillable array of nillable string elements.
	// The second return flag is true if the value is [immutable.Option[[]immutable.Option[string]]].
	NillableStringNillableArray() (immutable.Option[[]immutable.Option[string]], bool)
	// NillableBytesNillableArray returns the value as nillable array of nillable byte slice elements.
	// The second return flag is true if the value is [immutable.Option[[]immutable.Option[[]byte]]].
	NillableBytesNillableArray() (immutable.Option[[]immutable.Option[[]byte]], bool)
	// NillableTimeNillableArray returns the value as nillable array of nillable time.Time elements.
	// The second return flag is true if the value is [immutable.Option[[]immutable.Option[time.Time]]].
	NillableTimeNillableArray() (immutable.Option[[]immutable.Option[time.Time]], bool)
	// NillableDocumentNillableArray returns the value as nillable array of nillable *Document elements.
	// The second return flag is true if the value is [immutable.Option[[]immutable.Option[*Document]]].
	NillableDocumentNillableArray() (immutable.Option[[]immutable.Option[*Document]], bool)
}
