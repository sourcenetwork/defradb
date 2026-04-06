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
	"sync"

	"github.com/sourcenetwork/defradb/tests/state"
)

// Async executes the child action in its own child go-routine, the child routine will
// be fully initialized before `Execute()` unblocks, but it is not guaranteed to be executing by
// the runtime.
//
// This action is typically used alongside an `Await` action.
type Async struct {
	s *state.State

	// The child action that should be executed in asynchronously.
	Child Action
}

var _ Action = (*Async)(nil)
var _ Stateful = (*Async)(nil)

func (a *Async) SetState(s *state.State) {
	a.s = s

	if stateful, ok := a.Child.(Stateful); ok {
		stateful.SetState(s)
	}
}

func (a *Async) Execute() {
	a.s.AsyncWG.Add(1)

	// childReady is responsible for ensuring that all child routines have been set up and are
	// now waiting for the start lock to be unlocked.
	childReady := sync.WaitGroup{}
	childReady.Add(1)
	go func() {
		childReady.Done()

		a.Child.Execute()

		a.s.AsyncWG.Done()
	}()

	// Wait for all the children to be ready before returning.
	childReady.Wait()
}

// Await waits for all executing `Async` actions to complete.
type Await struct {
	stateful
}

var _ Action = (*Await)(nil)
var _ Stateful = (*Await)(nil)

func (a *Await) Execute() {
	a.s.AsyncWG.Wait()

	if len(a.s.SkipTest) > 0 {
		// Child actions cannot skip the test from within their own routine, so we must check here
		// to see if they want to.
		a.s.T.Skip(a.s.SkipTest)
	}
}
