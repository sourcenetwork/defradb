// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package encoding

import (
	"golang.org/x/exp/constraints"

	"github.com/sourcenetwork/defradb/client"
)

func encodeIntFieldValue[T constraints.Integer](b []byte, val T, descending bool) []byte {
	if descending {
		return EncodeVarintDescending(b, int64(val))
	}
	return EncodeVarintAscending(b, int64(val))
}

// EncodeFieldValue encodes a FieldValue into a byte slice.
// The encoded value is appended to the supplied buffer and the resulting buffer is returned.
func EncodeFieldValue(b []byte, val any, descending bool) []byte {
	if val == nil {
		if descending {
			return EncodeNullDescending(b)
		} else {
			return EncodeNullAscending(b)
		}
	}
	switch v := val.(type) {
	case bool:
		var boolInt int64 = 0
		if v {
			boolInt = 1
		}
		if descending {
			return EncodeVarintDescending(b, boolInt)
		}
		return EncodeVarintAscending(b, boolInt)
	case int:
		return encodeIntFieldValue(b, v, descending)
	case int32:
		return encodeIntFieldValue(b, v, descending)
	case int64:
		return encodeIntFieldValue(b, v, descending)
	case float64:
		if descending {
			return EncodeFloatDescending(b, v)
		}
		return EncodeFloatAscending(b, v)
	case string:
		if descending {
			return EncodeStringDescending(b, v)
		}
		return EncodeStringAscending(b, v)
	}

	return b
}

// DecodeFieldValue decodes a FieldValue from a byte slice.
// The decoded value is returned along with the remaining byte slice.
func DecodeFieldValue(b []byte, descending bool) ([]byte, any, error) {
	typ := PeekType(b)
	switch typ {
	case Null:
		b, _ = DecodeIfNull(b)
		return b, nil, nil
	case Int:
		var v int64
		var err error
		if descending {
			b, v, err = DecodeVarintDescending(b)
		} else {
			b, v, err = DecodeVarintAscending(b)
		}
		if err != nil {
			return nil, nil, NewErrCanNotDecodeFieldValue(b, client.FieldKind_NILLABLE_INT, err)
		}
		return b, v, nil
	case Float:
		var v float64
		var err error
		if descending {
			b, v, err = DecodeFloatDescending(b)
		} else {
			b, v, err = DecodeFloatAscending(b)
		}
		if err != nil {
			return nil, nil, NewErrCanNotDecodeFieldValue(b, client.FieldKind_NILLABLE_FLOAT, err)
		}
		return b, v, nil
	case Bytes, BytesDesc:
		var v []byte
		var err error
		if descending {
			b, v, err = DecodeBytesDescending(b)
		} else {
			b, v, err = DecodeBytesAscending(b)
		}
		if err != nil {
			return nil, nil, NewErrCanNotDecodeFieldValue(b, client.FieldKind_NILLABLE_STRING, err)
		}
		return b, v, nil
	}

	return nil, nil, NewErrCanNotDecodeFieldValue(b, client.FieldKind_NILLABLE_STRING)
}
