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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/tests/state"
)

// fakeAction fails a set number of times before passing, recording how often it ran.
type fakeAction struct {
	stateful

	failures int
	runs     int
	// panicWith is raised on the first run when set, standing in for a real bug
	// in an action rather than a failed assertion.
	panicWith string
}

var _ Action = (*fakeAction)(nil)
var _ Stateful = (*fakeAction)(nil)

func (a *fakeAction) Execute() {
	a.runs++
	if a.panicWith != "" {
		panic(a.panicWith)
	}
	if a.runs <= a.failures {
		require.Fail(a.s.T, "not ready yet")
	}
}

func newEventuallyState(t testing.TB) *state.State {
	return &state.State{T: t}
}

func TestEventually_PassesFirstTime_RunsOnce(t *testing.T) {
	inner := &fakeAction{}
	act := &Eventually{Action: inner}
	act.SetState(newEventuallyState(t))

	act.Execute()

	assert.Equal(t, 1, inner.runs, "a passing action should not be retried")
}

func TestEventually_PassesAfterRetries_KeepsTrying(t *testing.T) {
	inner := &fakeAction{failures: 3}
	act := &Eventually{Action: inner}
	act.SetState(newEventuallyState(t))

	act.Execute()

	assert.Equal(t, 4, inner.runs, "should retry until the action passes")
}

func TestEventually_RestoresTheRealT(t *testing.T) {
	// The real T is swapped out while an attempt runs so its failures can be
	// captured. It has to be put back, or later actions would report into a
	// recorder nobody reads.
	s := newEventuallyState(t)
	act := &Eventually{Action: &fakeAction{failures: 1}}
	act.SetState(s)

	act.Execute()

	assert.Same(t, t, s.T, "the original T must be restored")
}

func TestEventually_NeverPasses_FailsWithTheLastError(t *testing.T) {
	inner := &fakeAction{failures: 1000}
	act := &Eventually{Action: inner, Timeout: 300 * time.Millisecond}

	recorder := &recordingT{TB: t}
	act.SetState(&state.State{T: recorder})

	// The final failure ends the run the same way an assertion would, so it has
	// to be called on its own goroutine.
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() { _ = recover() }()
		act.Execute()
	}()
	<-done

	assert.True(t, recorder.failed, "a timeout must fail the test")
	assert.Contains(t, recorder.message, "not ready yet", "the last failure should be reported")
	assert.Greater(t, inner.runs, 1, "should have retried before giving up")
}

func TestEventually_RespectsTheTimeout(t *testing.T) {
	act := &Eventually{Action: &fakeAction{failures: 1000}, Timeout: 200 * time.Millisecond}
	recorder := &recordingT{TB: t}
	act.SetState(&state.State{T: recorder})

	start := time.Now()
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() { _ = recover() }()
		act.Execute()
	}()
	<-done
	elapsed := time.Since(start)

	assert.True(t, recorder.failed)
	assert.Less(t, elapsed, 5*time.Second, "should give up near the timeout, not run on")
}

func TestEventually_RealPanic_IsNotSwallowed(t *testing.T) {
	// A panic from a bug in the action is not a failed attempt. Retrying it
	// would hide the bug and burn the whole timeout, so it propagates.
	act := &Eventually{Action: &fakeAction{panicWith: "nil map write"}, Timeout: time.Second}
	act.SetState(newEventuallyState(t))

	assert.PanicsWithValue(t, "nil map write", func() { act.Execute() })
}

func TestEventually_RealPanic_RestoresTheRealT(t *testing.T) {
	// The recorder swallows failures so a failed attempt can be retried. If a
	// panic left it in place, every later assertion in the test would be written
	// to something nothing reads, turning real failures into passes.
	st := newEventuallyState(t)
	realT := st.T
	act := &Eventually{Action: &fakeAction{panicWith: "nil map write"}, Timeout: time.Second}
	act.SetState(st)

	assert.Panics(t, func() { act.Execute() })

	assert.Same(t, realT, st.T, "the real T must be restored even when the action panics")
}

func TestEventually_SetsStateOnTheNestedAction(t *testing.T) {
	// The nested action is given the state each attempt, so it can assert
	// through the recorder rather than the real T.
	inner := &fakeAction{failures: 1}
	act := &Eventually{Action: inner}
	act.SetState(newEventuallyState(t))

	act.Execute()

	assert.NotNil(t, inner.s, "the nested action must receive the state")
}

func TestEventually_DefaultTimeoutIsUsedWhenUnset(t *testing.T) {
	act := &Eventually{Action: &fakeAction{}}
	act.SetState(newEventuallyState(t))

	act.Execute()

	assert.Equal(t, time.Duration(0), act.Timeout, "an unset timeout stays unset and falls back to the default")
}
