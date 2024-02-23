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

import (
	"bytes"
	"regexp"
	"testing"
)

func TestEncodeDecodeBytes(t *testing.T) {
	testCases := []struct {
		value   []byte
		encoded []byte
	}{
		{[]byte{0, 1, 'a'}, []byte{0x12, 0x00, 0xff, 1, 'a', 0x00, 0x01}},
		{[]byte{0, 'a'}, []byte{0x12, 0x00, 0xff, 'a', 0x00, 0x01}},
		{[]byte{0, 0xff, 'a'}, []byte{0x12, 0x00, 0xff, 0xff, 'a', 0x00, 0x01}},
		{[]byte{'a'}, []byte{0x12, 'a', 0x00, 0x01}},
		{[]byte{'b'}, []byte{0x12, 'b', 0x00, 0x01}},
		{[]byte{'b', 0}, []byte{0x12, 'b', 0x00, 0xff, 0x00, 0x01}},
		{[]byte{'b', 0, 0}, []byte{0x12, 'b', 0x00, 0xff, 0x00, 0xff, 0x00, 0x01}},
		{[]byte{'b', 0, 0, 'a'}, []byte{0x12, 'b', 0x00, 0xff, 0x00, 0xff, 'a', 0x00, 0x01}},
		{[]byte{'b', 0xff}, []byte{0x12, 'b', 0xff, 0x00, 0x01}},
		{[]byte("hello"), []byte{0x12, 'h', 'e', 'l', 'l', 'o', 0x00, 0x01}},
	}
	for i, c := range testCases {
		enc := EncodeBytesAscending(nil, c.value)
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
		if !bytes.Equal(c.value, dec) {
			t.Errorf("unexpected decoding mismatch for %v. got %v", c.value, dec)
		}
		if len(remainder) != 0 {
			t.Errorf("unexpected remaining bytes: %v", remainder)
		}

		enc = append(enc, []byte("remainder")...)
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

func TestEncodeDecodeBytesDescending(t *testing.T) {
	testCases := []struct {
		value   []byte
		encoded []byte
	}{
		{[]byte("hello"), []byte{0x13, ^byte('h'), ^byte('e'), ^byte('l'), ^byte('l'), ^byte('o'), 0xff, 0xfe}},
		{[]byte{'b', 0xff}, []byte{0x13, ^byte('b'), 0x00, 0xff, 0xfe}},
		{[]byte{'b', 0, 0, 'a'}, []byte{0x13, ^byte('b'), 0xff, 0x00, 0xff, 0x00, ^byte('a'), 0xff, 0xfe}},
		{[]byte{'b', 0, 0}, []byte{0x13, ^byte('b'), 0xff, 0x00, 0xff, 0x00, 0xff, 0xfe}},
		{[]byte{'b', 0}, []byte{0x13, ^byte('b'), 0xff, 0x00, 0xff, 0xfe}},
		{[]byte{'b'}, []byte{0x13, ^byte('b'), 0xff, 0xfe}},
		{[]byte{'a'}, []byte{0x13, ^byte('a'), 0xff, 0xfe}},
		{[]byte{0, 0xff, 'a'}, []byte{0x13, 0xff, 0x00, 0x00, ^byte('a'), 0xff, 0xfe}},
		{[]byte{0, 'a'}, []byte{0x13, 0xff, 0x00, ^byte('a'), 0xff, 0xfe}},
		{[]byte{0, 1, 'a'}, []byte{0x13, 0xff, 0x00, 0xfe, ^byte('a'), 0xff, 0xfe}},
	}
	for i, c := range testCases {
		enc := EncodeBytesDescending(nil, c.value)
		if !bytes.Equal(enc, c.encoded) {
			t.Errorf("%d: unexpected encoding mismatch for %v ([% x]). expected [% x], got [% x]",
				i, c.value, c.value, c.encoded, enc)
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
		if !bytes.Equal(c.value, dec) {
			t.Errorf("unexpected decoding mismatch for %v. got %v", c.value, dec)
		}
		if len(remainder) != 0 {
			t.Errorf("unexpected remaining bytes: %v", remainder)
		}

		enc = append(enc, []byte("remainder")...)
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

// TestDecodeInvalid tests that decoding invalid bytes panics.
func TestDecodeInvalid(t *testing.T) {
	tests := []struct {
		name    string             // name printed with errors.
		buf     []byte             // buf contains an invalid uvarint to decode.
		pattern string             // pattern matches the panic string.
		decode  func([]byte) error // decode is called with buf.
	}{
		{
			name:    "DecodeVarint, overflows int64",
			buf:     []byte{IntMax, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			pattern: "varint [0-9]+ overflows int64",
			decode:  func(b []byte) error { _, _, err := DecodeVarintAscending(b); return err },
		},
		{
			name:    "Bytes, no marker",
			buf:     []byte{'a'},
			pattern: "did not find marker",
			decode:  func(b []byte) error { _, _, err := DecodeBytesAscending(b, nil); return err },
		},
		{
			name:    "Bytes, no terminator",
			buf:     []byte{bytesMarker, 'a'},
			pattern: "did not find terminator",
			decode:  func(b []byte) error { _, _, err := DecodeBytesAscending(b, nil); return err },
		},
		{
			name:    "Bytes, malformed escape",
			buf:     []byte{bytesMarker, 'a', 0x00},
			pattern: "malformed escape",
			decode:  func(b []byte) error { _, _, err := DecodeBytesAscending(b, nil); return err },
		},
		{
			name:    "Bytes, invalid escape 1",
			buf:     []byte{bytesMarker, 'a', 0x00, 0x00},
			pattern: "unknown escape",
			decode:  func(b []byte) error { _, _, err := DecodeBytesAscending(b, nil); return err },
		},
		{
			name:    "Bytes, invalid escape 2",
			buf:     []byte{bytesMarker, 'a', 0x00, 0x02},
			pattern: "unknown escape",
			decode:  func(b []byte) error { _, _, err := DecodeBytesAscending(b, nil); return err },
		},
		{
			name:    "BytesDescending, no marker",
			buf:     []byte{'a'},
			pattern: "did not find marker",
			decode:  func(b []byte) error { _, _, err := DecodeBytesAscending(b, nil); return err },
		},
		{
			name:    "BytesDescending, no terminator",
			buf:     []byte{bytesDescMarker, ^byte('a')},
			pattern: "did not find terminator",
			decode:  func(b []byte) error { _, _, err := DecodeBytesDescending(b, nil); return err },
		},
		{
			name:    "BytesDescending, malformed escape",
			buf:     []byte{bytesDescMarker, ^byte('a'), 0xff},
			pattern: "malformed escape",
			decode:  func(b []byte) error { _, _, err := DecodeBytesDescending(b, nil); return err },
		},
		{
			name:    "BytesDescending, invalid escape 1",
			buf:     []byte{bytesDescMarker, ^byte('a'), 0xff, 0xff},
			pattern: "unknown escape",
			decode:  func(b []byte) error { _, _, err := DecodeBytesDescending(b, nil); return err },
		},
		{
			name:    "BytesDescending, invalid escape 2",
			buf:     []byte{bytesDescMarker, ^byte('a'), 0xff, 0xfd},
			pattern: "unknown escape",
			decode:  func(b []byte) error { _, _, err := DecodeBytesDescending(b, nil); return err },
		},
	}
	for _, test := range tests {
		err := test.decode(test.buf)
		if !regexp.MustCompile(test.pattern).MatchString(err.Error()) {
			t.Errorf("%q, pattern %q doesn't match %q", test.name, test.pattern, err)
		}
	}
}
