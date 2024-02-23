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
