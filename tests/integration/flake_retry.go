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
	"context"
	"fmt"
	"runtime"
	"testing"
)

// runTestWithRetry runs the given test function up to maxRetries+1 times.
// It returns true if any attempt passes, false if all attempts fail.
//
// Each retry attempt runs in a separate goroutine with a flakeTestCapture
// that intercepts test failures. This is necessary because t.Fatal/t.FailNow
// call runtime.Goexit(), which terminates the current goroutine. By running
// each attempt in its own goroutine, the retry loop in the calling goroutine
// survives failed attempts.
func runTestWithRetry(t testing.TB, maxRetries uint, run func(st testing.TB)) bool {
	totalAttempts := int(maxRetries) + 1
	for attempt := range totalAttempts {
		ft := &flakeTestCapture{TB: t}
		done := make(chan struct{})
		go func() {
			defer close(done)
			run(ft)
		}()
		<-done
		if !ft.hasFailed {
			if attempt > 0 {
				log.InfoContext(context.Background(),
					fmt.Sprintf("Flaky test passed on attempt %d/%d", attempt+1, totalAttempts))
			}
			return true
		}
		if attempt < totalAttempts-1 {
			log.InfoContext(context.Background(),
				fmt.Sprintf("Flaky test attempt %d/%d failed, retrying...", attempt+1, totalAttempts))
		}
	}
	return false
}

// flakeTestCapture wraps a testing.TB and intercepts failure calls.
// It records whether a failure occurred without propagating it to the parent test.
// Fatal/FailNow call runtime.Goexit() to terminate the current goroutine
// (matching the contract of testing.TB), but since each attempt runs in its
// own goroutine, this doesn't affect the retry loop.
type flakeTestCapture struct {
	testing.TB
	hasFailed bool
}

func (f *flakeTestCapture) Fail() {
	f.hasFailed = true
}

func (f *flakeTestCapture) FailNow() {
	f.hasFailed = true
	runtime.Goexit()
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
	runtime.Goexit()
}

func (f *flakeTestCapture) Fatalf(format string, args ...any) {
	f.hasFailed = true
	f.Logf(format, args...)
	runtime.Goexit()
}

func (f *flakeTestCapture) Failed() bool {
	return f.hasFailed
}

func (f *flakeTestCapture) Name() string {
	return f.TB.Name()
}
