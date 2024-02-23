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

	"github.com/stretchr/testify/assert"
)

func TestEncodeDecodeBytes(t *testing.T) {
	testCases := []struct {
		value   []byte
		encoded []byte
	}{
		{[]byte{0, 1, 'a'}, []byte{bytesMarker, 0x00, escaped00, 1, 'a', escape, escapedTerm}},
		{[]byte{0, 'a'}, []byte{bytesMarker, 0x00, escaped00, 'a', escape, escapedTerm}},
		{[]byte{0, 0xff, 'a'}, []byte{bytesMarker, 0x00, escaped00, 0xff, 'a', escape, escapedTerm}},
		{[]byte{'a'}, []byte{bytesMarker, 'a', escape, escapedTerm}},
		{[]byte{'b'}, []byte{bytesMarker, 'b', escape, escapedTerm}},
		{[]byte{'b', 0}, []byte{bytesMarker, 'b', 0x00, escaped00, escape, escapedTerm}},
		{[]byte{'b', 0, 0}, []byte{bytesMarker, 'b', 0x00, escaped00, 0x00, escaped00, escape, escapedTerm}},
		{[]byte{'b', 0, 0, 'a'}, []byte{bytesMarker, 'b', 0x00, escaped00, 0x00, escaped00, 'a', escape, escapedTerm}},
		{[]byte{'b', 0xff}, []byte{bytesMarker, 'b', 0xff, escape, escapedTerm}},
		{[]byte("hello"), []byte{bytesMarker, 'h', 'e', 'l', 'l', 'o', escape, escapedTerm}},
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
		{[]byte("hello"), []byte{bytesDescMarker, ^byte('h'), ^byte('e'), ^byte('l'), ^byte('l'), ^byte('o'), escapeDesc, escapedTermDesc}},
		{[]byte{'b', 0xff}, []byte{bytesDescMarker, ^byte('b'), 0x00, escapeDesc, escapedTermDesc}},
		{[]byte{'b', 0, 0, 'a'}, []byte{bytesDescMarker, ^byte('b'), 0xff, escaped00Desc, 0xff, escaped00Desc, ^byte('a'), escapeDesc, escapedTermDesc}},
		{[]byte{'b', 0, 0}, []byte{bytesDescMarker, ^byte('b'), 0xff, escaped00Desc, 0xff, escaped00Desc, escapeDesc, escapedTermDesc}},
		{[]byte{'b', 0}, []byte{bytesDescMarker, ^byte('b'), 0xff, escaped00Desc, escapeDesc, escapedTermDesc}},
		{[]byte{'b'}, []byte{bytesDescMarker, ^byte('b'), escapeDesc, escapedTermDesc}},
		{[]byte{'a'}, []byte{bytesDescMarker, ^byte('a'), escapeDesc, escapedTermDesc}},
		{[]byte{0, 0xff, 'a'}, []byte{bytesDescMarker, 0xff, escaped00Desc, 0x00, ^byte('a'), escapeDesc, escapedTermDesc}},
		{[]byte{0, 'a'}, []byte{bytesDescMarker, 0xff, escaped00Desc, ^byte('a'), escapeDesc, escapedTermDesc}},
		{[]byte{0, 1, 'a'}, []byte{bytesDescMarker, 0xff, escaped00Desc, ^byte(1), ^byte('a'), escapeDesc, escapedTermDesc}},
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
		name        string             // name printed with errors.
		buf         []byte             // buf contains an invalid uvarint to decode.
		expectedErr error              // expectedErr is the expected error.
		decode      func([]byte) error // decode is called with buf.
	}{
		{
			name:        "DecodeVarint, overflows int64",
			buf:         []byte{IntMax, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			expectedErr: ErrVarintOverflow,
			decode:      func(b []byte) error { _, _, err := DecodeVarintAscending(b); return err },
		},
		{
			name:        "Bytes, no marker",
			buf:         []byte{'a'},
			expectedErr: ErrMarkersNotFound,
			decode:      func(b []byte) error { _, _, err := DecodeBytesAscending(b, nil); return err },
		},
		{
			name:        "Bytes, no terminator",
			buf:         []byte{bytesMarker, 'a'},
			expectedErr: ErrTerminatorNotFound,
			decode:      func(b []byte) error { _, _, err := DecodeBytesAscending(b, nil); return err },
		},
		{
			name:        "Bytes, malformed escape",
			buf:         []byte{bytesMarker, 'a', 0x00},
			expectedErr: ErrMalformedEscape,
			decode:      func(b []byte) error { _, _, err := DecodeBytesAscending(b, nil); return err },
		},
		{
			name:        "Bytes, invalid escape 1",
			buf:         []byte{bytesMarker, 'a', 0x00, 0x00},
			expectedErr: ErrUnknownEscapeSequence,
			decode:      func(b []byte) error { _, _, err := DecodeBytesAscending(b, nil); return err },
		},
		{
			name:        "Bytes, invalid escape 2",
			buf:         []byte{bytesMarker, 'a', 0x00, 0x02},
			expectedErr: ErrUnknownEscapeSequence,
			decode:      func(b []byte) error { _, _, err := DecodeBytesAscending(b, nil); return err },
		},
		{
			name:        "BytesDescending, no marker",
			buf:         []byte{'a'},
			expectedErr: ErrMarkersNotFound,
			decode:      func(b []byte) error { _, _, err := DecodeBytesAscending(b, nil); return err },
		},
		{
			name:        "BytesDescending, no terminator",
			buf:         []byte{bytesDescMarker, ^byte('a')},
			expectedErr: ErrTerminatorNotFound,
			decode:      func(b []byte) error { _, _, err := DecodeBytesDescending(b, nil); return err },
		},
		{
			name:        "BytesDescending, malformed escape",
			buf:         []byte{bytesDescMarker, ^byte('a'), 0xff},
			expectedErr: ErrMalformedEscape,
			decode:      func(b []byte) error { _, _, err := DecodeBytesDescending(b, nil); return err },
		},
		{
			name:        "BytesDescending, invalid escape 1",
			buf:         []byte{bytesDescMarker, ^byte('a'), 0xff, 0xff},
			expectedErr: ErrUnknownEscapeSequence,
			decode:      func(b []byte) error { _, _, err := DecodeBytesDescending(b, nil); return err },
		},
		{
			name:        "BytesDescending, invalid escape 2",
			buf:         []byte{bytesDescMarker, ^byte('a'), 0xff, 0xfd},
			expectedErr: ErrUnknownEscapeSequence,
			decode:      func(b []byte) error { _, _, err := DecodeBytesDescending(b, nil); return err },
		},
	}
	for _, test := range tests {
		err := test.decode(test.buf)
		assert.ErrorIs(t, err, test.expectedErr)
	}
}
