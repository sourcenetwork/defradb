// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

//go:build !js

package tests

func init() {
	if !goClient && !httpClient && !cliClient && !cClient {
		// Default is to test go client type.
		goClient = true
	}
	if cClient {
		skipNetworkTests = false
		skipBackupTests = true
	}
}
