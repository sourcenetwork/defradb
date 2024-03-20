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
	if v, ok := val.Bool(); ok {
		var boolInt int64 = 0
		if v {
			boolInt = 1
		}
		if descending {
			return EncodeVarintDescending(b, boolInt)
		}
		return EncodeVarintAscending(b, boolInt)
	}
	if v, ok := val.NillableBool(); ok {
		var boolInt int64 = 0
		if v.Value() {
			boolInt = 1
		}
		if descending {
			return EncodeVarintDescending(b, boolInt)
		}
		return EncodeVarintAscending(b, boolInt)
	}
	if v, ok := val.Int(); ok {
		if descending {
			return EncodeVarintDescending(b, v)
		}
		return EncodeVarintAscending(b, v)
	}
	if v, ok := val.NillableInt(); ok {
		if descending {
			return EncodeVarintDescending(b, v.Value())
		}
		return EncodeVarintAscending(b, v.Value())
	}
	if v, ok := val.Float(); ok {
		if descending {
			return EncodeFloatDescending(b, v)
		}
		return EncodeFloatAscending(b, v)
	}
	if v, ok := val.NillableFloat(); ok {
		if descending {
			return EncodeFloatDescending(b, v.Value())
		}
		return EncodeFloatAscending(b, v.Value())
	}
	if v, ok := val.String(); ok {
		if descending {
			return EncodeStringDescending(b, v)
		}
		return EncodeStringAscending(b, v)
	}
	if v, ok := val.NillableString(); ok {
		if descending {
			return EncodeStringDescending(b, v.Value())
		}
		return EncodeStringAscending(b, v.Value())
	}

	return b
}

// DecodeFieldValue decodes a field value from a byte slice.
// The decoded value is returned along with the remaining byte slice.
func DecodeFieldValue(b []byte, descending bool, kind client.FieldKind) ([]byte, client.NormalValue, error) {
	typ := PeekType(b)
	switch typ {
	case Null:
		b, _ = DecodeIfNull(b)
		nilVal, err := client.NewNormalNil(kind)
		return b, nilVal, err
	case Int:
		var v int64
		var err error
		if descending {
			b, v, err = DecodeVarintDescending(b)
		} else {
			b, v, err = DecodeVarintAscending(b)
		}
		if err != nil {
			return nil, nil, NewErrCanNotDecodeFieldValue(b, err)
		}
		return b, client.NewNormalInt(v), nil
	case Float:
		var v float64
		var err error
		if descending {
			b, v, err = DecodeFloatDescending(b)
		} else {
			b, v, err = DecodeFloatAscending(b)
		}
		if err != nil {
			return nil, nil, NewErrCanNotDecodeFieldValue(b, err)
		}
		return b, client.NewNormalFloat(v), nil
	case Bytes, BytesDesc:
		var v []byte
		var err error
		if descending {
			b, v, err = DecodeBytesDescending(b)
		} else {
			b, v, err = DecodeBytesAscending(b)
		}
		if err != nil {
			return nil, nil, NewErrCanNotDecodeFieldValue(b, err)
		}
		return b, client.NewNormalString(v), nil
	}

	return nil, nil, NewErrCanNotDecodeFieldValue(b)
}
