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
	"bytes"
	"testing"
)

func TestEncodeDecodeUnsafeString(t *testing.T) {
	testCases := []struct {
		value   string
		encoded []byte
	}{
		{"\x00\x01a", []byte{bytesMarker, 0x00, escaped00, 1, 'a', escape, escapedTerm}},
		{"\x00a", []byte{bytesMarker, 0x00, escaped00, 'a', escape, escapedTerm}},
		{"\x00\xffa", []byte{bytesMarker, 0x00, escaped00, 0xff, 'a', escape, escapedTerm}},
		{"a", []byte{bytesMarker, 'a', escape, escapedTerm}},
		{"b", []byte{bytesMarker, 'b', escape, escapedTerm}},
		{"b\x00", []byte{bytesMarker, 'b', 0x00, escaped00, escape, escapedTerm}},
		{"b\x00\x00", []byte{bytesMarker, 'b', 0x00, escaped00, 0x00, escaped00, escape, escapedTerm}},
		{"b\x00\x00a", []byte{bytesMarker, 'b', 0x00, escaped00, 0x00, escaped00, 'a', escape, escapedTerm}},
		{"b\xff", []byte{bytesMarker, 'b', 0xff, escape, escapedTerm}},
		{"hello", []byte{bytesMarker, 'h', 'e', 'l', 'l', 'o', escape, escapedTerm}},
	}
	for i, c := range testCases {
		enc := EncodeStringAscending(nil, c.value)
		if !bytes.Equal(enc, c.encoded) {
			t.Errorf("unexpected encoding mismatch for %v. expected [% x], got [% x]",
				c.value, c.encoded, enc)
		}
		if i > 0 {
			if bytes.Compare(testCases[i-1].encoded, enc) >= 0 {
				t.Errorf("%v: expected [% x] to be less than [% x]",
					c.value, testCases[i-1].encoded, enc)
			}
		}
		remainder, dec, err := DecodeBytesAscending(enc, nil)
		if err != nil {
			t.Error(err)
			continue
		}
		if c.value != string(dec) {
			t.Errorf("unexpected decoding mismatch for %v. got %v", c.value, string(dec))
		}
		if len(remainder) != 0 {
			t.Errorf("unexpected remaining bytes: %v", remainder)
		}

		enc = append(enc, "remainder"...)
		remainder, _, err = DecodeBytesAscending(enc, nil)
		if err != nil {
			t.Error(err)
			continue
		}
		if string(remainder) != "remainder" {
			t.Errorf("unexpected remaining bytes: %v", remainder)
		}
	}
}

func TestEncodeDecodeUnsafeStringDescending(t *testing.T) {
	testCases := []struct {
		value   string
		encoded []byte
	}{
		{"hello", []byte{bytesDescMarker, ^byte('h'), ^byte('e'), ^byte('l'), ^byte('l'), ^byte('o'), escapeDesc, escapedTermDesc}},
		{"b\xff", []byte{bytesDescMarker, ^byte('b'), ^byte(0xff), escapeDesc, escapedTermDesc}},
		{"b\x00\x00a", []byte{bytesDescMarker, ^byte('b'), ^byte(0), escaped00Desc, ^byte(0), escaped00Desc, ^byte('a'), escapeDesc, escapedTermDesc}},
		{"b\x00\x00", []byte{bytesDescMarker, ^byte('b'), ^byte(0), escaped00Desc, ^byte(0), escaped00Desc, escapeDesc, escapedTermDesc}},
		{"b\x00", []byte{bytesDescMarker, ^byte('b'), ^byte(0), escaped00Desc, escapeDesc, escapedTermDesc}},
		{"b", []byte{bytesDescMarker, ^byte('b'), escapeDesc, escapedTermDesc}},
		{"a", []byte{bytesDescMarker, ^byte('a'), escapeDesc, escapedTermDesc}},
		{"\x00\xffa", []byte{bytesDescMarker, ^byte(0), escaped00Desc, ^byte(0xff), ^byte('a'), escapeDesc, escapedTermDesc}},
		{"\x00a", []byte{bytesDescMarker, ^byte(0), escaped00Desc, ^byte('a'), escapeDesc, escapedTermDesc}},
		{"\x00\x01a", []byte{bytesDescMarker, ^byte(0), escaped00Desc, ^byte(1), ^byte('a'), escapeDesc, escapedTermDesc}},
	}
	for i, c := range testCases {
		enc := EncodeStringDescending(nil, c.value)
		if !bytes.Equal(enc, c.encoded) {
			t.Errorf("unexpected encoding mismatch for %v. expected [% x], got [% x]",
				c.value, c.encoded, enc)
		}
		if i > 0 {
			if bytes.Compare(testCases[i-1].encoded, enc) >= 0 {
				t.Errorf("%v: expected [% x] to be less than [% x]",
					c.value, testCases[i-1].encoded, enc)
			}
		}
		remainder, dec, err := DecodeBytesDescending(enc, nil)
		if err != nil {
			t.Error(err)
			continue
		}
		if c.value != string(dec) {
			t.Errorf("unexpected decoding mismatch for %v. got [% x]", c.value, string(dec))
		}
		if len(remainder) != 0 {
			t.Errorf("unexpected remaining bytes: %v", remainder)
		}

		enc = append(enc, "remainder"...)
		remainder, _, err = DecodeBytesDescending(enc, nil)
		if err != nil {
			t.Error(err)
			continue
		}
		if string(remainder) != "remainder" {
			t.Errorf("unexpected remaining bytes: %v", remainder)
		}
	}
}
