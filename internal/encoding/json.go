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

import "github.com/sourcenetwork/defradb/client"

// EncodeJSONAscending encodes a JSON value in ascending order.
func EncodeJSONAscending(b []byte, v client.JSON) []byte {
	b = encodeJSONPath(b, v)

	if str, ok := v.String(); ok {
		b = EncodeStringAscending(b, str)
	} else if num, ok := v.Number(); ok {
		b = EncodeFloatAscending(b, num)
	} else if boolVal, ok := v.Bool(); ok {
		b = EncodeBoolAscending(b, boolVal)
	} else if v.IsNull() {
		b = EncodeNullAscending(b)
	} else {
		return nil
	}

	return b
}

// EncodeJSONDescending encodes a JSON value in descending order.
func EncodeJSONDescending(b []byte, v client.JSON) []byte {
	b = encodeJSONPath(b, v)

	if str, ok := v.String(); ok {
		b = EncodeStringDescending(b, str)
	} else if num, ok := v.Number(); ok {
		b = EncodeFloatDescending(b, num)
	} else if boolVal, ok := v.Bool(); ok {
		b = EncodeBoolDescending(b, boolVal)
	} else if v.IsNull() {
		b = EncodeNullDescending(b)
	} else {
		return nil
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
	case Float:
		if ascending {
			b, jsonValue, err = DecodeFloatAscending(b)
		} else {
			b, jsonValue, err = DecodeFloatDescending(b)
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
		err = NewErrInvalidJSONPayload(b, path)
	}

	if err != nil {
		return b, nil, err
	}

	result, err := client.NewJSON(jsonValue)

	if err != nil {
		return b, nil, err
	}

	return b, result, nil
}

func decodeJSONPath(b []byte) ([]byte, []string, error) {
	var path []string
	for {
		if len(b) == 0 {
			break
		}
		if b[0] == ascendingBytesEscapes.escapedTerm {
			b = b[1:]
			break
		}
		rem, part, err := DecodeBytesAscending(b)
		if err != nil {
			return b, nil, NewErrInvalidJSONPath(b, err)
		}
		path = append(path, string(part))
		b = rem
	}
	return b, path, nil
}

func encodeJSONPath(b []byte, v client.JSON) []byte {
	b = append(b, jsonMarker)
	for _, part := range v.GetPath() {
		pathBytes := unsafeConvertStringToBytes(part)
		//b = encodeBytesAscendingWithTerminator(b, pathBytes, ascendingBytesEscapes.escapedTerm)
		b = EncodeBytesAscending(b, pathBytes)
	}
	b = append(b, ascendingBytesEscapes.escapedTerm)
	return b
}
