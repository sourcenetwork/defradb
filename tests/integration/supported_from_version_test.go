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

// gateRecorder captures what skipUnsupportedVersion did without ending the real test.
//
// Skipf and Fatalf both end the calling goroutine, so runGate runs the call in one
// of its own.
type gateRecorder struct {
	testing.TB
	skipped bool
	failed  bool
	message string
}

func (r *gateRecorder) Skipf(format string, args ...any) {
	r.skipped = true
	r.message = fmt.Sprintf(format, args...)
	r.SkipNow()
}

func (r *gateRecorder) SkipNow() {
	runtime.Goexit()
}

func (r *gateRecorder) Errorf(format string, args ...any) {
	r.failed = true
	r.message = fmt.Sprintf(format, args...)
}

func (r *gateRecorder) FailNow() {
	r.failed = true
	runtime.Goexit()
}

func (r *gateRecorder) Fatalf(format string, args ...any) {
	r.failed = true
	r.message = fmt.Sprintf(format, args...)
	runtime.Goexit()
}

func (r *gateRecorder) Helper() {}

// runGate calls the gate and reports what it decided.
func runGate(t *testing.T, supportedFrom string, activeNames string) *gateRecorder {
	t.Helper()

	rec := &gateRecorder{TB: t}
	done := make(chan struct{})
	go func() {
		defer close(done)
		skipUnsupportedVersion(rec, supportedFrom, activeNames)
	}()
	<-done

	return rec
}

func TestSkipUnsupportedVersion_TargetOlderThanRequired_Skips(t *testing.T) {
	rec := runGate(t, "v99.0.0", defraMultiplier.CrossVersionOldSource)

	assert.True(t, rec.skipped)
	assert.False(t, rec.failed)
	assert.Contains(t, rec.message, defraMultiplier.CrossVersionOldSource)
	assert.Contains(t, rec.message, defraMultiplier.CrossVersionTargetVersion)
	assert.Contains(t, rec.message, "v99.0.0")
}

func TestSkipUnsupportedVersion_TargetEqualToRequired_Runs(t *testing.T) {
	rec := runGate(t, defraMultiplier.CrossVersionTargetVersion, defraMultiplier.CrossVersionOldSource)

	assert.False(t, rec.skipped)
	assert.False(t, rec.failed)
}

func TestSkipUnsupportedVersion_TargetNewerThanRequired_Runs(t *testing.T) {
	require.Equal(t, "v1.0.0", defraMultiplier.CrossVersionTargetVersion,
		"this test assumes the target is newer than v0.9.0")

	rec := runGate(t, "v0.9.0", defraMultiplier.CrossVersionOldSource)

	assert.False(t, rec.skipped)
	assert.False(t, rec.failed)
}

func TestSkipUnsupportedVersion_Empty_Runs(t *testing.T) {
	rec := runGate(t, "", defraMultiplier.CrossVersionOldSource)

	assert.False(t, rec.skipped)
	assert.False(t, rec.failed)
}

func TestSkipUnsupportedVersion_NonVersionMultiplier_Runs(t *testing.T) {
	// signed-docs says nothing about the release under test, so even a version the
	// current build does not have must not gate on it.
	rec := runGate(t, "v99.0.0", defraMultiplier.SignedDocs)

	assert.False(t, rec.skipped)
	assert.False(t, rec.failed)
}

func TestSkipUnsupportedVersion_NoActiveMultipliers_Runs(t *testing.T) {
	rec := runGate(t, "v99.0.0", "")

	assert.False(t, rec.skipped)
	assert.False(t, rec.failed)
}

func TestSkipUnsupportedVersion_InvalidVersion_Fails(t *testing.T) {
	// A missing leading v would compare as older than every target and silently
	// stop gating, so it fails the test instead.
	for _, invalid := range []string{"1.1.0", "latest", "v", "vX.Y.Z"} {
		t.Run(invalid, func(t *testing.T) {
			rec := runGate(t, invalid, defraMultiplier.CrossVersionOldSource)

			assert.True(t, rec.failed, "expected %q to fail the test", invalid)
			assert.False(t, rec.skipped)
			assert.Contains(t, rec.message, "SupportedFromVersion")
		})
	}
}

func TestSkipUnsupportedVersion_InvalidVersionCheckedBeforeMultipliers(t *testing.T) {
	// The value is wrong whether or not a version-targeting multiplier is active,
	// so a run with none must still report it.
	rec := runGate(t, "1.1.0", defraMultiplier.SignedDocs)

	assert.True(t, rec.failed)
	assert.False(t, rec.skipped)
}

func TestSkipUnsupportedVersion_SeveralMultipliers_SkipsIfAnyTargetIsTooOld(t *testing.T) {
	active := fmt.Sprintf("%s,%s", defraMultiplier.SignedDocs, defraMultiplier.CrossVersionNewSource)

	rec := runGate(t, "v99.0.0", active)

	assert.True(t, rec.skipped)
	assert.Contains(t, rec.message, defraMultiplier.CrossVersionNewSource)
}

func TestSkipUnsupportedVersion_SpacedNames_AreTrimmed(t *testing.T) {
	// multiplier.Get joins on a comma, but DEFRA_MULTIPLIERS is user-written and
	// testo trims its own copy, so the harness must not depend on the spacing.
	rec := runGate(t, "v99.0.0", " signed-docs , cross-version-old-source ")

	assert.True(t, rec.skipped)
}

// TestSupportedFromVersion_ComparisonIsNotLexical guards the reason x/mod/semver
// is used rather than plain string comparison.
//
// The gate itself can only be given the one real target, so the ordering the gate
// relies on is checked directly.
func TestSupportedFromVersion_ComparisonIsNotLexical(t *testing.T) {
	// "v1.10.0" sorts before "v1.9.0" as a string, so a lexical gate would decide a
	// v1.10.0 target is too old for a test needing v1.9.0 and skip it forever.
	require.Less(t, "v1.10.0", "v1.9.0", "precondition: these compare the wrong way lexically")

	assert.Positive(t, semver.Compare("v1.10.0", "v1.9.0"),
		"v1.10.0 must compare as newer than v1.9.0")
	assert.Negative(t, semver.Compare("v1.9.0", "v1.10.0"))
	assert.Zero(t, semver.Compare("v1.9.0", "v1.9.0"))
}
