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

package tests

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"golang.org/x/mod/semver"

	defraMultiplier "github.com/sourcenetwork/defradb/tests/multiplier"
)

// versionRecorder captures a failure from applyTestCaseVersion without ending
// the real test.
//
// Fatalf ends the calling goroutine, so runVersion runs the call in one of its
// own.
type versionRecorder struct {
	testing.TB
	failed  bool
	skipped bool
	message string
}

func (r *versionRecorder) Skipf(format string, args ...any) {
	r.skipped = true
	r.message = fmt.Sprintf(format, args...)
	runtime.Goexit()
}

func (r *versionRecorder) SkipNow() {
	r.skipped = true
	runtime.Goexit()
}

func (r *versionRecorder) Errorf(format string, args ...any) {
	r.failed = true
	r.message = fmt.Sprintf(format, args...)
}

func (r *versionRecorder) FailNow() {
	r.failed = true
	runtime.Goexit()
}

func (r *versionRecorder) Fatalf(format string, args ...any) {
	r.failed = true
	r.message = fmt.Sprintf(format, args...)
	runtime.Goexit()
}

func (r *versionRecorder) Helper() {}

// runVersion applies the version resolution and reports the version each
// multiplier ended up pointed at, leaving the defaults restored.
func runVersion(t *testing.T, supportedFrom string, activeNames string, names ...defraMultiplier.Name) (
	map[defraMultiplier.Name]string, *versionRecorder,
) {
	t.Helper()
	return runVersionMode(t, supportedFrom, activeNames, false, names...)
}

// runVersionExact is [runVersion] with exact mode on.
func runVersionExact(t *testing.T, supportedFrom string, activeNames string, names ...defraMultiplier.Name) (
	map[defraMultiplier.Name]string, *versionRecorder,
) {
	t.Helper()
	return runVersionMode(t, supportedFrom, activeNames, true, names...)
}

func runVersionMode(
	t *testing.T,
	supportedFrom string,
	activeNames string,
	exact bool,
	names ...defraMultiplier.Name,
) (map[defraMultiplier.Name]string, *versionRecorder) {
	t.Helper()

	rec := &versionRecorder{TB: t}
	resolved := map[defraMultiplier.Name]string{}

	done := make(chan struct{})
	go func() {
		defer close(done)

		restore := applyTestCaseVersion(rec, supportedFrom, activeNames, exact)
		defer restore()

		for _, name := range names {
			resolved[name] = defraMultiplier.TargetVersionInEffect(name)
		}
	}()
	<-done

	return resolved, rec
}

func TestApplyTestCaseVersion_NoSupportedFrom_UsesDefaultTarget(t *testing.T) {
	resolved, rec := runVersion(t, "", defraMultiplier.CrossVersionOldSource,
		defraMultiplier.CrossVersionOldSource)

	require.False(t, rec.failed)
	assert.Equal(t, defraMultiplier.CrossVersionTargetVersion,
		resolved[defraMultiplier.CrossVersionOldSource])
}

func TestApplyTestCaseVersion_SupportedFromNewerThanTarget_RunsAtSupportedFrom(t *testing.T) {
	// The point of the change: rather than skipping, the test runs against the
	// oldest release it supports.
	resolved, rec := runVersion(t, "v99.0.0", defraMultiplier.CrossVersionOldSource,
		defraMultiplier.CrossVersionOldSource)

	require.False(t, rec.failed)
	assert.Equal(t, "v99.0.0", resolved[defraMultiplier.CrossVersionOldSource])
}

func TestApplyTestCaseVersion_SupportedFromEqualToTarget_UsesTarget(t *testing.T) {
	resolved, rec := runVersion(t, defraMultiplier.CrossVersionTargetVersion,
		defraMultiplier.CrossVersionOldSource, defraMultiplier.CrossVersionOldSource)

	require.False(t, rec.failed)
	assert.Equal(t, defraMultiplier.CrossVersionTargetVersion,
		resolved[defraMultiplier.CrossVersionOldSource])
}

func TestApplyTestCaseVersion_SupportedFromOlderThanTarget_UsesTarget(t *testing.T) {
	// The default target already satisfies the test, so it stays: it is the
	// pairing most likely to have drifted from the current build.
	require.Equal(t, "v1.0.0", defraMultiplier.CrossVersionTargetVersion)

	resolved, rec := runVersion(t, "v0.9.0", defraMultiplier.CrossVersionOldSource,
		defraMultiplier.CrossVersionOldSource)

	require.False(t, rec.failed)
	assert.Equal(t, "v1.0.0", resolved[defraMultiplier.CrossVersionOldSource])
}

func TestApplyTestCaseVersion_BothDirections_AreSetIndependently(t *testing.T) {
	active := fmt.Sprintf("%s,%s",
		defraMultiplier.CrossVersionOldSource, defraMultiplier.CrossVersionNewSource)

	resolved, rec := runVersion(t, "v99.0.0", active,
		defraMultiplier.CrossVersionOldSource, defraMultiplier.CrossVersionNewSource)

	require.False(t, rec.failed)
	assert.Equal(t, "v99.0.0", resolved[defraMultiplier.CrossVersionOldSource])
	assert.Equal(t, "v99.0.0", resolved[defraMultiplier.CrossVersionNewSource])
}

func TestApplyTestCaseVersion_RestoresDefaultAfterwards(t *testing.T) {
	// A version must not leak into the next test in the package.
	_, rec := runVersion(t, "v99.0.0", defraMultiplier.CrossVersionOldSource,
		defraMultiplier.CrossVersionOldSource)
	require.False(t, rec.failed)

	assert.Equal(t, defraMultiplier.CrossVersionTargetVersion,
		defraMultiplier.TargetVersionInEffect(defraMultiplier.CrossVersionOldSource))
}

func TestApplyTestCaseVersion_NonVersionMultiplier_Unaffected(t *testing.T) {
	// signed-docs targets no release, so a declared version must not reach it and
	// must not disturb the cross-version default.
	_, rec := runVersion(t, "v99.0.0", defraMultiplier.SignedDocs)

	require.False(t, rec.failed)
	assert.Equal(t, defraMultiplier.CrossVersionTargetVersion,
		defraMultiplier.TargetVersionInEffect(defraMultiplier.CrossVersionOldSource))
}

func TestApplyTestCaseVersion_InvalidVersion_Fails(t *testing.T) {
	// A missing leading v would compare as older than every target and silently
	// leave the default in place, running against a release that cannot support
	// the test.
	for _, invalid := range []string{"1.1.0", "latest", "v", "vX.Y.Z"} {
		t.Run(invalid, func(t *testing.T) {
			_, rec := runVersion(t, invalid, defraMultiplier.CrossVersionOldSource)

			assert.True(t, rec.failed, "expected %q to fail the test", invalid)
			assert.Contains(t, rec.message, "SupportedFromVersion")
		})
	}
}

func TestApplyTestCaseVersion_InvalidVersionCheckedWithoutVersionMultiplier(t *testing.T) {
	// The value is wrong whether or not a version-targeting multiplier is active.
	_, rec := runVersion(t, "1.1.0", defraMultiplier.SignedDocs)

	assert.True(t, rec.failed)
}

func TestApplyTestCaseVersion_SpacedNames_AreTrimmed(t *testing.T) {
	resolved, rec := runVersion(t, "v99.0.0", " signed-docs , cross-version-old-source ",
		defraMultiplier.CrossVersionOldSource)

	require.False(t, rec.failed)
	assert.Equal(t, "v99.0.0", resolved[defraMultiplier.CrossVersionOldSource])
}

// TestSupportedFromVersion_ComparisonIsNotLexical guards the reason x/mod/semver
// is used rather than plain string comparison.
func TestSupportedFromVersion_ComparisonIsNotLexical(t *testing.T) {
	// "v1.10.0" sorts before "v1.9.0" as a string, so a lexical comparison would
	// treat a v1.10.0 target as older than a v1.9.0 requirement and needlessly
	// move the run to v1.9.0.
	require.Less(t, "v1.10.0", "v1.9.0", "precondition: these compare the wrong way lexically")

	assert.Positive(t, semver.Compare("v1.10.0", "v1.9.0"))
	assert.Negative(t, semver.Compare("v1.9.0", "v1.10.0"))
	assert.Zero(t, semver.Compare("v1.9.0", "v1.9.0"))
}

func TestApplyTestCaseVersion_Exact_SupportedFromNewerThanTarget_Skips(t *testing.T) {
	// In exact mode the release the test needs is covered by its own run, so
	// promoting it here would run it twice and report coverage of a release this
	// run never touched.
	_, rec := runVersionExact(t, "v99.0.0", defraMultiplier.CrossVersionOldSource,
		defraMultiplier.CrossVersionOldSource)

	assert.True(t, rec.skipped)
	assert.False(t, rec.failed)
	assert.Contains(t, rec.message, defraMultiplier.CrossVersionOldSource)
	assert.Contains(t, rec.message, defraMultiplier.CrossVersionTargetVersion)
	assert.Contains(t, rec.message, "v99.0.0")
}

func TestApplyTestCaseVersion_Exact_SupportedFromEqualToTarget_Runs(t *testing.T) {
	resolved, rec := runVersionExact(t, defraMultiplier.CrossVersionTargetVersion,
		defraMultiplier.CrossVersionOldSource, defraMultiplier.CrossVersionOldSource)

	require.False(t, rec.skipped)
	require.False(t, rec.failed)
	assert.Equal(t, defraMultiplier.CrossVersionTargetVersion,
		resolved[defraMultiplier.CrossVersionOldSource])
}

func TestApplyTestCaseVersion_Exact_SupportedFromOlderThanTarget_Runs(t *testing.T) {
	resolved, rec := runVersionExact(t, "v0.9.0", defraMultiplier.CrossVersionOldSource,
		defraMultiplier.CrossVersionOldSource)

	require.False(t, rec.skipped)
	require.False(t, rec.failed)
	assert.Equal(t, defraMultiplier.CrossVersionTargetVersion,
		resolved[defraMultiplier.CrossVersionOldSource])
}

func TestApplyTestCaseVersion_Exact_NoSupportedFrom_Runs(t *testing.T) {
	resolved, rec := runVersionExact(t, "", defraMultiplier.CrossVersionOldSource,
		defraMultiplier.CrossVersionOldSource)

	require.False(t, rec.skipped)
	assert.Equal(t, defraMultiplier.CrossVersionTargetVersion,
		resolved[defraMultiplier.CrossVersionOldSource])
}

func TestApplyTestCaseVersion_Exact_SkipRestoresEarlierMultipliers(t *testing.T) {
	// The skip ends the test partway through the loop, so a version set for an
	// earlier multiplier must not leak into the next test.
	active := fmt.Sprintf("%s,%s",
		defraMultiplier.CrossVersionOldSource, defraMultiplier.CrossVersionNewSource)

	_, rec := runVersionExact(t, "v99.0.0", active)
	require.True(t, rec.skipped)

	assert.Equal(t, defraMultiplier.CrossVersionTargetVersion,
		defraMultiplier.TargetVersionInEffect(defraMultiplier.CrossVersionOldSource))
	assert.Equal(t, defraMultiplier.CrossVersionTargetVersion,
		defraMultiplier.TargetVersionInEffect(defraMultiplier.CrossVersionNewSource))
}
