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

package multiplier

// TargetVersionFor returns the release a version-targeting multiplier runs
// against, and whether the named multiplier targets one at all.
//
// The name to version mapping lives here rather than in the harness so adding a
// version pair does not require touching the harness.
func TargetVersionFor(name Name) (string, bool) {
	switch name {
	case CrossVersionOldSource, CrossVersionNewSource:
		return CrossVersionTargetVersion, true
	default:
		return "", false
	}
}
