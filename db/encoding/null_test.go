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

func TestEncodeDecodeNull(t *testing.T) {
	const hello = "hello"

	buf := EncodeNullAscending([]byte(hello))
	expected := []byte(hello + "\x00")
	if !bytes.Equal(expected, buf) {
		t.Fatalf("expected %q, but found %q", expected, buf)
	}

	if remaining, isNull := DecodeIfNull([]byte(hello)); isNull {
		t.Fatalf("expected isNull=false, but found isNull=%v", isNull)
	} else if hello != string(remaining) {
		t.Fatalf("expected %q, but found %q", hello, remaining)
	}

	if remaining, isNull := DecodeIfNull([]byte("\x00" + hello)); !isNull {
		t.Fatalf("expected isNull=true, but found isNull=%v", isNull)
	} else if hello != string(remaining) {
		t.Fatalf("expected %q, but found %q", hello, remaining)
	}
}
