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

package tests

func init() {
	goClient = false
	httpClient = false
	cliClient = false
	cClient = false
	jsClient = true
	// JavaScript networking stack is managed externally
	skipNetworkTests = true
	// Backup API is not suitable for browser environments
	skipBackupTests = true
}
