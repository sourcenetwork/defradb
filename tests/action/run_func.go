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

import "github.com/sourcenetwork/defradb/tests/state"

// RunFunc runs a function at its point in the action sequence. It is an escape hatch for test
// coordination with no dedicated action, such as releasing a build gate or polling for a condition.
// Prefer a dedicated action where one exists.
//
// Set Func for a plain callback, or FuncWithState when the callback needs the test state (e.g. to
// read a node). If both are set, both run, Func first.
type RunFunc struct {
	stateful

	Func func()

	FuncWithState func(*state.State)
}

var _ Action = (*RunFunc)(nil)
var _ Stateful = (*RunFunc)(nil)

func (a *RunFunc) Execute() {
	if a.Func != nil {
		a.Func()
	}
	if a.FuncWithState != nil {
		a.FuncWithState(a.s)
	}
}

// NewRunFunc returns a RunFunc action that runs the given function.
func NewRunFunc(fn func()) *RunFunc {
	return &RunFunc{Func: fn}
}

// NewRunFuncWithState returns a RunFunc action that runs the given function with the test state.
func NewRunFuncWithState(fn func(*state.State)) *RunFunc {
	return &RunFunc{FuncWithState: fn}
}
