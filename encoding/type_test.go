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
	"testing"
)

func TestPeekType(t *testing.T) {
	testCases := []struct {
		enc []byte
		typ Type
	}{
		{EncodeNullAscending(nil), Null},
		{EncodeNullDescending(nil), Null},
		{EncodeVarintAscending(nil, 0), Int},
		{EncodeVarintDescending(nil, 0), Int},
		{EncodeUvarintAscending(nil, 0), Int},
		{EncodeUvarintDescending(nil, 0), Int},
		{EncodeFloatAscending(nil, 0), Float},
		{EncodeFloatDescending(nil, 0), Float},
		{EncodeBytesAscending(nil, []byte("")), Bytes},
		{EncodeBytesDescending(nil, []byte("")), BytesDesc},
	}
	for i, c := range testCases {
		typ := PeekType(c.enc)
		if c.typ != typ {
			t.Fatalf("%d: expected %d, but found %d", i, c.typ, typ)
		}
	}
}
