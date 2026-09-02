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

// VersionResolution says what a version-targeting multiplier should do with a
// test.
type VersionResolution int

const (
	// VersionNotTargeted means the multiplier does not target a release, so the
	// test is unaffected by it.
	VersionNotTargeted VersionResolution = iota

	// VersionRun means the test should run against the returned release.
	VersionRun

	// VersionSkip means the test needs a newer release than this multiplier
	// targets and cannot run under it.
	VersionSkip
)

// ResolveTargetVersion decides what the named multiplier should do with a test
// declaring supportedFrom, and which release it should run against.
//
// When exact is false, a test needing a release newer than the default target
// runs against the release it needs rather than being dropped. This keeps a test
// that names a version contributing to compatibility coverage when only one
// release is being tested.
//
// When exact is true, the multiplier runs only the release it targets and skips
// anything needing newer. This is for runs that test several releases, where
// promoting a test to a newer release would duplicate the run that release
// already gets, and would report coverage of a release the test never touched.
//
// supportedFrom must be valid semver; callers validate it first.
func ResolveTargetVersion(name Name, supportedFrom string, exact bool) (string, VersionResolution) {
	target, targetsVersion := TargetVersionFor(name)
	if !targetsVersion {
		return "", VersionNotTargeted
	}

	// Compared as semver rather than as strings: "v1.10.0" sorts before "v1.9.0"
	// lexically, which would read a newer target as older than the requirement.
	if supportedFrom != "" && semver.Compare(target, supportedFrom) < 0 {
		if exact {
			return "", VersionSkip
		}
		return supportedFrom, VersionRun
	}

	return target, VersionRun
}
