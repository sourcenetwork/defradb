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

package action

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	m "github.com/sourcenetwork/testo/multiplier"

	"github.com/sourcenetwork/defradb/client/options"
)

func TestNodeConfig_ImplementsAction(t *testing.T) {
	// Node config must be an action so that it survives the action.Action filter
	// the harness applies before handing the set to a multiplier. Anything that
	// is not an action is silently dropped there, which would leave a multiplier
	// targeting node config running green while changing nothing.
	var _ Action = (*NewNode)(nil)
	var _ Stateful = (*NewNode)(nil)

	var cfg any = &NewNode{}
	_, ok := cfg.(Action)
	assert.True(t, ok, "NewNode must implement Action")
}

func TestNodeConfig_SurvivesActionFilter(t *testing.T) {
	// Mirrors the filter in the integration harness' applyMultipliers: only
	// elements implementing action.Action reach the multiplier engine.
	testCaseActions := []any{
		RandomNetworkingConfig(),
		&AddCollection{SDL: "type User { name: String }"},
	}

	var actions Actions
	for _, a := range testCaseActions {
		if act, ok := a.(Action); ok {
			actions = append(actions, act)
		}
	}

	require.Len(t, actions, 2, "node config must not be dropped by the action filter")

	var found bool
	for _, a := range actions {
		if _, ok := a.(*NewNode); ok {
			found = true
		}
	}
	assert.True(t, found, "node config must be visible to multipliers")
}

func TestNodeConfig_ApplyCanRewriteVersion(t *testing.T) {
	// The reason node config became an action: a multiplier must be able to
	// rewrite it as plain data.
	source := Actions{
		RandomNetworkingConfig(),
		RandomNetworkingConfig(),
	}

	for _, a := range source {
		if cfg, ok := a.(*NewNode); ok {
			cfg.Version = "v1.0.0"
			break
		}
	}

	first, ok := source[0].(*NewNode)
	require.True(t, ok)
	second, ok := source[1].(*NewNode)
	require.True(t, ok)

	assert.Equal(t, "v1.0.0", first.Version)
	assert.Equal(t, "", second.Version, "only the targeted node should be rewritten")
}

func TestNodeConfigP2POptions_WithNilNetwork_ReturnsDefaults(t *testing.T) {
	cfg := &NewNode{}

	assert.Equal(t, options.NodeP2POptions{}, cfg.P2POptions())
}

func TestNodeConfigP2POptions_WithNetwork_ReturnsConfigured(t *testing.T) {
	cfg := &NewNode{
		Network: func() options.NodeP2POptions {
			return options.NodeP2POptions{EnablePubSub: true}
		},
	}

	assert.True(t, cfg.P2POptions().EnablePubSub)
}

func TestNodeConfigWithVersion_SetsVersion(t *testing.T) {
	cfg := RandomNetworkingConfig().WithVersion("v1.0.0")

	assert.Equal(t, "v1.0.0", cfg.Version)
	assert.NotNil(t, cfg.Network, "networking config must be preserved")
}

func TestNodeConfigWithVersion_DoesNotMutateReceiver(t *testing.T) {
	// WithVersion takes a pointer receiver, so it must copy rather than write
	// through to the shared value a constructor handed out.
	original := RandomNetworkingConfig()

	versioned := original.WithVersion("v1.0.0")

	assert.Equal(t, "", original.Version, "WithVersion must not mutate its receiver")
	assert.Equal(t, "v1.0.0", versioned.Version)
	assert.NotSame(t, original, versioned)
}

func TestRandomNetworkingConfig_ReturnsPointer(t *testing.T) {
	// The pointer return is what lets the value satisfy Action while leaving the
	// existing call sites, which only ever append the result, untouched.
	cfg := RandomNetworkingConfig()

	require.NotNil(t, cfg)
	var _ Action = cfg
}

func TestNodeConfig_ZeroValueIsNative(t *testing.T) {
	cfg := &NewNode{}

	assert.Equal(t, "", cfg.Version, "the zero value must describe a native, current-build node")
}

func TestNodeConfig_NotAnActionAwareSkipper(t *testing.T) {
	// Node config carries no skip logic of its own; skipping is decided by the
	// multipliers that rewrite it.
	var cfg any = &NewNode{}
	_, ok := cfg.(m.ActionAwareSkipper)
	assert.False(t, ok)
}
