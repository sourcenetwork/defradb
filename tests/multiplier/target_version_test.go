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
	"testing"

	"github.com/stretchr/testify/assert"

	"golang.org/x/mod/semver"
)

func TestTargetVersion_CrossVersionMultipliers(t *testing.T) {
	for _, name := range []Name{CrossVersionOldFirst, CrossVersionOldLast} {
		assert.Equal(t, crossVersionTarget, TargetVersion(name))
	}
}

func TestParseCrossVersionTarget_Empty_UsesDefault(t *testing.T) {
	version, err := parseCrossVersionTarget("")
	assert.NoError(t, err)
	assert.Equal(t, CrossVersionTargetVersion, version)
}

func TestParseCrossVersionTarget_ReleaseTag_UsesIt(t *testing.T) {
	version, err := parseCrossVersionTarget("v1.1.0")
	assert.NoError(t, err)
	assert.Equal(t, "v1.1.0", version)
}

func TestParseCrossVersionTarget_Malformed_Errors(t *testing.T) {
	// A value that is not a full release tag would compare wrongly against the
	// release a test declares it needs.
	for _, value := range []string{"1.1.0", "v1.1", "latest"} {
		_, err := parseCrossVersionTarget(value)
		assert.Error(t, err, "%q should be rejected", value)
	}
}

func TestTargetVersion_NonVersionMultipliers(t *testing.T) {
	// These multipliers say nothing about the release under test, so a test
	// declaring the release it needs must still run under them.
	for _, name := range []Name{SignedDocs, SecondaryIndex, EncryptedDocs} {
		assert.Equal(t, "", TargetVersion(name), "%s should not target a version", name)
	}
}

func TestTargetVersion_UnknownName(t *testing.T) {
	assert.Equal(t, "", TargetVersion("no-such-multiplier"))
}

func TestTargetVersion_ReturnsComparableSemver(t *testing.T) {
	// The harness feeds this straight into semver.Compare, so a target that is not
	// valid semver would compare as older than everything and skip every gated test.
	for _, name := range []Name{CrossVersionOldFirst, CrossVersionOldLast} {
		version := TargetVersion(name)
		assert.True(t, semver.IsValid(version), "%s targets %q which is not valid semver", name, version)
	}
}
