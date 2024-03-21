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

// NormalVoid is dummy implementation of NormalValue to be embedded in other types.
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
