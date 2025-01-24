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

// Type represents the type of a value encoded by
// Encode{Null,Varint,Uvarint,Float,Bytes}.
type Type int

const (
	Unknown   Type = 0
	Null      Type = 1
	Bool      Type = 2
	Int       Type = 3
	Float64   Type = 4
	Bytes     Type = 6
	BytesDesc Type = 7
	Time      Type = 8
	JSON      Type = 9
	Float32   Type = 10
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
		case m >= float32NaN && m <= float32NaNDesc:
			return Float32
		case m >= float64NaN && m <= float64NaNDesc:
			return Float64
		case m == timeMarker:
			return Time
		case m == falseMarker, m == trueMarker:
			return Bool
		case m == jsonMarker:
			return JSON
		}
	}
	return Unknown
}
