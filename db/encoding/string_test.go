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
		remainder, dec, err := DecodeUnsafeStringAscending(enc, nil)
		if err != nil {
			t.Error(err)
			continue
		}
		if c.value != dec {
			t.Errorf("unexpected decoding mismatch for %v. got %v", c.value, dec)
		}
		if len(remainder) != 0 {
			t.Errorf("unexpected remaining bytes: %v", remainder)
		}

		enc = append(enc, "remainder"...)
		remainder, _, err = DecodeUnsafeStringAscending(enc, nil)
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
		remainder, dec, err := DecodeUnsafeStringDescending(enc, nil)
		if err != nil {
			t.Error(err)
			continue
		}
		if c.value != dec {
			t.Errorf("unexpected decoding mismatch for %v. got [% x]", c.value, dec)
		}
		if len(remainder) != 0 {
			t.Errorf("unexpected remaining bytes: %v", remainder)
		}

		enc = append(enc, "remainder"...)
		remainder, _, err = DecodeUnsafeStringDescending(enc, nil)
		if err != nil {
			t.Error(err)
			continue
		}
		if string(remainder) != "remainder" {
			t.Errorf("unexpected remaining bytes: %v", remainder)
		}
	}
}
