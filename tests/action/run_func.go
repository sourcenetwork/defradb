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

// RunFunc runs a function at its point in the action sequence. It is an escape hatch for test
// coordination with no dedicated action, such as releasing a build gate. Prefer a dedicated action
// where one exists.
type RunFunc struct {
	stateful

	// Func is the function to run.
	Func func()
}

var _ Action = (*RunFunc)(nil)
var _ Stateful = (*RunFunc)(nil)

func (a *RunFunc) Execute() {
	if a.Func != nil {
		a.Func()
	}
}

// NewRunFunc returns a RunFunc action that runs the given function.
func NewRunFunc(fn func()) *RunFunc {
	return &RunFunc{Func: fn}
}
