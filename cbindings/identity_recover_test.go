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

import "testing"

// This file wraps the unit tests from test_identity_recover.go
// The purpose is to allow the unit tests to be run by `go test`, which would otherwise have
// problems due to the CGO directives used by the tests.

func Test_IdentityRecover(t *testing.T) {
	TestIdentityRecover(t)
}

func Test_NewIdentityFromPrivateKey_InvalidHex(t *testing.T) {
	TestNewIdentityFromPrivateKey_InvalidHex(t)
}

func Test_NewIdentityFromPrivateKey_EmptyKey(t *testing.T) {
	TestNewIdentityFromPrivateKey_EmptyKey(t)
}

func Test_ExportIdentityPrivateKey_InvalidPointer(t *testing.T) {
	TestExportIdentityPrivateKey_InvalidPointer(t)
}
