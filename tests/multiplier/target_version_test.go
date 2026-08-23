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

	"github.com/sourcenetwork/defradb/tests/action"
)

// twoNodeActions returns the minimal action set the cross-version multiplier
// acts on.
func twoNodeActions() action.Actions {
	return action.Actions{
		action.RandomNetworkingConfig(),
		action.RandomNetworkingConfig(),
	}
}

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

func TestResolveTargetVersion_NoRequirement_UsesDefaultTarget(t *testing.T) {
	version, resolution := ResolveTargetVersion(CrossVersionOldSource, "", false)

	assert.Equal(t, VersionRun, resolution)
	assert.Equal(t, CrossVersionTargetVersion, version)
}

func TestResolveTargetVersion_RequirementNewerThanTarget_UsesRequirement(t *testing.T) {
	version, resolution := ResolveTargetVersion(CrossVersionOldSource, "v99.0.0", false)

	assert.Equal(t, VersionRun, resolution)
	assert.Equal(t, "v99.0.0", version)
}

func TestResolveTargetVersion_RequirementOlderThanTarget_UsesDefaultTarget(t *testing.T) {
	// The default target already supports the test, and it is the pairing most
	// likely to have drifted from the current build.
	version, resolution := ResolveTargetVersion(CrossVersionOldSource, "v0.9.0", false)

	assert.Equal(t, VersionRun, resolution)
	assert.Equal(t, CrossVersionTargetVersion, version)
}

func TestResolveTargetVersion_RequirementEqualToTarget_UsesTarget(t *testing.T) {
	version, resolution := ResolveTargetVersion(CrossVersionOldSource, CrossVersionTargetVersion, false)

	assert.Equal(t, VersionRun, resolution)
	assert.Equal(t, CrossVersionTargetVersion, version)
}

func TestResolveTargetVersion_NonVersionMultiplier_TargetsNothing(t *testing.T) {
	version, resolution := ResolveTargetVersion(SignedDocs, "v99.0.0", false)

	assert.Equal(t, VersionNotTargeted, resolution)
	assert.Equal(t, "", version)
}

func TestResolveTargetVersion_Exact_RequirementNewerThanTarget_Skips(t *testing.T) {
	version, resolution := ResolveTargetVersion(CrossVersionOldSource, "v99.0.0", true)

	assert.Equal(t, VersionSkip, resolution)
	assert.Equal(t, "", version)
}

func TestResolveTargetVersion_Exact_RequirementTargetOrOlder_Runs(t *testing.T) {
	for _, supportedFrom := range []string{"", "v0.9.0", CrossVersionTargetVersion} {
		version, resolution := ResolveTargetVersion(CrossVersionOldSource, supportedFrom, true)

		assert.Equal(t, VersionRun, resolution, "supportedFrom %q", supportedFrom)
		assert.Equal(t, CrossVersionTargetVersion, version)
	}
}

func TestResolveTargetVersion_Exact_NonVersionMultiplier_TargetsNothing(t *testing.T) {
	version, resolution := ResolveTargetVersion(SignedDocs, "v99.0.0", true)

	assert.Equal(t, VersionNotTargeted, resolution)
	assert.Equal(t, "", version)
}

func TestSetTargetVersion_AppliesToNodeAndRestores(t *testing.T) {
	// The whole point of the override: Apply stamps the requested release rather
	// than the default target.
	m := oldSource()

	restore := SetTargetVersion(CrossVersionOldSource, "v99.0.0")
	applied := m.Apply(twoNodeActions())
	assert.Equal(t, "v99.0.0", nodeActions(applied)[0].Version)

	restore()
	applied = m.Apply(twoNodeActions())
	assert.Equal(t, CrossVersionTargetVersion, nodeActions(applied)[0].Version)
}

func TestSetTargetVersion_EmptyVersion_KeepsDefaultTarget(t *testing.T) {
	restore := SetTargetVersion(CrossVersionOldSource, "")
	defer restore()

	assert.Equal(t, CrossVersionTargetVersion, TargetVersionInEffect(CrossVersionOldSource))
}

func TestSetTargetVersion_IsPerMultiplier(t *testing.T) {
	restore := SetTargetVersion(CrossVersionOldSource, "v99.0.0")
	defer restore()

	assert.Equal(t, "v99.0.0", TargetVersionInEffect(CrossVersionOldSource))
	assert.Equal(t, CrossVersionTargetVersion, TargetVersionInEffect(CrossVersionNewSource))
}

func TestSetTargetVersion_NestedRestoreReturnsPreviousValue(t *testing.T) {
	outer := SetTargetVersion(CrossVersionOldSource, "v98.0.0")
	defer outer()

	inner := SetTargetVersion(CrossVersionOldSource, "v99.0.0")
	assert.Equal(t, "v99.0.0", TargetVersionInEffect(CrossVersionOldSource))

	inner()
	assert.Equal(t, "v98.0.0", TargetVersionInEffect(CrossVersionOldSource))
}
