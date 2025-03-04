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

import "github.com/sourcenetwork/defradb/client"

const jsonPathEnd = '/'

// EncodeJSONAscending encodes a JSON value in ascending order.
func EncodeJSONAscending(b []byte, v client.JSON) []byte {
	b = encodeJSONPath(b, v)

	b = append(b, jsonPathEnd)

	if str, ok := v.String(); ok {
		b = EncodeStringAscending(b, str)
	} else if num, ok := v.Number(); ok {
		b = EncodeFloat64Ascending(b, num)
	} else if boolVal, ok := v.Bool(); ok {
		b = EncodeBoolAscending(b, boolVal)
	} else if v.IsNull() {
		b = EncodeNullAscending(b)
	}

	return b
}

// EncodeJSONDescending encodes a JSON value in descending order.
func EncodeJSONDescending(b []byte, v client.JSON) []byte {
	b = encodeJSONPath(b, v)

	b = append(b, jsonPathEnd)

	if str, ok := v.String(); ok {
		b = EncodeStringDescending(b, str)
	} else if num, ok := v.Number(); ok {
		b = EncodeFloat64Descending(b, num)
	} else if boolVal, ok := v.Bool(); ok {
		b = EncodeBoolDescending(b, boolVal)
	} else if v.IsNull() {
		b = EncodeNullDescending(b)
	}

	return b
}

// DecodeJSONAscending decodes a JSON value encoded in ascending order.
func DecodeJSONAscending(b []byte) ([]byte, client.JSON, error) {
	return decodeJSON(b, true)
}

// DecodeJSONDescending decodes a JSON value encoded in descending order.
func DecodeJSONDescending(b []byte) ([]byte, client.JSON, error) {
	return decodeJSON(b, false)
}

func decodeJSON(b []byte, ascending bool) ([]byte, client.JSON, error) {
	if PeekType(b) != JSON {
		return b, nil, NewErrMarkersNotFound(b, jsonMarker)
	}

	b = b[1:] // Skip the JSON marker
	b, path, err := decodeJSONPath(b)
	if err != nil {
		return b, nil, err
	}

	b = b[1:] // Skip the path end marker

	var jsonValue any

	switch PeekType(b) {
	case Bytes, BytesDesc:
		var v []byte
		if ascending {
			b, v, err = DecodeBytesAscending(b)
		} else {
			b, v, err = DecodeBytesDescending(b)
		}
		if err != nil {
			return nil, nil, err
		}
		jsonValue = string(v)
	case Float64:
		if ascending {
			b, jsonValue, err = DecodeFloat64Ascending(b)
		} else {
			b, jsonValue, err = DecodeFloat64Descending(b)
		}
	case Bool:
		if ascending {
			b, jsonValue, err = DecodeBoolAscending(b)
		} else {
			b, jsonValue, err = DecodeBoolDescending(b)
		}
	case Null:
		b = decodeNull(b)
	default:
		err = NewErrInvalidJSONPayload(b, path.String())
	}

	if err != nil {
		return b, nil, err
	}

	result, err := client.NewJSONWithPath(jsonValue, path)

	if err != nil {
		return b, nil, err
	}

	return b, result, nil
}

func decodeJSONPath(b []byte) ([]byte, client.JSONPath, error) {
	var path client.JSONPath
	for {
		if len(b) == 0 {
			break
		}
		if b[0] == ascendingBytesEscapes.escapedTerm {
			b = b[1:]
			break
		}

		if PeekType(b) == Bytes {
			remainder, part, err := DecodeBytesAscending(b)
			if err != nil {
				return b, nil, NewErrInvalidJSONPath(b, err)
			}
			path = path.AppendProperty(string(part))
			b = remainder
		} else {
			// a part of the path can be either a property or an index, so if the type of the underlying
			// encoded value is not Bytes it must be Uvarint.
			rem, part, err := DecodeUvarintAscending(b)
			if err != nil {
				return b, nil, NewErrInvalidJSONPath(b, err)
			}
			path = path.AppendIndex(part)
			b = rem
		}
	}
	return b, path, nil
}

func encodeJSONPath(b []byte, v client.JSON) []byte {
	b = append(b, jsonMarker)
	for _, part := range v.GetPath() {
		if prop, ok := part.Property(); ok {
			pathBytes := unsafeConvertStringToBytes(prop)
			b = EncodeBytesAscending(b, pathBytes)
		} else if _, ok := part.Index(); ok {
			// the given json value is an array element and we want all array elements to be
			// distinguishable. That's why we add a constant 0 prefix.
			// We ignore the actual array index value because we have no way of using it at the moment.
			b = EncodeUvarintAscending(b, 0)
		}
	}
	b = append(b, ascendingBytesEscapes.escapedTerm)
	return b
}
