// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

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
		// T.Skip() respects `defer` statements, and `a.s.AsyncWG.Wait()` in the parent routine, but
		// it does not respect lines after the `T.Skip()` call within this child-routine.
		defer a.s.AsyncWG.Done()

		childReady.Done()

		a.Child.Execute()
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
}
