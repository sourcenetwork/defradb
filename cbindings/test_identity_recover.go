// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cbindings

/*
#include <stdlib.h>
#include "defra_structs.h"
*/
import "C"
import (
	"testing"
	"unsafe"
)

// Note: This file deliberately avoids the _test.go naming suffix. If it were a _test.go file, `go test`
// would compile it in a separate test compilation unit, conflicting with the //export directives and
// func main() present in this package. The tests here are called via wrappers in identity_recover_test.go.

// TestIdentityRecover tests that a new identity can be created, exported, and recovered
func TestIdentityRecover(t *testing.T) {
	identResult := NewIdentity(nil)
	if identResult.status != 0 {
		t.Fatalf("NewIdentity failed: %s", C.GoString(identResult.error))
	}
	defer FreeIdentity(identResult.identityPtr)

	exportResult := ConvertAndFreeCResult(ExportIdentityPrivateKey(identResult.identityPtr))
	if exportResult.Status != 0 {
		t.Fatalf("ExportIdentityPrivateKey failed: %s", exportResult.Error)
	}
	originalHex := exportResult.Value

	hexKey := C.CString(originalHex)
	defer C.free(unsafe.Pointer(hexKey))
	recoveredResult := NewIdentityFromPrivateKey(hexKey)
	if recoveredResult.status != 0 {
		t.Fatalf("NewIdentityFromPrivateKey failed: %s", C.GoString(recoveredResult.error))
	}
	defer FreeIdentity(recoveredResult.identityPtr)

	reExportResult := ConvertAndFreeCResult(ExportIdentityPrivateKey(recoveredResult.identityPtr))
	if reExportResult.Status != 0 {
		t.Fatalf("re-export failed: %s", reExportResult.Error)
	}

	if originalHex != reExportResult.Value {
		t.Fatalf("round-trip failed:\n  original:  %s\n  recovered: %s", originalHex, reExportResult.Value)
	}
}

// Test_NewIdentityFromPrivateKey_InvalidHex tests that a new identity cannot be created from an invalid hex string
func TestNewIdentityFromPrivateKey_InvalidHex(t *testing.T) {
	invalidHex := C.CString("not-valid-hex")
	defer C.free(unsafe.Pointer(invalidHex))
	result := NewIdentityFromPrivateKey(invalidHex)
	if result.status == 0 {
		t.Fatal("expected error for invalid hex input, got success")
	}
}

// Test_NewIdentityFromPrivateKey_EmptyKey tests that a new identity cannot be created from an empty hex string
func TestNewIdentityFromPrivateKey_EmptyKey(t *testing.T) {
	emptyKey := C.CString("")
	defer C.free(unsafe.Pointer(emptyKey))
	result := NewIdentityFromPrivateKey(emptyKey)
	if result.status == 0 {
		t.Fatal("expected error for empty key input, got success")
	}
}

// Test_ExportIdentityPrivateKey_InvalidPointer tests that an error is returned when an invalid identity pointer is passed to ExportIdentityPrivateKey
func TestExportIdentityPrivateKey_InvalidPointer(t *testing.T) {
	result := ConvertAndFreeCResult(ExportIdentityPrivateKey(0))
	if result.Status == 0 {
		t.Fatal("expected error for nil pointer, got success")
	}
}
