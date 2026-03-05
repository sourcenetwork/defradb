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

// Parallel executes all child actions in their own child go-routine, scheduled to start as
// close in time as the runtime will allow.
//
// This action will not complete it's execution until all child actions have completed their
// execution.
type Parallel struct {
	s *state.State

	// The child actions that should be executed in parallel.
	Children []Action
}

var _ Action = (*Parallel)(nil)
var _ Stateful = (*Parallel)(nil)

func (a *Parallel) SetState(s *state.State) {
	a.s = s

	for _, child := range a.Children {
		if stateful, ok := child.(Stateful); ok {
			stateful.SetState(s)
		}
	}
}

func (a *Parallel) Execute() {
	// startLock is responsible for ensuring that all child actions are scheduled/ready before
	// any of them begin execution.
	startLock := sync.RWMutex{}
	startLock.Lock()

	// finishedWG is responsible for ensuring that all child actions complete their execution
	// before this action-function unblocks/completes.
	finishedWG := sync.WaitGroup{}

	// childrenReady is responsible for ensuring that all child routines have been set up and are
	// now waiting for the start lock to be unlocked.
	childrenReady := sync.WaitGroup{}
	for _, childAction := range a.Children {
		childrenReady.Add(1)
		finishedWG.Go(func() {
			childrenReady.Done()

			// Block behind the startWG until all child actions are ready to proceed
			startLock.RLock()
			defer startLock.RUnlock()

			childAction.Execute()
		})
	}

	// Wait for all the children to be ready before allowing them to start.
	childrenReady.Wait()

	// Release the start lock once all child actions are queued so that they may all begin at
	// as close a point in time as the runtime/machine allows.
	startLock.Unlock()

	finishedWG.Wait()
}
