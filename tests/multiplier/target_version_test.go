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

func TestTargetVersionFor_CrossVersionMultipliers(t *testing.T) {
	for _, name := range []Name{CrossVersionOldSource, CrossVersionNewSource} {
		version, ok := TargetVersionFor(name)
		assert.True(t, ok, "%s should target a version", name)
		assert.Equal(t, CrossVersionTargetVersion, version)
	}
}

func TestTargetVersionFor_NonVersionMultipliers(t *testing.T) {
	// These multipliers say nothing about the release under test, so a test
	// declaring SupportedFromVersion must still run under them.
	for _, name := range []Name{SignedDocs, SecondaryIndex, EncryptedDocs} {
		version, ok := TargetVersionFor(name)
		assert.False(t, ok, "%s should not target a version", name)
		assert.Equal(t, "", version)
	}
}

func TestTargetVersionFor_UnknownName(t *testing.T) {
	version, ok := TargetVersionFor("no-such-multiplier")
	assert.False(t, ok)
	assert.Equal(t, "", version)
}

func TestTargetVersionFor_ReturnsComparableSemver(t *testing.T) {
	// The harness feeds this straight into semver.Compare, so a target that is not
	// valid semver would compare as older than everything and skip every gated test.
	for _, name := range []Name{CrossVersionOldSource, CrossVersionNewSource} {
		version, _ := TargetVersionFor(name)
		assert.True(t, semver.IsValid(version), "%s targets %q which is not valid semver", name, version)
	}
}
