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
	"testing"
)

func TestEncodeDecodeUnsafeString(t *testing.T) {
	testCases := []struct {
		value   string
		encoded []byte
	}{
		{"\x00\x01a", []byte{0x12, 0x00, 0xff, 1, 'a', 0x00, 0x01}},
		{"\x00a", []byte{0x12, 0x00, 0xff, 'a', 0x00, 0x01}},
		{"\x00\xffa", []byte{0x12, 0x00, 0xff, 0xff, 'a', 0x00, 0x01}},
		{"a", []byte{0x12, 'a', 0x00, 0x01}},
		{"b", []byte{0x12, 'b', 0x00, 0x01}},
		{"b\x00", []byte{0x12, 'b', 0x00, 0xff, 0x00, 0x01}},
		{"b\x00\x00", []byte{0x12, 'b', 0x00, 0xff, 0x00, 0xff, 0x00, 0x01}},
		{"b\x00\x00a", []byte{0x12, 'b', 0x00, 0xff, 0x00, 0xff, 'a', 0x00, 0x01}},
		{"b\xff", []byte{0x12, 'b', 0xff, 0x00, 0x01}},
		{"hello", []byte{0x12, 'h', 'e', 'l', 'l', 'o', 0x00, 0x01}},
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
		{"hello", []byte{0x13, ^byte('h'), ^byte('e'), ^byte('l'), ^byte('l'), ^byte('o'), 0xff, 0xfe}},
		{"b\xff", []byte{0x13, ^byte('b'), 0x00, 0xff, 0xfe}},
		{"b\x00\x00a", []byte{0x13, ^byte('b'), 0xff, 0x00, 0xff, 0x00, ^byte('a'), 0xff, 0xfe}},
		{"b\x00\x00", []byte{0x13, ^byte('b'), 0xff, 0x00, 0xff, 0x00, 0xff, 0xfe}},
		{"b\x00", []byte{0x13, ^byte('b'), 0xff, 0x00, 0xff, 0xfe}},
		{"b", []byte{0x13, ^byte('b'), 0xff, 0xfe}},
		{"a", []byte{0x13, ^byte('a'), 0xff, 0xfe}},
		{"\x00\xffa", []byte{0x13, 0xff, 0x00, 0x00, ^byte('a'), 0xff, 0xfe}},
		{"\x00a", []byte{0x13, 0xff, 0x00, ^byte('a'), 0xff, 0xfe}},
		{"\x00\x01a", []byte{0x13, 0xff, 0x00, 0xfe, ^byte('a'), 0xff, 0xfe}},
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
