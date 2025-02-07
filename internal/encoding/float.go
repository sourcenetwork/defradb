// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License

package encoding

import (
	"math"
)

// EncodeFloat32Ascending returns the resulting byte slice with the encoded float32
// appended to b. The encoded format for a float32 value f is, for positive f, the
// encoding of the 32 bits (in IEEE 754 format) re-interpreted as an int32 and
// encoded using EncodeUint32Ascending. For negative f, we keep the sign bit and
// invert all other bits, encoding this value using EncodeUint32Descending. This
// approach was inspired by in github.com/google/orderedcode/orderedcode.go.
//
// One of five single-byte prefix tags are appended to the front of the encoding.
// These tags enforce logical ordering of keys for both ascending and descending
// encoding directions. The tags split the encoded floats into five categories:
// - NaN for an ascending encoding direction
// - Negative valued floats
// - Zero (positive and negative)
// - Positive valued floats
// - NaN for a descending encoding direction
// This ordering ensures that NaNs are always sorted first in either encoding
// direction, and that after them a logical ordering is followed.
func EncodeFloat32Ascending(b []byte, f float32) []byte {
	// Handle the simplistic cases first.
	switch {
	case Float32IsNaN(f):
		return append(b, float32NaN)
	case f == 0:
		// This encodes both positive and negative zero the same. Negative zero uses
		// composite indexes to decode itself correctly.
		return append(b, float32Zero)
	}
	u := math.Float32bits(f)
	if u&(1<<31) != 0 {
		u = ^u
		b = append(b, float32Neg)
	} else {
		b = append(b, float32Pos)
	}
	return EncodeUint32Ascending(b, u)
}

// EncodeFloat32Descending is the descending version of EncodeFloat32Ascending.
func EncodeFloat32Descending(b []byte, f float32) []byte {
	if Float32IsNaN(f) {
		return append(b, float32NaNDesc)
	}
	return EncodeFloat32Ascending(b, -f)
}

// DecodeFloat32Ascending returns the remaining byte slice after decoding and the decoded
// float32 from buf.
func DecodeFloat32Ascending(buf []byte) ([]byte, float32, error) {
	if PeekType(buf) != Float32 {
		return buf, 0, NewErrMarkersNotFound(buf, float32NaN, float32Neg, float32Zero, float32Pos, float32NaNDesc)
	}
	switch buf[0] {
	case float32NaN, float32NaNDesc:
		return buf[1:], Float32NaN(), nil
	case float32Neg:
		b, u, err := DecodeUint32Ascending(buf[1:])
		if err != nil {
			return b, 0, err
		}
		u = ^u
		return b, math.Float32frombits(u), nil
	case float32Zero:
		return buf[1:], 0, nil
	case float32Pos:
		b, u, err := DecodeUint32Ascending(buf[1:])
		if err != nil {
			return b, 0, err
		}
		return b, math.Float32frombits(u), nil
	default:
		return nil, 0, NewErrMarkersNotFound(buf, float32NaN, float32Neg, float32Zero, float32Pos, float32NaNDesc)
	}
}

// DecodeFloat32Descending decodes floats encoded with EncodeFloat32Descending.
func DecodeFloat32Descending(buf []byte) ([]byte, float32, error) {
	b, r, err := DecodeFloat32Ascending(buf)
	return b, -r, err
}

// EncodeFloat64Ascending returns the resulting byte slice with the encoded float64
// appended to b. The encoded format for a float64 value f is, for positive f, the
// encoding of the 64 bits (in IEEE 754 format) re-interpreted as an int64 and
// encoded using EncodeUint64Ascending. For negative f, we keep the sign bit and
// invert all other bits, encoding this value using EncodeUint64Descending. This
// approach was inspired by in github.com/google/orderedcode/orderedcode.go.
//
// One of five single-byte prefix tags are appended to the front of the encoding.
// These tags enforce logical ordering of keys for both ascending and descending
// encoding directions. The tags split the encoded floats into five categories:
// - NaN for an ascending encoding direction
// - Negative valued floats
// - Zero (positive and negative)
// - Positive valued floats
// - NaN for a descending encoding direction
// This ordering ensures that NaNs are always sorted first in either encoding
// direction, and that after them a logical ordering is followed.
func EncodeFloat64Ascending(b []byte, f float64) []byte {
	// Handle the simplistic cases first.
	switch {
	case math.IsNaN(f):
		return append(b, float64NaN)
	case f == 0:
		// This encodes both positive and negative zero the same. Negative zero uses
		// composite indexes to decode itself correctly.
		return append(b, float64Zero)
	}
	u := math.Float64bits(f)
	if u&(1<<63) != 0 {
		u = ^u
		b = append(b, float64Neg)
	} else {
		b = append(b, float64Pos)
	}
	return EncodeUint64Ascending(b, u)
}

// EncodeFloat64Descending is the descending version of EncodeFloatAscending.
func EncodeFloat64Descending(b []byte, f float64) []byte {
	if math.IsNaN(f) {
		return append(b, float64NaNDesc)
	}
	return EncodeFloat64Ascending(b, -f)
}

// DecodeFloat64Ascending returns the remaining byte slice after decoding and the decoded
// float64 from buf.
func DecodeFloat64Ascending(buf []byte) ([]byte, float64, error) {
	if PeekType(buf) != Float64 {
		return buf, 0, NewErrMarkersNotFound(buf, float64NaN, float64Neg, float64Zero, float64Pos, float64NaNDesc)
	}
	switch buf[0] {
	case float64NaN, float64NaNDesc:
		return buf[1:], math.NaN(), nil
	case float64Neg:
		b, u, err := DecodeUint64Ascending(buf[1:])
		if err != nil {
			return b, 0, err
		}
		u = ^u
		return b, math.Float64frombits(u), nil
	case float64Zero:
		return buf[1:], 0, nil
	case float64Pos:
		b, u, err := DecodeUint64Ascending(buf[1:])
		if err != nil {
			return b, 0, err
		}
		return b, math.Float64frombits(u), nil
	default:
		return nil, 0, NewErrMarkersNotFound(buf, float64NaN, float64Neg, float64Zero, float64Pos, float64NaNDesc)
	}
}

// DecodeFloat64Descending decodes floats encoded with EncodeFloatDescending.
func DecodeFloat64Descending(buf []byte) ([]byte, float64, error) {
	b, r, err := DecodeFloat64Ascending(buf)
	return b, -r, err
}

// Float32 specific constants
const (
	uvnan    = 0x7FC00001
	uvinf    = 0x7F800000
	uvneginf = 0xFF800000
)

// Inf returns positive infinity if sign >= 0, negative infinity if sign < 0.
func Float32Inf(sign int) float32 {
	var v uint32
	if sign >= 0 {
		v = uvinf
	} else {
		v = uvneginf
	}
	return math.Float32frombits(v)
}

// NaN returns an IEEE 754 “not-a-number” value.
func Float32NaN() float32 { return math.Float32frombits(uvnan) }

// IsNaN reports whether f is an IEEE 754 “not-a-number” value.
func Float32IsNaN(f float32) (is bool) {
	// IEEE 754 says that only NaNs satisfy f != f.
	return f != f
}

// IsInf reports whether f is an infinity, according to sign.
// If sign > 0, IsInf reports whether f is positive infinity.
// If sign < 0, IsInf reports whether f is negative infinity.
// If sign == 0, IsInf reports whether f is either infinity.
func Float32IsInf(f float32, sign int) bool {
	return sign >= 0 && f > math.MaxFloat32 || sign <= 0 && f < -math.MaxFloat32
}

// Copysign returns a value with the magnitude of f
// and the sign of sign.
func Float32Copysign(f, sign float32) float32 {
	const signBit = 1 << 31
	return math.Float32frombits(math.Float32bits(f)&^signBit | math.Float32bits(sign)&signBit)
}
