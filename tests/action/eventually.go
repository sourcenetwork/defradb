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
	"fmt"
	"testing"
	"time"
)

const (
	// eventuallyTimeout is how long [Eventually] retries before giving up.
	eventuallyTimeout = 20 * time.Second
	// eventuallyInterval is how long [Eventually] waits between attempts.
	eventuallyInterval = 100 * time.Millisecond
)

// Eventually runs another action until it stops failing.
//
// Use it when the harness cannot tell that something has finished and the test
// has to keep asking. The clearest case is a node running an older release: it
// can hold a document written against a schema it has never seen, but it cannot
// report the commit, so there is no signal to wait for.
//
// The wrapped action's assertions are captured rather than failing the test, so
// a failed attempt is just a retry. The final attempt's failure is reported as
// this action's failure.
type Eventually struct {
	stateful

	// Action is retried until it passes or the timeout is reached.
	Action Action

	// Timeout overrides the default retry window.
	Timeout time.Duration
}

var _ Action = (*Eventually)(nil)
var _ Stateful = (*Eventually)(nil)

func (a *Eventually) Execute() {
	timeout := a.Timeout
	if timeout == 0 {
		timeout = eventuallyTimeout
	}

	realT := a.s.T
	// Restore with a defer: a real panic from the nested action skips past the
	// assignment below, which would leave the state pointing at a recorder that
	// nothing reads, silently swallowing later failures.
	defer func() { a.s.T = realT }()
	deadline := time.Now().Add(timeout)

	var lastErr string
	for {
		recorder := &recordingT{TB: realT}
		a.s.T = recorder
		if stateful, ok := a.Action.(Stateful); ok {
			stateful.SetState(a.s)
		}
		failed := a.attempt(recorder)
		a.s.T = realT

		if !failed {
			return
		}
		lastErr = recorder.message

		if time.Now().After(deadline) {
			a.s.T.Errorf("action did not pass within %s: %s", timeout, lastErr)
			a.s.T.FailNow()
			return
		}
		time.Sleep(eventuallyInterval)
	}
}

// attempt runs the action once, reporting whether it failed rather than failing
// the test.
func (a *Eventually) attempt(recorder *recordingT) (failed bool) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		if _, ok := r.(attemptFailure); ok {
			failed = true
			return
		}
		// Anything else is a real panic and belongs to the caller.
		panic(r)
	}()

	a.Action.Execute()
	return recorder.failed
}

// attemptFailure marks an attempt that ended early because an assertion failed.
//
// The failure is raised as a panic rather than by ending the goroutine, so the
// retry loop can recover from it in place. This mirrors how the flake retry
// helper in the integration package handles the same problem.
type attemptFailure struct{}

// recordingT captures assertion failures instead of failing the test.
type recordingT struct {
	testing.TB

	failed  bool
	message string
}

func (t *recordingT) Errorf(format string, args ...any) {
	t.fail(fmt.Sprintf(format, args...))
}

func (t *recordingT) Error(args ...any) {
	t.fail(fmt.Sprint(args...))
}

func (t *recordingT) Fatal(args ...any) {
	t.fail(fmt.Sprint(args...))
	t.FailNow()
}

func (t *recordingT) Fatalf(format string, args ...any) {
	t.fail(fmt.Sprintf(format, args...))
	t.FailNow()
}

func (t *recordingT) Fail() {
	t.fail("")
}

func (t *recordingT) FailNow() {
	t.failed = true
	panic(attemptFailure{})
}

func (t *recordingT) Failed() bool {
	return t.failed
}

func (t *recordingT) fail(message string) {
	t.failed = true
	if message != "" {
		t.message = message
	}
}

// NewEventually returns an [Eventually] wrapping the given action.
func NewEventually(action Action) *Eventually {
	return &Eventually{Action: action}
}
