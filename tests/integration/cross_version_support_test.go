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

	defraMultiplier "github.com/sourcenetwork/defradb/tests/multiplier"
)

// versionRecorder captures a skip or failure from skipUnsupportedDirection
// without ending the real test.
//
// Skipf and Fatalf end the calling goroutine, so runDirection runs the call in
// one of its own.
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

// runDirection runs skipUnsupportedDirection against the test case and reports
// whether it skipped or failed.
func runDirection(t *testing.T, testCase *TestCase, activeNames string) *versionRecorder {
	t.Helper()

	rec := &versionRecorder{TB: t}
	done := make(chan struct{})
	go func() {
		defer close(done)
		skipUnsupportedDirection(rec, testCase, activeNames)
	}()
	<-done
	return rec
}

func TestSkipUnsupportedDirection_OldSourceNeedsNewer_SkipsOldSource(t *testing.T) {
	rec := runDirection(t, &TestCase{OldSourceSupportedFromVersion: "v99.0.0"},
		defraMultiplier.CrossVersionOldSource)

	assert.True(t, rec.skipped)
	assert.False(t, rec.failed)
	assert.Contains(t, rec.message, defraMultiplier.CrossVersionOldSource)
	assert.Contains(t, rec.message, "v99.0.0")
}

func TestSkipUnsupportedDirection_OldSourceNeedsNewer_RunsNewSource(t *testing.T) {
	// The other direction is unaffected, which is the point of declaring it per
	// direction.
	rec := runDirection(t, &TestCase{OldSourceSupportedFromVersion: "v99.0.0"},
		defraMultiplier.CrossVersionNewSource)

	assert.False(t, rec.skipped)
	assert.False(t, rec.failed)
}

func TestSkipUnsupportedDirection_NewSourceNeedsNewer_SkipsNewSource(t *testing.T) {
	rec := runDirection(t, &TestCase{NewSourceSupportedFromVersion: "v99.0.0"},
		defraMultiplier.CrossVersionNewSource)

	assert.True(t, rec.skipped)
	assert.Contains(t, rec.message, defraMultiplier.CrossVersionNewSource)
}

func TestSkipUnsupportedDirection_TargetSupported_Runs(t *testing.T) {
	target := defraMultiplier.TargetVersion(defraMultiplier.CrossVersionOldSource)
	rec := runDirection(t, &TestCase{OldSourceSupportedFromVersion: target},
		defraMultiplier.CrossVersionOldSource)

	assert.False(t, rec.skipped)
	assert.False(t, rec.failed)
}

func TestSkipUnsupportedDirection_InvalidVersion_FailsEvenWhenInactive(t *testing.T) {
	// A typo must not go unnoticed just because the run does not use that
	// direction.
	rec := runDirection(t, &TestCase{NewSourceSupportedFromVersion: "1.2.0"},
		defraMultiplier.SignedDocs)

	assert.True(t, rec.failed)
	assert.False(t, rec.skipped)
}
