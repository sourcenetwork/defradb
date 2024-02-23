// Copyright 2014 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package encoding

// Type represents the type of a value encoded by
// Encode{Null,NotNull,Varint,Uvarint,Float,Bytes}.
//
//go:generate stringer -type=Type
type Type int

// Type values.
// TODO(dan, arjun): Make this into a proto enum.
// The 'Type' annotations are necessary for producing stringer-generated values.
const (
	Unknown   Type = 0
	Null      Type = 1
	Int       Type = 3
	Float     Type = 4
	Bytes     Type = 6
	BytesDesc Type = 7 // Bytes encoded descendingly
)

// typMap maps an encoded type byte to a decoded Type. It's got 256 slots, one
// for every possible byte value.
var typMap [256]Type

func init() {
	buf := []byte{0}
	for i := range typMap {
		buf[0] = byte(i)
		typMap[i] = slowPeekType(buf)
	}
}

// PeekType peeks at the type of the value encoded at the start of b.
func PeekType(b []byte) Type {
	if len(b) >= 1 {
		return typMap[b[0]]
	}
	return Unknown
}

// slowPeekType is the old implementation of PeekType. It's used to generate
// the lookup table for PeekType.
func slowPeekType(b []byte) Type {
	if len(b) >= 1 {
		m := b[0]
		switch {
		case m == encodedNull, m == encodedNullDesc:
			return Null
		case m == bytesMarker:
			return Bytes
		case m == bytesDescMarker:
			return BytesDesc
		case m >= IntMin && m <= IntMax:
			return Int
		case m >= floatNaN && m <= floatNaNDesc:
			return Float
		}
	}
	return Unknown
}
