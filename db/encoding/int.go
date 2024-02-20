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
	"encoding/binary"
	"math"
)

// EncodeUint64Ascending encodes the uint64 value using a big-endian 8 byte
// representation. The bytes are appended to the supplied buffer and
// the final buffer is returned.
func EncodeUint64Ascending(b []byte, v uint64) []byte {
	return append(b,
		byte(v>>56), byte(v>>48), byte(v>>40), byte(v>>32),
		byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

// EncodeUint64Descending encodes the uint64 value so that it sorts in
// reverse order, from largest to smallest.
func EncodeUint64Descending(b []byte, v uint64) []byte {
	return EncodeUint64Ascending(b, ^v)
}

// DecodeUint64Ascending decodes a uint64 from the input buffer, treating
// the input as a big-endian 8 byte uint64 representation. The remainder
// of the input buffer and the decoded uint64 are returned.
func DecodeUint64Ascending(b []byte) ([]byte, uint64, error) {
	if len(b) < 8 {
		return nil, 0, errors.Errorf("insufficient bytes to decode uint64 int value")
	}
	v := binary.BigEndian.Uint64(b)
	return b[8:], v, nil
}

// DecodeUint64Descending decodes a uint64 value which was encoded
// using EncodeUint64Descending.
func DecodeUint64Descending(b []byte) ([]byte, uint64, error) {
	leftover, v, err := DecodeUint64Ascending(b)
	return leftover, ^v, err
}

// EncodeVarintAscending encodes the int64 value using a variable length
// (length-prefixed) representation. The length is encoded as a single
// byte. If the value to be encoded is negative the length is encoded
// as 8-numBytes. If the value is positive it is encoded as
// 8+numBytes. The encoded bytes are appended to the supplied buffer
// and the final buffer is returned.
func EncodeVarintAscending(b []byte, v int64) []byte {
	if v < 0 {
		switch {
		case v >= -0xff:
			return append(b, IntMin+7, byte(v))
		case v >= -0xffff:
			return append(b, IntMin+6, byte(v>>8), byte(v))
		case v >= -0xffffff:
			return append(b, IntMin+5, byte(v>>16), byte(v>>8), byte(v))
		case v >= -0xffffffff:
			return append(b, IntMin+4, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
		case v >= -0xffffffffff:
			return append(b, IntMin+3, byte(v>>32), byte(v>>24), byte(v>>16), byte(v>>8),
				byte(v))
		case v >= -0xffffffffffff:
			return append(b, IntMin+2, byte(v>>40), byte(v>>32), byte(v>>24), byte(v>>16),
				byte(v>>8), byte(v))
		case v >= -0xffffffffffffff:
			return append(b, IntMin+1, byte(v>>48), byte(v>>40), byte(v>>32), byte(v>>24),
				byte(v>>16), byte(v>>8), byte(v))
		default:
			return append(b, IntMin, byte(v>>56), byte(v>>48), byte(v>>40), byte(v>>32),
				byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
		}
	}
	return EncodeUvarintAscending(b, uint64(v))
}

// EncodeVarintDescending encodes the int64 value so that it sorts in reverse
// order, from largest to smallest.
func EncodeVarintDescending(b []byte, v int64) []byte {
	return EncodeVarintAscending(b, ^v)
}

// DecodeVarintAscending decodes a value encoded by EncodeVarintAscending.
func DecodeVarintAscending(b []byte) ([]byte, int64, error) {
	if len(b) == 0 {
		return nil, 0, errors.Errorf("insufficient bytes to decode uvarint value")
	}
	length := int(b[0]) - intZero
	if length < 0 {
		length = -length
		remB := b[1:]
		if len(remB) < length {
			return nil, 0, errors.Errorf("insufficient bytes to decode uvarint value: %q", remB)
		}
		var v int64
		// Use the ones-complement of each encoded byte in order to build
		// up a positive number, then take the ones-complement again to
		// arrive at our negative value.
		for _, t := range remB[:length] {
			v = (v << 8) | int64(^t)
		}
		return remB[length:], ^v, nil
	}

	remB, v, err := DecodeUvarintAscending(b)
	if err != nil {
		return remB, 0, err
	}
	if v > math.MaxInt64 {
		return nil, 0, errors.Errorf("varint %d overflows int64", v)
	}
	return remB, int64(v), nil
}

// DecodeVarintDescending decodes a int64 value which was encoded
// using EncodeVarintDescending.
func DecodeVarintDescending(b []byte) ([]byte, int64, error) {
	leftover, v, err := DecodeVarintAscending(b)
	return leftover, ^v, err
}

// EncodeUvarintAscending encodes the uint64 value using a variable length
// (length-prefixed) representation. The length is encoded as a single
// byte indicating the number of encoded bytes (-8) to follow. See
// EncodeVarintAscending for rationale. The encoded bytes are appended to the
// supplied buffer and the final buffer is returned.
func EncodeUvarintAscending(b []byte, v uint64) []byte {
	switch {
	case v <= intSmall:
		return append(b, intZero+byte(v))
	case v <= 0xff:
		return append(b, IntMax-7, byte(v))
	case v <= 0xffff:
		return append(b, IntMax-6, byte(v>>8), byte(v))
	case v <= 0xffffff:
		return append(b, IntMax-5, byte(v>>16), byte(v>>8), byte(v))
	case v <= 0xffffffff:
		return append(b, IntMax-4, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
	case v <= 0xffffffffff:
		return append(b, IntMax-3, byte(v>>32), byte(v>>24), byte(v>>16), byte(v>>8),
			byte(v))
	case v <= 0xffffffffffff:
		return append(b, IntMax-2, byte(v>>40), byte(v>>32), byte(v>>24), byte(v>>16),
			byte(v>>8), byte(v))
	case v <= 0xffffffffffffff:
		return append(b, IntMax-1, byte(v>>48), byte(v>>40), byte(v>>32), byte(v>>24),
			byte(v>>16), byte(v>>8), byte(v))
	default:
		return append(b, IntMax, byte(v>>56), byte(v>>48), byte(v>>40), byte(v>>32),
			byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
	}
}

// EncodeUvarintDescending encodes the uint64 value so that it sorts in
// reverse order, from largest to smallest.
func EncodeUvarintDescending(b []byte, v uint64) []byte {
	switch {
	case v == 0:
		return append(b, IntMin+8)
	case v <= 0xff:
		v = ^v
		return append(b, IntMin+7, byte(v))
	case v <= 0xffff:
		v = ^v
		return append(b, IntMin+6, byte(v>>8), byte(v))
	case v <= 0xffffff:
		v = ^v
		return append(b, IntMin+5, byte(v>>16), byte(v>>8), byte(v))
	case v <= 0xffffffff:
		v = ^v
		return append(b, IntMin+4, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
	case v <= 0xffffffffff:
		v = ^v
		return append(b, IntMin+3, byte(v>>32), byte(v>>24), byte(v>>16), byte(v>>8),
			byte(v))
	case v <= 0xffffffffffff:
		v = ^v
		return append(b, IntMin+2, byte(v>>40), byte(v>>32), byte(v>>24), byte(v>>16),
			byte(v>>8), byte(v))
	case v <= 0xffffffffffffff:
		v = ^v
		return append(b, IntMin+1, byte(v>>48), byte(v>>40), byte(v>>32), byte(v>>24),
			byte(v>>16), byte(v>>8), byte(v))
	default:
		v = ^v
		return append(b, IntMin, byte(v>>56), byte(v>>48), byte(v>>40), byte(v>>32),
			byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
	}
}

// DecodeUvarintAscending decodes a uvarint encoded uint64 from the input
// buffer. The remainder of the input buffer and the decoded uint64
// are returned.
func DecodeUvarintAscending(b []byte) ([]byte, uint64, error) {
	if len(b) == 0 {
		return nil, 0, errors.Errorf("insufficient bytes to decode uvarint value")
	}
	length := int(b[0]) - intZero
	b = b[1:] // skip length byte
	if length <= intSmall {
		return b, uint64(length), nil
	}
	length -= intSmall
	if length < 0 || length > 8 {
		return nil, 0, errors.Errorf("invalid uvarint length of %d", length)
	} else if len(b) < length {
		return nil, 0, errors.Errorf("insufficient bytes to decode uvarint value: %q", b)
	}
	var v uint64
	// It is faster to range over the elements in a slice than to index
	// into the slice on each loop iteration.
	for _, t := range b[:length] {
		v = (v << 8) | uint64(t)
	}
	return b[length:], v, nil
}

// DecodeUvarintDescending decodes a uint64 value which was encoded
// using EncodeUvarintDescending.
func DecodeUvarintDescending(b []byte) ([]byte, uint64, error) {
	if len(b) == 0 {
		return nil, 0, errors.Errorf("insufficient bytes to decode uvarint value")
	}
	length := intZero - int(b[0])
	b = b[1:] // skip length byte
	if length < 0 || length > 8 {
		return nil, 0, errors.Errorf("invalid uvarint length of %d", length)
	} else if len(b) < length {
		return nil, 0, errors.Errorf("insufficient bytes to decode uvarint value: %q", b)
	}
	var x uint64
	for _, t := range b[:length] {
		x = (x << 8) | uint64(^t)
	}
	return b[length:], x, nil
}
