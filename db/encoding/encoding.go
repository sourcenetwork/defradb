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

const (
	encodedNull = iota
	floatNaN
	floatNeg
	floatZero
	floatPos
	floatNaNDesc
	bytesMarker
	bytesDescMarker

	// These constants define a range of values and are used to determine how many bytes are
	// needed to represent the given uint64 value. The constants IntMin and IntMax define the
	// lower and upper bounds of the range, while intMaxWidth is the maximum width (in bytes)
	// for encoding an integer. intZero is the starting point for encoding small integers,
	// and intSmall represents the threshold below which a value can be encoded in a single byte.

	// IntMin is set to 0x80 (128) to avoid overlap with the ASCII range, enhancing testing clarity.
	IntMin = 0x80 // 128
	// Maximum number of bytes to represent an integer, affecting encoding size.
	intMaxWidth = 8
	// intZero is the base value for encoding non-negative integers, calculated to avoid ASCII conflicts.
	intZero = IntMin + intMaxWidth // 136
	// intSmall defines the upper limit for integers that can be encoded in a single byte, considering offset.
	intSmall = IntMax - intZero - intMaxWidth // 109
	// IntMax marks the upper bound for integer tag values, reserved for encoding use.
	IntMax = 0xfd // 253

	encodedNullDesc = 0xff
)

func onesComplement(b []byte) {
	for i := range b {
		b[i] = ^b[i]
	}
}
