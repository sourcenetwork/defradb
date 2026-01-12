// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package wizard

import (
	tea "github.com/charmbracelet/bubbletea"
)

// The purpose of a brancher model, is that we can have a step whose only purpose.
// is to branch to a different step based on an evaluator function.
type modelBrancher struct {
	// nextSteps can be assigned dynamically. It should be set to a slice of steps,
	// in order of the choices. For example, if the choices are ["Yes", "No"], then
	// the nextSteps value should be set to []step{step2, step2}, corresponding to
	// how the next choice should be branched to. nil is valid, and will be treated
	// as the end of the chain.
	nextSteps []step

	// The evaluator function should return an integer value, which will be used to
	// index into the nextSteps slice and decide which step to branch to.
	evaluator func(ctx *WizardContext) (int, error)

	// callback is a function that will be called when this step is done.
	callback func(s step, ctx *WizardContext) error
}

// initialModelBlank should be called instead of manually constructing the struct
func initialModelBrancher() *modelBrancher {
	return &modelBrancher{
		nextSteps: nil,
		evaluator: nil,
		callback:  nil,
	}
}

// Init() should not be called except by the main model
func (m *modelBrancher) Init() tea.Cmd {
	return nil
}

// Update() should not be called except by the main model
func (m *modelBrancher) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

// View() should not be called except by the main model
func (m *modelBrancher) View() string {
	return ""
}

// Next() will invoke the evaluator function to find the next step. If the
// evaluator is not set, or returns an invalid index, then the next step will be nil.
func (m *modelBrancher) Next(ctx *WizardContext) (step, error) {
	if m.evaluator == nil {
		return nil, nil
	}
	evaluatorResult, err := m.evaluator(ctx)
	if err != nil {
		return nil, err
	}
	if evaluatorResult < 0 || evaluatorResult >= len(m.nextSteps) {
		return nil, nil
	}
	return m.nextSteps[evaluatorResult], nil
}

// There is no result for this step
func (m *modelBrancher) Result() any {
	return nil
}

// Because this step does not do anything, it is always done
func (m *modelBrancher) Done() bool {
	return true
}

// Returns a dummy string
func (m *modelBrancher) ID() string {
	return BlankStepID
}

// Callback() should not be called except by the main model
func (m *modelBrancher) Callback(ctx *WizardContext) error {
	if m.callback == nil {
		return nil
	}
	return m.callback(m, ctx)
}
