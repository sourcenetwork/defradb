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

// EncodeBoolAscending encodes a boolean value in ascending order.
func EncodeBoolAscending(b []byte, v bool) []byte {
	if v {
		b = append(b, trueMarker)
	} else {
		b = append(b, falseMarker)
	}

	return b
}

// EncodeBoolDescending encodes a boolean value in descending order.
func EncodeBoolDescending(b []byte, v bool) []byte {
	return EncodeBoolAscending(b, !v)
}

// DecodeBoolAscending decodes a boolean value encoded in ascending order.
func DecodeBoolAscending(b []byte) ([]byte, bool, error) {
	if PeekType(b) != Bool {
		return b, false, NewErrMarkersNotFound(b, falseMarker, trueMarker)
	}

	byte0 := b[0]
	return b[1:], byte0 == trueMarker, nil
}

// DecodeBoolDescending decodes a boolean value encoded in descending order.
func DecodeBoolDescending(b []byte) ([]byte, bool, error) {
	b, v, err := DecodeBoolAscending(b)
	return b, !v, err
}
