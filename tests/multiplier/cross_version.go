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
	"sync"

	"github.com/sourcenetwork/testo/multiplier"

	"github.com/sourcenetwork/defradb/tests/action"
)

func init() {
	multiplier.Register(&crossVersion{name: CrossVersionOldSource, oldNodeFirst: true})
	multiplier.Register(&crossVersion{name: CrossVersionNewSource, oldNodeFirst: false})
}

// CrossVersionTargetVersion is the older release run against the current build.
// v1.0.0 is the only post-1.0 release, so it is the only pair for now.
const CrossVersionTargetVersion = "v1.0.0"

// CrossVersionOldSource runs the first node on the older release, so data starts
// on the old node.
//
// Tests needing behaviour the older release lacks opt out with MultiplierExcludes.
// A version-aware gate is tracked in
// https://github.com/sourcenetwork/defradb/issues/5121
const CrossVersionOldSource Name = "cross-version-old-source"

// CrossVersionNewSource runs the last node on the older release, so data starts
// on the current build.
const CrossVersionNewSource Name = "cross-version-new-source"

// crossVersion runs one node of a networked test on an older release, so the
// existing P2P suite also checks compatibility with that release.
//
// Both directions are worth running because they fail differently: a new node
// sending to an old one relies on the old node ignoring fields it does not know,
// and an old node sending to a new one relies on the new node reading a missing
// field as a zero value. Each direction is registered separately because a
// multiplier applies one transformation per run.
type crossVersion struct {
	name Name
	// oldNodeFirst picks which node carries the older version. Node 0 is the
	// source in most of the suite, so this sets the direction.
	oldNodeFirst bool
}

// targetVersionOverride is the release the next Apply runs against, set by the
// harness for a test that needs a release newer than the default target.
//
// testo hands Apply only the action set, so the version a given test needs
// cannot be read there. Tests run sequentially within a package and the harness
// sets this immediately before Apply, so a package-level value is enough. It is
// guarded because the harness may run test packages in parallel.
var targetVersionOverride struct {
	sync.Mutex
	byName map[Name]string
}

// SetTargetVersion tells the named multiplier which release to run against for
// the next Apply. An empty version restores the default target.
//
// It returns a function that restores the previous value, so a test cannot leak
// its version into the next one.
func SetTargetVersion(name Name, version string) func() {
	targetVersionOverride.Lock()
	defer targetVersionOverride.Unlock()

	if targetVersionOverride.byName == nil {
		targetVersionOverride.byName = map[Name]string{}
	}

	previous := targetVersionOverride.byName[name]
	targetVersionOverride.byName[name] = version

	return func() {
		targetVersionOverride.Lock()
		defer targetVersionOverride.Unlock()
		targetVersionOverride.byName[name] = previous
	}
}

// TargetVersionInEffect returns the release the named multiplier currently runs
// against, which is [CrossVersionTargetVersion] unless [SetTargetVersion] has
// pointed it elsewhere.
func TargetVersionInEffect(name Name) string {
	targetVersionOverride.Lock()
	defer targetVersionOverride.Unlock()

	if v := targetVersionOverride.byName[name]; v != "" {
		return v
	}

	return CrossVersionTargetVersion
}

var _ Multiplier = (*crossVersion)(nil)
var _ multiplier.ActionAwareSkipper = (*crossVersion)(nil)

func (m *crossVersion) Name() Name {
	return m.name
}

// ShouldSkip implements [multiplier.ActionAwareSkipper].
//
// A test with one node has nothing to check compatibility against, and a test
// that already sets a version is checking something specific that this would
// overwrite.
func (m *crossVersion) ShouldSkip(actions action.Actions) bool {
	nodes := nodeActions(actions)
	if len(nodes) < 2 {
		return true
	}

	for _, node := range nodes {
		if node.Version != "" {
			return true
		}
	}

	return false
}

func (m *crossVersion) Apply(source action.Actions) action.Actions {
	nodes := nodeActions(source)
	if len(nodes) < 2 {
		return source
	}

	target := nodes[len(nodes)-1]
	if m.oldNodeFirst {
		target = nodes[0]
	}

	result := make(action.Actions, len(source))
	for i, a := range source {
		if cfg, ok := a.(*action.NewNode); ok && cfg == target {
			result[i] = cfg.WithVersion(TargetVersionInEffect(m.name))
			continue
		}
		result[i] = a
	}

	return result
}

// nodeActions returns the node creation actions in the action set, in order.
func nodeActions(actions action.Actions) []*action.NewNode {
	var configs []*action.NewNode
	for _, a := range actions {
		if cfg, ok := a.(*action.NewNode); ok {
			configs = append(configs, cfg)
		}
	}
	return configs
}

// MakesNodeExternal reports whether the named multiplier runs one of the nodes as a
// separate process.
//
// Such a node is reached over HTTP whatever the run-wide client type.
func MakesNodeExternal(name Name) bool {
	return name == CrossVersionOldSource || name == CrossVersionNewSource
}
