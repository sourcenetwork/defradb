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

// NewNormalNil creates a new NormalValue that represents a nil value of a given field kind.
func NewNormalNil(kind FieldKind) (NormalValue, error) {
	if kind.IsObject() {
		return NewNormalNillableDocument(immutable.None[*Document]()), nil
	}
	switch kind {
	case FieldKind_NILLABLE_BOOL:
		return NewNormalNillableBool(immutable.None[bool]()), nil
	case FieldKind_NILLABLE_INT:
		return NewNormalNillableInt(immutable.None[int64]()), nil
	case FieldKind_NILLABLE_FLOAT:
		return NewNormalNillableFloat(immutable.None[float64]()), nil
	case FieldKind_NILLABLE_DATETIME:
		return NewNormalNillableTime(immutable.None[time.Time]()), nil
	case FieldKind_NILLABLE_STRING, FieldKind_NILLABLE_JSON, FieldKind_DocID:
		return NewNormalNillableString(immutable.None[string]()), nil
	case FieldKind_NILLABLE_BLOB:
		return NewNormalNillableBytes(immutable.None[[]byte]()), nil
	case FieldKind_BOOL_ARRAY:
		return NewNormalBoolNillableArray(immutable.None[[]bool]()), nil
	case FieldKind_INT_ARRAY:
		return NewNormalIntNillableArray(immutable.None[[]int64]()), nil
	case FieldKind_FLOAT_ARRAY:
		return NewNormalFloatNillableArray(immutable.None[[]float64]()), nil
	case FieldKind_STRING_ARRAY:
		return NewNormalStringNillableArray(immutable.None[[]string]()), nil
	case FieldKind_NILLABLE_BOOL_ARRAY:
		return NewNormalNillableBoolNillableArray(immutable.None[[]immutable.Option[bool]]()), nil
	case FieldKind_NILLABLE_INT_ARRAY:
		return NewNormalNillableIntNillableArray(immutable.None[[]immutable.Option[int]]()), nil
	case FieldKind_NILLABLE_FLOAT_ARRAY:
		return NewNormalNillableFloatNillableArray(immutable.None[[]immutable.Option[float64]]()), nil
	case FieldKind_NILLABLE_STRING_ARRAY:
		return NewNormalNillableStringNillableArray(immutable.None[[]immutable.Option[string]]()), nil
	default:
		return nil, NewCanNotMakeNormalNilFromFieldKind(kind)
	}
}
