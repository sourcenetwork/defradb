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
	"unsafe"
)

// unsafeConvertStringToBytes converts a string to a byte array to be used with
// string encoding functions. Note that the output byte array should not be
// modified if the input string is expected to be used again - doing so could
// violate Go semantics.
func unsafeConvertStringToBytes(s string) []byte {
	if len(s) == 0 {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
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
	unsafeString := unsafeConvertStringToBytes(s)
	return EncodeBytesDescending(b, unsafeString)
}
