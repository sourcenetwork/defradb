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

// Type represents the type of a value encoded by
// Encode{Null,Varint,Uvarint,Float,Bytes}.
//
//go:generate stringer -type=Type
type Type int

const (
	Unknown   Type = 0
	Null      Type = 1
	Int       Type = 3
	Float     Type = 4
	Bytes     Type = 6
	BytesDesc Type = 7
)

// PeekType peeks at the type of the value encoded at the start of b.
func PeekType(b []byte) Type {
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
