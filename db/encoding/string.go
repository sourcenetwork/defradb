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
	"reflect"
	"unsafe"
)

// unsafeConvertStringToBytes converts a string to a byte array to be used with
// string encoding functions. Note that the output byte array should not be
// modified if the input string is expected to be used again - doing so could
// violate Go semantics.
func unsafeConvertStringToBytes(s string) []byte {
	// unsafe.StringData output is unspecified for empty string input so always
	// return nil.
	if len(s) == 0 {
		return nil
	}
	// We unsafely convert the string to a []byte to avoid the
	// usual allocation when converting to a []byte. This is
	// kosher because we know that EncodeBytes{,Descending} does
	// not keep a reference to the value it encodes. The first
	// step is getting access to the string internals.
	hdr := (*reflect.StringHeader)(unsafe.Pointer(&s))
	// Next we treat the string data as a maximally sized array which we
	// slice. This usage is safe because the pointer value remains in the string.
	return (*[0x7fffffff]byte)(unsafe.Pointer(hdr.Data))[:len(s):len(s)]
}

// EncodeStringAscending encodes the string value using an escape-based encoding. See
// EncodeBytes for details. The encoded bytes are append to the supplied buffer
// and the resulting buffer is returned.
func EncodeStringAscending(b []byte, s string) []byte {
	return encodeStringAscendingWithTerminatorAndPrefix(b, s, ascendingBytesEscapes.escapedTerm, bytesMarker)
}

// encodeStringAscendingWithTerminatorAndPrefix encodes the string value using an escape-based encoding. See
// EncodeBytes for details. The encoded bytes are append to the supplied buffer
// and the resulting buffer is returned. We can also pass a terminator byte to be used with
// JSON key encoding.
func encodeStringAscendingWithTerminatorAndPrefix(
	b []byte, s string, terminator byte, prefix byte,
) []byte {
	unsafeString := unsafeConvertStringToBytes(s)
	return encodeBytesAscendingWithTerminatorAndPrefix(b, unsafeString, terminator, prefix)
}

// EncodeStringDescending is the descending version of EncodeStringAscending.
func EncodeStringDescending(b []byte, s string) []byte {
	arg := unsafeConvertStringToBytes(s)
	return EncodeBytesDescending(b, arg)
}
