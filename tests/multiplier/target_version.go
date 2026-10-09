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
	"fmt"
	"os"

	"golang.org/x/mod/semver"
)

// crossVersionTargetEnvName names the env var that picks the older release the
// cross-version multipliers run against, so CI can run one job per release.
const crossVersionTargetEnvName = "DEFRA_CROSS_VERSION_TARGET"

// crossVersionTarget is the older release the cross-version multipliers run
// against.
var crossVersionTarget = mustParseCrossVersionTarget(os.Getenv(crossVersionTargetEnvName))

// mustParseCrossVersionTarget returns the release named by value, or
// CrossVersionTargetVersion if value is empty. It panics on a malformed value,
// since every version comparison would then be wrong.
func mustParseCrossVersionTarget(value string) string {
	version, err := parseCrossVersionTarget(value)
	if err != nil {
		panic(err)
	}
	return version
}

func parseCrossVersionTarget(value string) (string, error) {
	if value == "" {
		return CrossVersionTargetVersion, nil
	}
	if !semver.IsValid(value) || semver.Canonical(value) != value {
		return "", fmt.Errorf("%s must be a release tag like v1.1.0, got %q", crossVersionTargetEnvName, value)
	}
	return value, nil
}

// TargetVersion returns the release the named multiplier runs against, or an
// empty string if it targets none.
func TargetVersion(name Name) string {
	switch name {
	case CrossVersionOldFirst, CrossVersionOldLast:
		return crossVersionTarget
	default:
		return ""
	}
}
