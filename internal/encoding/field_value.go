// Copyright 2025 Democratized Data Foundation
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
	"time"

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
		if descending {
			return EncodeBoolDescending(b, v)
		}
		return EncodeBoolAscending(b, v)
	}
	if v, ok := val.NillableBool(); ok {
		if descending {
			return EncodeBoolDescending(b, v.Value())
		}
		return EncodeBoolAscending(b, v.Value())
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
	if v, ok := val.Float32(); ok {
		if descending {
			return EncodeFloat32Descending(b, v)
		}
		return EncodeFloat32Ascending(b, v)
	}
	if v, ok := val.NillableFloat32(); ok {
		if descending {
			return EncodeFloat32Descending(b, v.Value())
		}
		return EncodeFloat32Ascending(b, v.Value())
	}
	if v, ok := val.Float64(); ok {
		if descending {
			return EncodeFloat64Descending(b, v)
		}
		return EncodeFloat64Ascending(b, v)
	}
	if v, ok := val.NillableFloat64(); ok {
		if descending {
			return EncodeFloat64Descending(b, v.Value())
		}
		return EncodeFloat64Ascending(b, v.Value())
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
	if v, ok := val.Time(); ok {
		if descending {
			return EncodeTimeDescending(b, v)
		}
		return EncodeTimeAscending(b, v)
	}
	if v, ok := val.NillableTime(); ok {
		if descending {
			return EncodeTimeDescending(b, v.Value())
		}
		return EncodeTimeAscending(b, v.Value())
	}
	if v, ok := val.JSON(); ok {
		if descending {
			return EncodeJSONDescending(b, v)
		}
		return EncodeJSONAscending(b, v)
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
	case Bool:
		var v bool
		var err error
		if descending {
			b, v, err = DecodeBoolDescending(b)
		} else {
			b, v, err = DecodeBoolAscending(b)
		}
		if err != nil {
			return nil, nil, NewErrCanNotDecodeFieldValue(b, kind, err)
		}
		return b, client.NewNormalBool(v), nil
	case Int:
		var v int64
		var err error
		if descending {
			b, v, err = DecodeVarintDescending(b)
		} else {
			b, v, err = DecodeVarintAscending(b)
		}
		if err != nil {
			return nil, nil, NewErrCanNotDecodeFieldValue(b, kind, err)
		}
		return b, client.NewNormalInt(v), nil
	case Float32:
		var v float32
		var err error
		if descending {
			b, v, err = DecodeFloat32Descending(b)
		} else {
			b, v, err = DecodeFloat32Ascending(b)
		}
		if err != nil {
			return nil, nil, NewErrCanNotDecodeFieldValue(b, kind, err)
		}
		return b, client.NewNormalFloat32(v), nil
	case Float64:
		var v float64
		var err error
		if descending {
			b, v, err = DecodeFloat64Descending(b)
		} else {
			b, v, err = DecodeFloat64Ascending(b)
		}
		if err != nil {
			return nil, nil, NewErrCanNotDecodeFieldValue(b, kind, err)
		}
		return b, client.NewNormalFloat64(v), nil
	case Bytes, BytesDesc:
		var v []byte
		var err error
		if descending {
			b, v, err = DecodeBytesDescending(b)
		} else {
			b, v, err = DecodeBytesAscending(b)
		}
		if err != nil {
			return nil, nil, NewErrCanNotDecodeFieldValue(b, kind, err)
		}
		return b, client.NewNormalString(v), nil
	case Time:
		var v time.Time
		var err error
		if descending {
			b, v, err = DecodeTimeDescending(b)
		} else {
			b, v, err = DecodeTimeAscending(b)
		}
		if err != nil {
			return nil, nil, NewErrCanNotDecodeFieldValue(b, kind, err)
		}
		return b, client.NewNormalTime(v), nil
	case JSON:
		var v client.JSON
		var err error
		if descending {
			b, v, err = DecodeJSONDescending(b)
		} else {
			b, v, err = DecodeJSONAscending(b)
		}
		if err != nil {
			return nil, nil, NewErrCanNotDecodeFieldValue(b, kind, err)
		}
		return b, client.NewNormalJSON(v), nil
	}

	return nil, nil, NewErrCanNotDecodeFieldValue(b, kind)
}
