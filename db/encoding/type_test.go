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
