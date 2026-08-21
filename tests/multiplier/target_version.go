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

import (
	"golang.org/x/mod/semver"
)

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

// ResolveTargetVersion returns the release the named multiplier should run
// against for a test declaring supportedFrom, and whether that multiplier
// targets a version at all.
//
// A test needing a release newer than the default target runs against the
// release it needs, rather than being dropped. Running the oldest release a
// test supports still exercises the pairing most likely to have drifted from
// the current build.
//
// supportedFrom must be valid semver; callers validate it first.
func ResolveTargetVersion(name Name, supportedFrom string) (string, bool) {
	target, targetsVersion := TargetVersionFor(name)
	if !targetsVersion {
		return "", false
	}

	if supportedFrom != "" && semver.Compare(target, supportedFrom) < 0 {
		return supportedFrom, true
	}

	return target, true
}
