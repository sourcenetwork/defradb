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
	"math"
	"testing"
)

func testBasicEncodeDecodeUint64(
	encFunc func([]byte, uint64) []byte,
	decFunc func([]byte) ([]byte, uint64, error),
	descending bool,
	t *testing.T,
) {
	testCases := []uint64{
		0, 1,
		1<<8 - 1, 1 << 8,
		1<<16 - 1, 1 << 16,
		1<<24 - 1, 1 << 24,
		1<<32 - 1, 1 << 32,
		1<<40 - 1, 1 << 40,
		1<<48 - 1, 1 << 48,
		1<<56 - 1, 1 << 56,
		math.MaxUint64 - 1, math.MaxUint64,
	}

	var lastEnc []byte
	for i, v := range testCases {
		enc := encFunc(nil, v)
		if i > 0 {
			if (descending && bytes.Compare(enc, lastEnc) >= 0) ||
				(!descending && bytes.Compare(enc, lastEnc) < 0) {
				t.Errorf("ordered constraint violated for %d: [% x] vs. [% x]", v, enc, lastEnc)
			}
		}
		b, decode, err := decFunc(enc)
		if err != nil {
			t.Error(err)
			continue
		}
		if len(b) != 0 {
			t.Errorf("leftover bytes: [% x]", b)
		}
		if decode != v {
			t.Errorf("decode yielded different value than input: %d vs. %d", decode, v)
		}
		lastEnc = enc
	}
}

var int64TestCases = [...]int64{
	math.MinInt64, math.MinInt64 + 1,
	-1<<56 - 1, -1 << 56,
	-1<<48 - 1, -1 << 48,
	-1<<40 - 1, -1 << 40,
	-1<<32 - 1, -1 << 32,
	-1<<24 - 1, -1 << 24,
	-1<<16 - 1, -1 << 16,
	-1<<8 - 1, -1 << 8,
	-1, 0, 1,
	1<<8 - 1, 1 << 8,
	1<<16 - 1, 1 << 16,
	1<<24 - 1, 1 << 24,
	1<<32 - 1, 1 << 32,
	1<<40 - 1, 1 << 40,
	1<<48 - 1, 1 << 48,
	1<<56 - 1, 1 << 56,
	math.MaxInt64 - 1, math.MaxInt64,
}

func testBasicEncodeDecodeInt64(
	encFunc func([]byte, int64) []byte,
	decFunc func([]byte) ([]byte, int64, error),
	descending bool,
	t *testing.T,
) {
	var lastEnc []byte
	for i, v := range int64TestCases {
		enc := encFunc(nil, v)
		if i > 0 {
			if (descending && bytes.Compare(enc, lastEnc) >= 0) ||
				(!descending && bytes.Compare(enc, lastEnc) < 0) {
				t.Errorf("ordered constraint violated for %d: [% x] vs. [% x]", v, enc, lastEnc)
			}
		}
		b, decode, err := decFunc(enc)
		if err != nil {
			t.Errorf("%v: %d [%x]", err, v, enc)
			continue
		}
		if len(b) != 0 {
			t.Errorf("leftover bytes: [% x]", b)
		}
		if decode != v {
			t.Errorf("decode yielded different value than input: %d vs. %d [%x]", decode, v, enc)
		}
		lastEnc = enc
	}
}

type testCaseInt64 struct {
	value  int64
	expEnc []byte
}

func testCustomEncodeInt64(
	testCases []testCaseInt64, encFunc func([]byte, int64) []byte, t *testing.T,
) {
	for _, test := range testCases {
		enc := encFunc(nil, test.value)
		if !bytes.Equal(enc, test.expEnc) {
			t.Errorf("expected [% x]; got [% x] (value: %d)", test.expEnc, enc, test.value)
		}
	}
}

type testCaseUint64 struct {
	value  uint64
	expEnc []byte
}

func testCustomEncodeUint64(
	testCases []testCaseUint64, encFunc func([]byte, uint64) []byte, t *testing.T,
) {
	for _, test := range testCases {
		enc := encFunc(nil, test.value)
		if !bytes.Equal(enc, test.expEnc) {
			t.Errorf("expected [% x]; got [% x] (value: %d)", test.expEnc, enc, test.value)
		}
	}
}

func TestEncodeDecodeUint64(t *testing.T) {
	testBasicEncodeDecodeUint64(EncodeUint64Ascending, DecodeUint64Ascending, false, t)
	testCases := []testCaseUint64{
		{0, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{1, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
		{1 << 8, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00}},
		{math.MaxUint64, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	}
	testCustomEncodeUint64(testCases, EncodeUint64Ascending, t)
}

func TestEncodeDecodeUint64Descending(t *testing.T) {
	testBasicEncodeDecodeUint64(EncodeUint64Descending, DecodeUint64Descending, true, t)
	testCases := []testCaseUint64{
		{0, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{1, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe}},
		{1 << 8, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe, 0xff}},
		{math.MaxUint64, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	}
	testCustomEncodeUint64(testCases, EncodeUint64Descending, t)
}

func TestEncodeDecodeVarint(t *testing.T) {
	testBasicEncodeDecodeInt64(EncodeVarintAscending, DecodeVarintAscending, false, t)
	testCases := []testCaseInt64{
		{math.MinInt64, []byte{0x80, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{math.MinInt64 + 1, []byte{0x80, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
		{-1 << 8, []byte{0x86, 0xff, 0x00}},
		{-1, []byte{0x87, 0xff}},
		{0, []byte{0x88}},
		{1, []byte{0x89}},
		{109, []byte{0xf5}},
		{112, []byte{0xf6, 0x70}},
		{1 << 8, []byte{0xf7, 0x01, 0x00}},
		{math.MaxInt64, []byte{0xfd, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	}
	testCustomEncodeInt64(testCases, EncodeVarintAscending, t)
}

func TestEncodeDecodeVarintDescending(t *testing.T) {
	testBasicEncodeDecodeInt64(EncodeVarintDescending, DecodeVarintDescending, true, t)
	testCases := []testCaseInt64{
		{math.MinInt64, []byte{0xfd, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{math.MinInt64 + 1, []byte{0xfd, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe}},
		{-1 << 8, []byte{0xf6, 0xff}},
		{-110, []byte{0xf5}},
		{-1, []byte{0x88}},
		{0, []byte{0x87, 0xff}},
		{1, []byte{0x87, 0xfe}},
		{1 << 8, []byte{0x86, 0xfe, 0xff}},
		{math.MaxInt64, []byte{0x80, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	}
	testCustomEncodeInt64(testCases, EncodeVarintDescending, t)
}

func TestEncodeDecodeUvarint(t *testing.T) {
	testBasicEncodeDecodeUint64(EncodeUvarintAscending, DecodeUvarintAscending, false, t)
	testCases := []testCaseUint64{
		{0, []byte{0x88}},
		{1, []byte{0x89}},
		{109, []byte{0xf5}},
		{110, []byte{0xf6, 0x6e}},
		{1 << 8, []byte{0xf7, 0x01, 0x00}},
		{math.MaxUint64, []byte{0xfd, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	}
	testCustomEncodeUint64(testCases, EncodeUvarintAscending, t)
}

func TestEncodeDecodeUvarintDescending(t *testing.T) {
	testBasicEncodeDecodeUint64(EncodeUvarintDescending, DecodeUvarintDescending, true, t)
	testCases := []testCaseUint64{
		{0, []byte{0x88}},
		{1, []byte{0x87, 0xfe}},
		{1 << 8, []byte{0x86, 0xfe, 0xff}},
		{math.MaxUint64 - 1, []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
		{math.MaxUint64, []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	}
	testCustomEncodeUint64(testCases, EncodeUvarintDescending, t)
}
