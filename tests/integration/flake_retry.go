// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tests

import (
	"fmt"
	"testing"
)

// flakeFailure is a sentinel value used to signal that a flaky test attempt
// has failed. It is panicked by flakeTestCapture's FailNow/Fatal/Fatalf methods
// and recovered by runTestWithRetry. This avoids the need for child goroutines
// that would be required if we used runtime.Goexit() (which terminates the
// current goroutine and cannot be recovered from).
type flakeFailure struct{}

// runTestWithRetry runs the given test function up to maxRetries+1 times.
// It returns true if any attempt passes, false if all attempts fail.
//
// The run function receives a flakeTestCapture that intercepts failure calls.
// When FailNow/Fatal/Fatalf is called, a flakeFailure panic is raised and
// recovered here, allowing the retry loop to continue without child goroutines.
//
// Note: t.Skip() calls within the run function will propagate to the parent test
// and skip the entire test, bypassing remaining retry attempts. Cleanup functions
// registered via t.Cleanup() also propagate to the parent test and will run when
// the parent test finishes.
func runTestWithRetry(t testing.TB, maxRetries uint, run func(st testing.TB)) bool {
	totalAttempts := int(maxRetries) + 1
	for attempt := range totalAttempts {
		ft := &flakeTestCapture{TB: t}
		failed := runFlakeAttempt(ft, run)
		if !failed {
			if attempt > 0 {
				t.Logf("Flaky test passed on attempt %d/%d", attempt+1, totalAttempts)
			}
			return true
		}
		if attempt < totalAttempts-1 {
			t.Logf("Flaky test attempt %d/%d failed, retrying...", attempt+1, totalAttempts)
		}
	}
	return false
}

// runFlakeAttempt runs the test function with the given flakeTestCapture and
// recovers from flakeFailure panics. Returns true if the attempt failed.
func runFlakeAttempt(ft *flakeTestCapture, run func(st testing.TB)) (failed bool) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(flakeFailure); ok {
				failed = true
				return
			}
			panic(r)
		}
	}()
	run(ft)
	return ft.hasFailed
}

// flakeTestCapture wraps a testing.TB and intercepts failure calls.
// It records whether a failure occurred without propagating it to the parent test.
//
// Methods that abort execution (FailNow, Fatal, Fatalf) panic with flakeFailure
// instead of calling runtime.Goexit(). This allows runTestWithRetry to recover
// and retry the test without using child goroutines.
//
// Non-aborting failure methods (Fail, Error, Errorf) set the failure flag and
// delegate logging to the parent testing.TB, matching standard testing.TB semantics.
type flakeTestCapture struct {
	testing.TB
	hasFailed bool
}

func (f *flakeTestCapture) Fail() {
	f.hasFailed = true
}

func (f *flakeTestCapture) FailNow() {
	f.hasFailed = true
	panic(flakeFailure{})
}

func (f *flakeTestCapture) Error(args ...any) {
	f.hasFailed = true
	f.Log(args...)
}

func (f *flakeTestCapture) Errorf(format string, args ...any) {
	f.hasFailed = true
	f.Logf(format, args...)
}

func (f *flakeTestCapture) Fatal(args ...any) {
	f.hasFailed = true
	f.Log(args...)
	panic(flakeFailure{})
}

func (f *flakeTestCapture) Fatalf(format string, args ...any) {
	f.hasFailed = true
	f.Logf(format, args...)
	panic(flakeFailure{})
}

func (f *flakeTestCapture) Failed() bool {
	return f.hasFailed
}

func (f *flakeTestCapture) Name() string {
	return fmt.Sprintf("%s (flake retry)", f.TB.Name())
}
