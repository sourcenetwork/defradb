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
	"github.com/stretchr/testify/require"

	m "github.com/sourcenetwork/testo/multiplier"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
)

func oldSource() *crossVersion {
	return &crossVersion{name: CrossVersionOldSource, oldNodeFirst: true}
}

func newSource() *crossVersion {
	return &crossVersion{name: CrossVersionNewSource, oldNodeFirst: false}
}

func TestCrossVersionNames_Stable(t *testing.T) {
	// The names are part of the CI contract, used in DEFRA_MULTIPLIERS and in
	// MultiplierExcludes. Changing them breaks workflow config and every test
	// that opts out.
	assert.Equal(t, "cross-version-old-source", string(CrossVersionOldSource))
	assert.Equal(t, "cross-version-new-source", string(CrossVersionNewSource))
	assert.Equal(t, CrossVersionOldSource, oldSource().Name())
	assert.Equal(t, CrossVersionNewSource, newSource().Name())
}

func TestCrossVersion_ImplementsInterfaces(t *testing.T) {
	var _ Multiplier = (*crossVersion)(nil)
	var _ m.Multiplier = (*crossVersion)(nil)
	var _ m.ActionAwareSkipper = (*crossVersion)(nil)
}

func TestCrossVersion_IsRegistered(t *testing.T) {
	m.Init("__cross_version_test_unset_env__", CrossVersionOldSource, CrossVersionNewSource)
	t.Cleanup(func() {
		m.Init("__cross_version_test_unset_env__")
	})

	active := m.Get()
	assert.Contains(t, active, string(CrossVersionOldSource))
	assert.Contains(t, active, string(CrossVersionNewSource))
}

func TestCrossVersionApply_WithEmptyActions_ReturnsEmpty(t *testing.T) {
	result := oldSource().Apply(action.Actions{})

	assert.Empty(t, result)
}

func TestCrossVersionApply_WithNilActions_ReturnsNil(t *testing.T) {
	result := oldSource().Apply(nil)

	assert.Nil(t, result)
}

func TestCrossVersionApply_OldSource_VersionsFirstNode(t *testing.T) {
	first := action.RandomNetworkingConfig()
	second := action.RandomNetworkingConfig()
	source := action.Actions{first, second}

	result := oldSource().Apply(source)

	require.Len(t, result, 2)
	assert.Equal(t, CrossVersionTargetVersion, result[0].(*action.NewNode).Version)
	assert.Equal(t, "", result[1].(*action.NewNode).Version)
}

func TestCrossVersionApply_NewSource_VersionsLastNode(t *testing.T) {
	first := action.RandomNetworkingConfig()
	second := action.RandomNetworkingConfig()
	source := action.Actions{first, second}

	result := newSource().Apply(source)

	require.Len(t, result, 2)
	assert.Equal(t, "", result[0].(*action.NewNode).Version)
	assert.Equal(t, CrossVersionTargetVersion, result[1].(*action.NewNode).Version)
}

func TestCrossVersionApply_WithThreeNodes_VersionsOnlyOne(t *testing.T) {
	source := action.Actions{
		action.RandomNetworkingConfig(),
		action.RandomNetworkingConfig(),
		action.RandomNetworkingConfig(),
	}

	result := newSource().Apply(source)

	require.Len(t, result, 3)
	assert.Equal(t, "", result[0].(*action.NewNode).Version)
	assert.Equal(t, "", result[1].(*action.NewNode).Version)
	assert.Equal(t, CrossVersionTargetVersion, result[2].(*action.NewNode).Version)
}

func TestCrossVersionApply_LeavesOtherActionsUntouched(t *testing.T) {
	add := &action.AddCollection{SDL: "type User { name: String }"}
	source := action.Actions{
		action.RandomNetworkingConfig(),
		add,
		action.RandomNetworkingConfig(),
	}

	result := oldSource().Apply(source)

	require.Len(t, result, 3)
	assert.Same(t, add, result[1], "non node-config actions must not be replaced")
}

func TestCrossVersionApply_DoesNotMutateSource(t *testing.T) {
	// Apply must not write through to the caller's config, otherwise a test run
	// would leak the version into the next multiplier or a later run.
	first := action.RandomNetworkingConfig()
	source := action.Actions{first, action.RandomNetworkingConfig()}

	oldSource().Apply(source)

	assert.Equal(t, "", first.Version, "the original config must be unchanged")
}

func TestCrossVersionApply_PreservesNetworkingConfig(t *testing.T) {
	source := action.Actions{
		action.RandomNetworkingConfig(),
		action.RandomNetworkingConfig(),
	}

	result := oldSource().Apply(source)

	versioned := result[0].(*action.NewNode)
	assert.NotNil(t, versioned.Network, "networking config must survive the rewrite")
}

func TestCrossVersionApply_WithSingleNode_ReturnsSourceUnchanged(t *testing.T) {
	source := action.Actions{action.RandomNetworkingConfig()}

	result := oldSource().Apply(source)

	assert.Equal(t, "", result[0].(*action.NewNode).Version)
}

func TestCrossVersionShouldSkip_WithSingleNode_Skips(t *testing.T) {
	// A single node has nothing to check compatibility against.
	actions := action.Actions{action.RandomNetworkingConfig()}

	assert.True(t, oldSource().ShouldSkip(actions))
}

func TestCrossVersionShouldSkip_WithNoNodes_Skips(t *testing.T) {
	actions := action.Actions{&action.AddCollection{SDL: "type User { name: String }"}}

	assert.True(t, oldSource().ShouldSkip(actions))
}

func TestCrossVersionShouldSkip_WithTwoNodes_DoesNotSkip(t *testing.T) {
	actions := action.Actions{
		action.RandomNetworkingConfig(),
		action.RandomNetworkingConfig(),
	}

	assert.False(t, oldSource().ShouldSkip(actions))
	assert.False(t, newSource().ShouldSkip(actions))
}

func TestCrossVersionShouldSkip_WithWritesNamingTheirNode_DoesNotSkip(t *testing.T) {
	actions := action.Actions{
		action.RandomNetworkingConfig(),
		action.RandomNetworkingConfig(),
		&action.AddDoc{NodeID: immutable.Some(0), Doc: `{"Name": "John"}`},
		&action.UpdateDoc{NodeID: immutable.Some(0), Doc: `{"Name": "Fred"}`},
	}

	assert.False(t, oldSource().ShouldSkip(actions))
	assert.False(t, newSource().ShouldSkip(actions))
}

func TestCrossVersionShouldSkip_WithWritesNotNamingTheirNode_DoesNotSkip(t *testing.T) {
	// A write with no node set lands on the versioned node too. That is still
	// worth running: what the node received is checked by querying it, so the
	// write does not need to say where it went.
	actions := action.Actions{
		action.RandomNetworkingConfig(),
		action.RandomNetworkingConfig(),
		&action.AddDoc{Doc: `{"Name": "John"}`},
		&action.UpdateDoc{Doc: `{"Name": "Fred"}`},
	}

	assert.False(t, oldSource().ShouldSkip(actions))
	assert.False(t, newSource().ShouldSkip(actions))
}

func TestCrossVersionShouldSkip_WithVersionAlreadySet_Skips(t *testing.T) {
	// The hand written cross version tests pin their own versions. Rewriting
	// them would test something other than what they were written to check.
	actions := action.Actions{
		action.RandomNetworkingConfig(),
		action.RandomNetworkingConfig().WithVersion("v1.0.0"),
	}

	assert.True(t, oldSource().ShouldSkip(actions))
}
