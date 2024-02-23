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
	"fmt"

	"github.com/sourcenetwork/defradb/client"
)

// EncodeFieldValue encodes a FieldValue into a byte slice.
// The encoded value is appended to the supplied buffer and the resulting buffer is returned.
func EncodeFieldValue(b []byte, val *client.FieldValue, descending bool) ([]byte, error) {
	if client.IsNillableKind(val.Kind()) {
		if val.IsNil() {
			if descending {
				return EncodeNullDescending(b), nil
			} else {
				return EncodeNullAscending(b), nil
			}
		}
	}
	switch val.Kind() {
	case client.FieldKind_NILLABLE_BOOL:
		v, err := val.Bool()
		if err != nil {
			return nil, err
		}
		var boolInt int64 = 0
		if v {
			boolInt = 1
		}
		if descending {
			return EncodeVarintDescending(b, boolInt), nil
		}
		return EncodeVarintAscending(b, boolInt), nil
	case client.FieldKind_NILLABLE_INT:
		v, err := val.Int()
		if err != nil {
			return nil, err
		}
		if descending {
			return EncodeVarintDescending(b, int64(v)), nil
		}
		return EncodeVarintAscending(b, int64(v)), nil
	case client.FieldKind_NILLABLE_FLOAT:
		v, err := val.Float()
		if err != nil {
			return nil, err
		}
		if descending {
			return EncodeFloatDescending(b, v), nil
		}
		return EncodeFloatAscending(b, v), nil
	case client.FieldKind_NILLABLE_STRING:
		v, err := val.String()
		if err != nil {
			return nil, err
		}
		if descending {
			return EncodeStringDescending(b, v), nil
		}
		return EncodeStringAscending(b, v), nil
	case client.FieldKind_DocID:
		v, err := val.String()
		if err != nil {
			return nil, err
		}
		if descending {
			return EncodeStringDescending(b, v), nil
		}
		return EncodeStringAscending(b, v), nil
	}

	return nil, nil
}

// DecodeFieldValue decodes a FieldValue from a byte slice.
// The decoded value is returned along with the remaining byte slice.
func DecodeFieldValue(b []byte, kind client.FieldKind, descending bool) ([]byte, *client.FieldValue, error) {
	typ := PeekType(b)
	switch typ {
	case Null:
		if !client.IsNillableKind(kind) {
			return nil, nil, NewErrCanNotDecodeFieldValue(b, kind, fmt.Errorf("nil value for non-nillable kind"))
		}
		b, _ = DecodeIfNull(b)
		return b, client.NewFieldValue(client.NONE_CRDT, nil, kind), nil
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
		if kind == client.FieldKind_NILLABLE_BOOL {
			if v == 0 || v == 1 {
				return b, client.NewFieldValue(client.NONE_CRDT, v != 0, kind), nil
			} else {
				return nil, nil, NewErrCanNotDecodeFieldValue(b, kind)
			}
		}
		if kind == client.FieldKind_NILLABLE_INT {
			return b, client.NewFieldValue(client.NONE_CRDT, v, kind), nil
		}
		return nil, nil, NewErrCanNotDecodeFieldValue(b, kind)
	case Float:
		if kind != client.FieldKind_NILLABLE_FLOAT {
			return nil, nil, NewErrCanNotDecodeFieldValue(b, kind)
		}
		var v float64
		var err error
		if descending {
			b, v, err = DecodeFloatDescending(b)
		} else {
			b, v, err = DecodeFloatAscending(b)
		}
		if err != nil {
			return nil, nil, NewErrCanNotDecodeFieldValue(b, kind, err)
		}
		return b, client.NewFieldValue(client.NONE_CRDT, v, kind), nil
	case Bytes, BytesDesc:
		if kind != client.FieldKind_DocID && kind != client.FieldKind_NILLABLE_STRING {
			return nil, nil, NewErrCanNotDecodeFieldValue(b, kind)
		}
		var v []byte
		var err error
		if descending {
			b, v, err = DecodeBytesDescending(b, []byte{})
		} else {
			b, v, err = DecodeBytesAscending(b, []byte{})
		}
		if err != nil {
			return nil, nil, NewErrCanNotDecodeFieldValue(b, kind, err)
		}
		return b, client.NewFieldValue(client.NONE_CRDT, string(v), kind), nil
	}

	return nil, nil, NewErrCanNotDecodeFieldValue(b, kind)
}
