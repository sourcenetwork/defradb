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
	"github.com/sourcenetwork/defradb/client"
)

// EncodeFieldValue encodes a FieldValue into a byte slice.
// The encoded value is appended to the supplied buffer and the resulting buffer is returned.
func EncodeFieldValue(b []byte, val client.NormalValue, descending bool) []byte {
	if val.IsNil() {
		if descending {
			return EncodeNullDescending(b)
		} else {
			return EncodeNullAscending(b)
		}
	}
	switch {
	case val.IsBool():
		var boolInt int64 = 0
		if val.Bool() {
			boolInt = 1
		}
		if descending {
			return EncodeVarintDescending(b, boolInt)
		}
		return EncodeVarintAscending(b, boolInt)
	case val.IsInt():
		if descending {
			return EncodeVarintDescending(b, val.Int())
		}
		return EncodeVarintAscending(b, val.Int())
	case val.IsFloat():
		if descending {
			return EncodeFloatDescending(b, val.Float())
		}
		return EncodeFloatAscending(b, val.Float())
	case val.IsString():
		if descending {
			return EncodeStringDescending(b, val.String())
		}
		return EncodeStringAscending(b, val.String())
	}

	return b
}

// DecodeFieldValue decodes a field value from a byte slice.
// The decoded value is returned along with the remaining byte slice.
func DecodeFieldValue(b []byte, descending bool) ([]byte, client.NormalValue, error) {
	typ := PeekType(b)
	switch typ {
	case Null:
		b, _ = DecodeIfNull(b)
		return b, client.NewNilNormalValue(), nil
	case Int:
		var v int64
		var err error
		if descending {
			b, v, err = DecodeVarintDescending(b)
		} else {
			b, v, err = DecodeVarintAscending(b)
		}
		if err != nil {
			return nil, client.NormalValue{}, NewErrCanNotDecodeFieldValue(b, err)
		}
		return b, client.NewIntNormalValue(v), nil
	case Float:
		var v float64
		var err error
		if descending {
			b, v, err = DecodeFloatDescending(b)
		} else {
			b, v, err = DecodeFloatAscending(b)
		}
		if err != nil {
			return nil, client.NormalValue{}, NewErrCanNotDecodeFieldValue(b, err)
		}
		return b, client.NewFloatNormalValue(v), nil
	case Bytes, BytesDesc:
		var v []byte
		var err error
		if descending {
			b, v, err = DecodeBytesDescending(b)
		} else {
			b, v, err = DecodeBytesAscending(b)
		}
		if err != nil {
			return nil, client.NormalValue{}, NewErrCanNotDecodeFieldValue(b, err)
		}
		return b, client.NewStringNormalValue(v), nil
	}

	return nil, client.NormalValue{}, NewErrCanNotDecodeFieldValue(b)
}
